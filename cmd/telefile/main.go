package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
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
		return nil
	})

	// Handle callbacks
	b.Router.OnCallback("start", handlers.Start)
	b.Router.OnCallback("help", handlers.Help)
	b.Router.OnCallback("myfiles", handlers.MyFiles)
	b.Router.OnCallback("plans", handlers.ShowPlans)
	b.Router.OnCallback("my_stats", handlers.ShowStats)
	b.Router.OnCallback("referral", func(ctx *bot.Context) error {
		return handlers.ShowReferral(ctx, cfg)
	})
	b.Router.OnCallback("redeem", handlers.StartRedeem)
	b.Router.OnCallback("report", handlers.StartReport)
	b.Router.OnCallback("upload_batch", handlers.StartBatch)
	b.Router.OnCallback("upload_batch_start", handlers.StartBatch)
	b.Router.OnCallback("upload_batch_done", func(ctx *bot.Context) error {
		return handlers.UploadPasswordNo(ctx) // Skip to expiry/finish
	})
	b.Router.OnCallback("upload_help", func(ctx *bot.Context) error {
		_, err := ctx.Reply("📤 Send any file to upload it!")
		return err
	})
	b.Router.OnCallback("upload_cancel", handlers.CancelUpload)

	// Upload password callbacks
	b.Router.OnCallback("upload_pwd_yes", handlers.UploadPasswordYes)
	b.Router.OnCallback("upload_pwd_no", handlers.UploadPasswordNo)

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
		if ctx.Message.Document != nil || ctx.Message.Photo != nil {
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

		// Handle admin user lookup input
		if _, ok := adminhandlers.UserStates[ctx.SenderID()]; ok {
			return adminhandlers.AdminUserLookup(ctx)
		}

		// Handle broadcast input
		if _, ok := adminhandlers.BroadcastStates[ctx.SenderID()]; ok {
			return adminhandlers.HandleBroadcast(ctx)
		}

		return nil
	})

	// Admin middleware and commands
	adminMiddleware := middleware.AdminOnly(cfg)
	adminBot := bot.New(cfg.AdminBotToken)
	instance.AdminBot = adminBot

	adminBot.Router.Use(adminMiddleware)

	// Admin commands
	adminBot.Router.On("/admin", adminhandlers.AdminPanel)
	adminBot.Router.On("/stats", adminhandlers.AdminStats)

	// Admin callbacks
	adminBot.Router.OnCallback("admin_panel", adminhandlers.AdminPanel)
	adminBot.Router.OnCallback("admin_stats", adminhandlers.AdminStats)
	adminBot.Router.OnCallback("admin_users", adminhandlers.AdminUsers)
	adminBot.Router.OnCallback("admin_gifts", adminhandlers.ShowGifts)
	adminBot.Router.OnCallback("admin_broadcast", adminhandlers.StartBroadcast)
	adminBot.Router.OnCallback("admin_cleanup", adminhandlers.AdminCleanup)

	// Gift creation callbacks
	adminBot.Router.OnCallback("admin_gift_create", adminhandlers.StartGiftCreate)
	adminBot.Router.OnCallback("admin_gift_10gb", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftStorage(ctx, 10)
	})
	adminBot.Router.OnCallback("admin_gift_50gb", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftStorage(ctx, 50)
	})
	adminBot.Router.OnCallback("admin_gift_100gb", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftStorage(ctx, 100)
	})
	adminBot.Router.OnCallback("admin_gift_500gb", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftStorage(ctx, 500)
	})
	adminBot.Router.OnCallback("admin_gift_claims_1", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftClaims(ctx, 1)
	})
	adminBot.Router.OnCallback("admin_gift_claims_10", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftClaims(ctx, 10)
	})
	adminBot.Router.OnCallback("admin_gift_claims_100", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftClaims(ctx, 100)
	})
	adminBot.Router.OnCallback("admin_gift_claims_unlimited", func(ctx *bot.Context) error {
		return adminhandlers.HandleGiftClaims(ctx, -1)
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
			if err := b.StartPolling(5*time.Second, 20*time.Second); err != nil {
				log.Printf("Bot polling error: %v", err)
			}
		}()
		go func() {
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
