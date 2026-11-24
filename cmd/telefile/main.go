package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	adminhandlers "github.com/Yashwanth-Kumar-26/telefile/internal/admin/handlers"
	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/handlers"
	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/telefile/internal/config"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/middleware"
	"github.com/Yashwanth-Kumar-26/telefile/internal/scheduler"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

func main() {
	log.Println("Starting TeleFile bot...")

	// Load config
	cfg := config.Load()

	// Connect to database
	db.MustConnect(cfg)
	defer db.Close()

	log.Println("Connected to database")

	// Initialize TeleRiHa bot
	b := bot.New(cfg.BotToken)
	instance.MainBot = b

	// Middleware chain
	b.Router.Use(middleware.BanCheck())
	b.Router.Use(middleware.EnsureUser(cfg.BaseStorageGB))

	// User commands
	b.Router.On("/start", handlers.Start)
	b.Router.On("/help", handlers.Help)
	b.Router.On("/myfiles", handlers.MyFiles)
	b.Router.On("/plans", handlers.ShowPlans)
	b.Router.On("/referral", func(ctx *bot.Context) error {
		return handlers.ShowReferral(ctx, cfg)
	})
	b.Router.On("/redeem", handlers.StartRedeem)
	b.Router.On("/report", handlers.StartReport)
	b.Router.On("/cancel", func(ctx *bot.Context) error {
		_ = handlers.CancelUpload(ctx)
		_ = handlers.CancelRedeem(ctx)
		_ = handlers.CancelDownload(ctx)
		_ = handlers.CancelReport(ctx)
		delete(handlers.BuyPlanStates, ctx.SenderID())
		adminhandlers.AdminFSM.End(ctx.SenderID())
		return nil
	})

	// Cancel confirm handlers
	b.Router.OnCallback("cancel_confirm", func(ctx *bot.Context) error {
		sid := ctx.SenderID()
		delete(handlers.PendingUploads, sid)
		delete(handlers.DownloadStates, sid)
		delete(handlers.RedeemStates, sid)
		delete(handlers.ReportStates, sid)
		delete(handlers.BuyPlanStates, sid)
		return handlers.Start(ctx)
	})
	b.Router.OnCallback("cancel_deny", func(ctx *bot.Context) error {
		return handlers.Start(ctx)
	})

	// Handle callbacks
	b.Router.OnCallback("start", handlers.Start)
	b.Router.OnCallback("help", handlers.Help)
	b.Router.OnCallback("myfiles_menu", handlers.MyFiles)
	b.Router.OnCallback("plans", handlers.ShowPlans)
	b.Router.OnCallback("buy_plan", handlers.ShowBuyPlanMenu)
	b.Router.OnCallback("bp_select_", func(ctx *bot.Context) error {
		id, err := strconv.Atoi(parseDynamicSuffix(ctx, "bp_select_"))
		if err != nil {
			return nil
		}
		return handlers.BuyPlanSelect(ctx, id)
	})
	b.Router.OnCallback("bp_upi", handlers.ShowUPIQR)
	b.Router.OnCallback("bp_cancel", handlers.BuyPlanHandleCancel)
	b.Router.OnCallback("my_stats", handlers.ShowStats)
	b.Router.OnCallback("referral", func(ctx *bot.Context) error {
		return handlers.ShowReferral(ctx, cfg)
	})
	b.Router.OnCallback("copy_referral", func(ctx *bot.Context) error {
		_ = ctx.AnswerCallback("", false)
		link := fmt.Sprintf("https://t.me/%s?start=ref_%d", cfg.BotUsername, ctx.SenderID())
		_, _ = ctx.Reply("Your referral link:\n" + link)
		return nil
	})
	b.Router.OnCallback("redeem", handlers.StartRedeem)
	b.Router.OnCallback("report", handlers.StartReport)
	b.Router.OnCallback("upload_batch_activate", handlers.StartBatch)
	b.Router.OnCallback("upload_batch_done", func(ctx *bot.Context) error {
		// Show password prompt before expiry selection
		senderID := ctx.SenderID()
		state, ok := handlers.PendingUploads[senderID]
		if ok && state != nil && state.Password == "" {
			markup := bot.NewInlineKeyboard(
				[]bot.InlineKeyboardButton{
					bot.InlineButton("Yes", "upload_pwd_yes"),
					bot.InlineButton("No", "upload_pwd_no"),
				},
			)
			_, _ = ctx.EditMessageText("Do you want to set a password?", bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
			return nil
		}
		return handlers.UploadPasswordNo(ctx)
	})
	b.Router.OnCallback("upload_help", func(ctx *bot.Context) error {
		_, err := ctx.Reply("Send any file to upload it! 📤")
		return err
	})
	b.Router.OnCallback("upload_cancel", handlers.CancelUpload)

	// Upload password callbacks
	b.Router.OnCallback("upload_pwd_yes", handlers.UploadPasswordYes)
	b.Router.OnCallback("upload_pwd_no", handlers.UploadPasswordNo)
	b.Router.OnCallback("myfiles_", func(ctx *bot.Context) error {
		pageStr := parseDynamicSuffix(ctx, "myfiles_")
		return handlers.HandleMyFilesPage(ctx, pageStr)
	})

	// Upload expiry callbacks
	b.Router.OnCallback("upload_exp_1h", func(ctx *bot.Context) error {
		return handlers.UploadExpiryHandler(ctx, cfg, 1)
	})
	b.Router.OnCallback("upload_exp_24h", func(ctx *bot.Context) error {
		return handlers.UploadExpiryHandler(ctx, cfg, 24)
	})
	b.Router.OnCallback("upload_exp_7d", func(ctx *bot.Context) error {
		return handlers.UploadExpiryHandler(ctx, cfg, 24*7)
	})
	b.Router.OnCallback("upload_exp_30d", func(ctx *bot.Context) error {
		return handlers.UploadExpiryHandler(ctx, cfg, 24*30)
	})
	b.Router.OnCallback("upload_exp_never", func(ctx *bot.Context) error {
		return handlers.UploadExpiryHandler(ctx, cfg, 0)
	})

	// Message handler for uploads
	b.Router.Default(func(ctx *bot.Context) error {
		if ctx.Message == nil {
			return nil
		}

		// Handle file upload
		if ctx.Message.Document != nil || ctx.Message.Photo != nil || ctx.Message.Video != nil || ctx.Message.Audio != nil || ctx.Message.Voice != nil || ctx.Message.Animation != nil {
			sid := ctx.SenderID()
			if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateBuyPlanUpload) {
				return handlers.BuyPlanHandlePhoto(ctx)
			}
			return handlers.HandleUpload(ctx, cfg)
		}

		text := ctx.Text()

		// Handle password input during upload
		state, ok := handlers.PendingUploads[ctx.SenderID()]
		if ok && state != nil && text != "" {
			return handlers.HandlePasswordInput(ctx)
		}

		// Handle password input during download
		if _, ok := handlers.DownloadStates[ctx.SenderID()]; ok {
			return handlers.HandleDownloadPassword(ctx)
		}

		// Handle redeem code input
		if _, ok := handlers.RedeemStates[ctx.SenderID()]; ok {
			return handlers.HandleRedeemCode(ctx)
		}

		// Handle report input
		if _, ok := handlers.ReportStates[ctx.SenderID()]; ok {
			return handlers.HandleReport(ctx)
		}

		// Handle admin FSM text inputs
		sid := ctx.SenderID()
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateLookupUser) {
			return adminhandlers.AdmLookupDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateBanUser) {
			return adminhandlers.AdmBanDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateUnbanUser) {
			return adminhandlers.AdmUnbanDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateSetStorage) {
			return adminhandlers.AdmSetStorageDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateAddBonus) {
			return adminhandlers.AdmAddBonusDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateBroadcast) {
			return adminhandlers.AdmBroadcastSend(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanName) {
			return adminhandlers.AdmCreatePlanName(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanPrice) {
			return adminhandlers.AdmCreatePlanPrice(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanStorage) {
			return adminhandlers.AdmCreatePlanStorage(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanDays) {
			return adminhandlers.AdmCreatePlanDays(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanDesc) {
			return adminhandlers.AdmCreatePlanDesc(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreateGift) {
			return adminhandlers.AdmCreateGiftDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreateGiftExpiry) {
			return adminhandlers.AdmCreateGiftExpiry(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreateGiftClaims) {
			return adminhandlers.AdmCreateGiftClaims(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateBuyPlanUpload) {
			return handlers.HandleUpload(ctx, cfg)
		}

		return nil
	})

	// Admin middleware and commands
	adminMiddleware := middleware.AdminOnly(cfg)
	adminBot := bot.New(cfg.AdminBotToken)
	instance.AdminBot = adminBot

	adminBot.Router.Use(adminMiddleware)

	// Admin commands
	adminBot.Router.On("/admin", func(ctx *bot.Context) error {
		log.Printf("[ADMIN] /admin received from %d", ctx.SenderID())
		_, err := ctx.Reply("Admin access granted")
		if err != nil {
			log.Printf("[ADMIN] Reply failed: %v", err)
			// Try using instance.MainBot instead
			_, err2 := instance.MainBot.SendMessage(ctx.ChatID(), "Admin access (via main bot)")
			if err2 != nil {
				log.Printf("[ADMIN] MainBot reply also failed: %v", err2)
			}
		}
		return nil
	})
	adminBot.Router.On("/stats", adminhandlers.AdmStats)
	adminBot.Router.On("/cancel", adminhandlers.AdminCancel)

	// Admin panel main callback
	adminBot.Router.OnCallback("admin_panel", adminhandlers.AdminPanel)

	// Stats
	adminBot.Router.OnCallback("adm_stats", adminhandlers.AdmStats)

	// User management
	adminBot.Router.OnCallback("adm_lookup", adminhandlers.AdmLookupPrompt)
	adminBot.Router.OnCallback("adm_ban_prompt", adminhandlers.AdmBanPrompt)
	adminBot.Router.OnCallback("adm_unban_prompt", adminhandlers.AdmUnbanPrompt)
	adminBot.Router.OnCallback("adm_set_storage", adminhandlers.AdmSetStoragePrompt)
	adminBot.Router.OnCallback("adm_add_bonus", adminhandlers.AdmAddBonusGeneral)
	adminBot.Router.OnCallback("adm_setstg_", func(ctx *bot.Context) error {
		uid, err := strconv.ParseInt(parseDynamicSuffix(ctx, "adm_setstg_"), 10, 64)
		if err != nil {
			return nil
		}
		return adminhandlers.AdmSetStorageForUser(ctx, uid)
	})
	adminBot.Router.OnCallback("adm_addbonus_", func(ctx *bot.Context) error {
		uid, err := strconv.ParseInt(parseDynamicSuffix(ctx, "adm_addbonus_"), 10, 64)
		if err != nil {
			return nil
		}
		return adminhandlers.AdmAddBonusPrompt(ctx, uid)
	})
	adminBot.Router.OnCallback("adm_ban_do_", func(ctx *bot.Context) error {
		uid, err := strconv.ParseInt(parseDynamicSuffix(ctx, "adm_ban_do_"), 10, 64)
		if err != nil {
			return nil
		}
		return adminhandlers.AdmBanDoByID(ctx, uid)
	})
	adminBot.Router.OnCallback("adm_unban_do_", func(ctx *bot.Context) error {
		uid, err := strconv.ParseInt(parseDynamicSuffix(ctx, "adm_unban_do_"), 10, 64)
		if err != nil {
			return nil
		}
		return adminhandlers.AdmUnbanDoByID(ctx, uid)
	})

	// Plans
	adminBot.Router.OnCallback("adm_plans", adminhandlers.AdmPlans)
	adminBot.Router.OnCallback("adm_create_plan", adminhandlers.AdmCreatePlanStart)
	adminBot.Router.OnCallback("adm_delplan_", func(ctx *bot.Context) error {
		id, err := strconv.Atoi(parseDynamicSuffix(ctx, "adm_delplan_"))
		if err != nil {
			return nil
		}
		return adminhandlers.AdmDelPlan(ctx, id)
	})

	// Pending Payments
	adminBot.Router.OnCallback("adm_payments", adminhandlers.AdmPayments)
	adminBot.Router.OnCallback("adm_pay_detail_", func(ctx *bot.Context) error {
		id, err := strconv.Atoi(parseDynamicSuffix(ctx, "adm_pay_detail_"))
		if err != nil {
			return nil
		}
		return adminhandlers.AdmPayDetail(ctx, id)
	})
	adminBot.Router.OnCallback("pay_approve_", func(ctx *bot.Context) error {
		id, err := strconv.Atoi(parseDynamicSuffix(ctx, "pay_approve_"))
		if err != nil {
			return nil
		}
		return adminhandlers.AdmPayApprove(ctx, id)
	})
	adminBot.Router.OnCallback("pay_reject_", func(ctx *bot.Context) error {
		id, err := strconv.Atoi(parseDynamicSuffix(ctx, "pay_reject_"))
		if err != nil {
			return nil
		}
		return adminhandlers.AdmPayReject(ctx, id)
	})

	// Gifts
	adminBot.Router.OnCallback("adm_gifts", adminhandlers.AdmGifts)
	adminBot.Router.OnCallback("adm_create_gift", adminhandlers.AdmCreateGiftPrompt)

	// Broadcast
	adminBot.Router.OnCallback("adm_broadcast", adminhandlers.AdmBroadcastPrompt)

	// Cleanup
	adminBot.Router.OnCallback("adm_cleanup", adminhandlers.AdmCleanup)

	// Default handler — routes FSM text inputs for admin flows
	adminBot.Router.Default(func(ctx *bot.Context) error {
		sid := ctx.SenderID()
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateLookupUser) {
			return adminhandlers.AdmLookupDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateBanUser) {
			return adminhandlers.AdmBanDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateUnbanUser) {
			return adminhandlers.AdmUnbanDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateSetStorage) {
			return adminhandlers.AdmSetStorageDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateAddBonus) {
			return adminhandlers.AdmAddBonusDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateBroadcast) {
			return adminhandlers.AdmBroadcastSend(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanName) {
			return adminhandlers.AdmCreatePlanName(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanPrice) {
			return adminhandlers.AdmCreatePlanPrice(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanStorage) {
			return adminhandlers.AdmCreatePlanStorage(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanDays) {
			return adminhandlers.AdmCreatePlanDays(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreatePlanDesc) {
			return adminhandlers.AdmCreatePlanDesc(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreateGift) {
			return adminhandlers.AdmCreateGiftDo(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreateGiftExpiry) {
			return adminhandlers.AdmCreateGiftExpiry(ctx)
		}
		if adminhandlers.AdminFSM.HasState(sid, adminhandlers.StateCreateGiftClaims) {
			return adminhandlers.AdmCreateGiftClaims(ctx)
		}
		return nil
	})

	// Start scheduler
	scheduler.StartScheduler()
	defer scheduler.StopScheduler()

	// Start webhook server
	if cfg.WebhookURL != "" {
		go func() {
			http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
				var update bot.Update
				if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
					log.Printf("Failed to decode update: %v", err)
					return
				}
				processBotUpdate(b, update)
			})
			http.HandleFunc("/admin/webhook", func(w http.ResponseWriter, r *http.Request) {
				var update bot.Update
				if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
					log.Printf("Failed to decode admin update: %v", err)
					return
				}
				processBotUpdate(adminBot, update)
			})

			log.Printf("Starting webhook server on :%s", cfg.Port)
			if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	} else {
		// Polling mode
		log.Println("Starting bots in polling mode...")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in bot polling: %v", r)
				}
			}()
			if err := b.StartPolling(5*time.Second, 20*time.Second); err != nil {
				log.Printf("Bot polling error: %v", err)
			}
		}()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic in admin bot polling: %v", r)
				}
			}()
			if err := adminBot.StartPolling(5*time.Second, 20*time.Second); err != nil {
				log.Printf("Admin bot polling error: %v", err)
			}
		}()
	}

	// Wait for shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	b.Stop()
	adminBot.Stop()
}

// parseDynamicSuffix extracts the suffix after a prefix from the callback data.
// e.g. parseDynamicSuffix(ctx, "admin_ban_") with callback "admin_ban_12345" returns "12345".
func parseDynamicSuffix(ctx *bot.Context, prefix string) string {
	data := ""
	if ctx.CallbackQuery != nil && ctx.CallbackQuery.Data != "" {
		data = ctx.CallbackQuery.Data
	} else if ctx.Update.CallbackQuery != nil {
		data = ctx.Update.CallbackQuery.Data
	}
	if len(data) > len(prefix) && data[:len(prefix)] == prefix {
		return data[len(prefix):]
	}
	return ""
}

// parseDynamicParts splits the suffix of a callback by a delimiter.
// e.g. parseDynamicParts(ctx, "admin_usr_files_", "_") with "admin_usr_files_123_2"
// returns ["123", "2"].
func parseDynamicParts(ctx *bot.Context, prefix string, sep string) []string {
	suffix := parseDynamicSuffix(ctx, prefix)
	if suffix == "" {
		return nil
	}
	parts := strings.Split(suffix, sep)
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// processBotUpdate manually routes the update to the appropriate handler.
func processBotUpdate(b *bot.Bot, update bot.Update) {
	ctx := bot.NewContext(b, update)

	switch {
	case update.Message != nil:
		b.Router.HandleMessage(ctx, *update.Message)
	case update.CallbackQuery != nil:
		b.Router.HandleCallbackQuery(ctx, *update.CallbackQuery)
	case update.InlineQuery != nil:
		b.Router.HandleInlineQuery(ctx, *update.InlineQuery)
	case update.ChosenInlineResult != nil:
		b.Router.HandleChosenInlineResult(ctx, *update.ChosenInlineResult)
	default:
		// Unknown update type
	}
}

// parseStartParam extracts the parameter from /start command text
func parseStartParam(text string) string {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
