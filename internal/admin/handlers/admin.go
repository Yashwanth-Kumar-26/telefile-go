package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// FSM state constants
const (
	StateNone                = ""
	StateBroadcast           = "broadcast"
	StateLookupUser          = "lookup_user"
	StateSetStorage          = "set_storage"
	StateAddBonus            = "add_bonus"
	StateBanUser             = "ban_user"
	StateUnbanUser           = "unban_user"
	StateCreateGift          = "create_gift"
	StateCreateGiftExpiry    = "create_gift_expiry"
	StateCreateGiftClaims    = "create_gift_claims"
	StateCreatePlanName      = "create_plan_name"
	StateCreatePlanPrice     = "create_plan_price"
	StateCreatePlanStorage   = "create_plan_storage"
	StateCreatePlanDays      = "create_plan_days"
	StateCreatePlanDesc      = "create_plan_desc"
	StateBuyPlanUpload       = "buy_plan_upload"
)

// AdminFSM is the package-level FSM instance shared by all admin handlers.
var AdminFSM *AdminFSMType

func init() {
	AdminFSM = NewAdminFSMInline()
}

// AdminFSMType is a minimal FSM implementation.
type AdminFSMType struct {
	convs map[int64]*adminConversation
}

type adminConversation struct {
	State string
	Data  map[string]interface{}
}

func NewAdminFSMInline() *AdminFSMType {
	return &AdminFSMType{convs: make(map[int64]*adminConversation)}
}

func (fsm *AdminFSMType) Get(senderID int64) *adminConversation {
	if c, ok := fsm.convs[senderID]; ok {
		return c
	}
	c := &adminConversation{State: StateNone, Data: make(map[string]interface{})}
	fsm.convs[senderID] = c
	return c
}

func (fsm *AdminFSMType) SetState(senderID int64, state string) {
	c := fsm.Get(senderID)
	if c.State == StateNone || c.State == "" {
		c.Data = make(map[string]interface{})
	}
	c.State = state
}

func (fsm *AdminFSMType) SetData(senderID int64, key string, value interface{}) {
	fsm.Get(senderID).Data[key] = value
}

func (fsm *AdminFSMType) GetData(senderID int64, key string) interface{} {
	return fsm.Get(senderID).Data[key]
}

func (fsm *AdminFSMType) GetIntData(senderID int64, key string) int64 {
	v := fsm.GetData(senderID, key)
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}

func (fsm *AdminFSMType) End(senderID int64) {
	fsm.Get(senderID).State = StateNone
}

func (fsm *AdminFSMType) HasState(senderID int64, state string) bool {
	return fsm.Get(senderID).State == state
}

// adminKB returns the main admin panel keyboard.
func adminKB() *bot.InlineKeyboardMarkup {
	return bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Statistics 📊", "adm_stats"),
			bot.InlineButton("User Lookup 👥", "adm_lookup"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Ban User 🚫", "adm_ban_prompt"),
			bot.InlineButton("Unban User ✅", "adm_unban_prompt"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Set Storage 💾", "adm_set_storage"),
			bot.InlineButton("Add Bonus 🎁", "adm_add_bonus"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Plans 💎", "adm_plans"),
			bot.InlineButton("Pending Payments 💳", "adm_payments"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Gift Codes 🎁", "adm_gifts"),
			bot.InlineButton("Broadcast 📢", "adm_broadcast"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Run Cleanup 🧹", "adm_cleanup"),
			bot.InlineButton("Main Menu", "admin_panel"),
		},
	)
}

func AdminPanel(ctx *bot.Context) error {
	log.Printf("[ADMIN] AdminPanel called by user %d", ctx.SenderID())
	userCount, _ := db.GetUserCount(ctx.Context())
	bannedCount, _ := db.GetBannedCount(ctx.Context())
	pendingPays, _ := db.GetPendingPaymentCount(ctx.Context())

	text := fmt.Sprintf("*Admin Panel* ⚙️\n\n*Total Users:* %d\n*Banned:* %d\n*Pending Payments:* %d\n\nSelect an option 👇", userCount, bannedCount, pendingPays)
	markup := adminKB()

	if ctx.CallbackQuery != nil {
		_, _ = ctx.EditMessageText(text, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}

func AdminCancel(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	AdminFSM.End(senderID)
	return AdminPanel(ctx)
}

func AdmStats(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)

	userCount, _ := db.GetUserCount(ctx.Context())
	newToday, _ := db.GetNewUsersCount(ctx.Context(), 1)
	banned, _ := db.GetBannedCount(ctx.Context())
	fileCount, _ := db.GetFileCount(ctx.Context())
	dailyUploads, _ := db.GetDailyUploads(ctx.Context())
	totalStorage, _ := db.GetTotalStorageUsed(ctx.Context())
	activeSubs, _ := db.GetActiveSubscriptionCount(ctx.Context())
	pendingPays, _ := db.GetPendingPaymentCount(ctx.Context())

	text := fmt.Sprintf("*Global Statistics* 📊\n\n*Users:* %d *(+%d today)* | *Banned:* %d 🚫\n*Files:* %d *(+%d today)* 📁\n*Total Storage Used:* %s 💾\n*Active Subscriptions:* %d 💎\n*Pending Payments:* %d 💳",
		userCount, newToday, banned, fileCount, dailyUploads, utils.FormatBytes(totalStorage), activeSubs, pendingPays)

	markup := backKB("admin_panel")
	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func AdmLookupPrompt(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateLookupUser)
	_, _ = ctx.Reply("*User Lookup*\n\nSend a user ID or @username:", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmLookupDo(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateLookupUser) {
		return nil
	}
	AdminFSM.End(senderID)

	query := strings.TrimSpace(ctx.Text())

	if strings.HasPrefix(query, "@") {
		username := strings.TrimPrefix(query, "@")
		user, err := db.GetUserByUsername(ctx.Context(), username)
		if err != nil {
			_, _ = ctx.Reply("User not found. ❌")
			return nil
		}
		return showUserDetail(ctx, user)
	}

	uid, err := strconv.ParseInt(query, 10, 64)
	if err != nil {
		_, _ = ctx.Reply("Invalid format. Send a user ID (numbers) or @username.")
		return nil
	}
	user, err := db.GetUser(ctx.Context(), uid)
	if err != nil {
		_, _ = ctx.Reply("User not found.")
		return nil
	}
	return showUserDetail(ctx, user)
}

func AdmBanPrompt(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateBanUser)
	_, _ = ctx.Reply("*Ban User*\n\nSend the user ID to ban:", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmBanDo(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateBanUser) {
		return nil
	}
	AdminFSM.End(senderID)

	uid, ok := parseID(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Invalid user ID.")
		return nil
	}
	return doBan(ctx, uid)
}

func AdmUnbanPrompt(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateUnbanUser)
	_, _ = ctx.Reply("*Unban User*\n\nSend the user ID to unban:", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmUnbanDo(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateUnbanUser) {
		return nil
	}
	AdminFSM.End(senderID)

	uid, ok := parseID(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Invalid user ID.")
		return nil
	}
	return doUnban(ctx, uid)
}

func AdmBanDoByID(ctx *bot.Context, uid int64) error {
	_ = ctx.AnswerCallback("", false)
	return doBan(ctx, uid)
}

func AdmUnbanDoByID(ctx *bot.Context, uid int64) error {
	_ = ctx.AnswerCallback("", false)
	return doUnban(ctx, uid)
}

func doBan(ctx *bot.Context, uid int64) error {
	_ = db.BanUser(ctx.Context(), uid)
	_, _ = ctx.Reply(fmt.Sprintf("User `%d` has been banned.", uid), bot.WithParseMode("Markdown"))
	_, _ = instance.MainBot.SendMessage(uid, "*You have been banned from TeleFile.*", bot.WithParseMode("Markdown"))
	return nil
}

func doUnban(ctx *bot.Context, uid int64) error {
	_ = db.UnbanUser(ctx.Context(), uid)
	_, _ = ctx.Reply(fmt.Sprintf("User `%d` has been unbanned.", uid), bot.WithParseMode("Markdown"))
	_, _ = instance.MainBot.SendMessage(uid, "*You have been unbanned from TeleFile!*", bot.WithParseMode("Markdown"))
	return nil
}

func AdmSetStoragePrompt(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateSetStorage)
	AdminFSM.SetData(ctx.SenderID(), "target_id", int64(0))
	_, _ = ctx.Reply("*Set Storage*\n\nSend: `USER_ID STORAGE_GB`\nExample: `123456789 50`", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmSetStorageForUser(ctx *bot.Context, uid int64) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateSetStorage)
	AdminFSM.SetData(ctx.SenderID(), "target_id", uid)
	_, _ = ctx.Reply(fmt.Sprintf("*Set Storage for* `%d`\n\nSend GB amount (e.g. `50`):", uid), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmSetStorageDo(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateSetStorage) {
		return nil
	}

	targetID := AdminFSM.GetIntData(senderID, "target_id")
	parts := strings.Fields(strings.TrimSpace(ctx.Text()))

	if targetID == 0 {
		if len(parts) != 2 || !isDigit(parts[0]) || !isDigit(parts[1]) {
			_, _ = ctx.Reply("Format: `USER_ID STORAGE_GB`")
			return nil
		}
		targetID, _ = strconv.ParseInt(parts[0], 10, 64)
		gb, _ := strconv.ParseInt(parts[1], 10, 64)
		_ = db.UpdateUserStorageLimit(ctx.Context(), targetID, gb)
		AdminFSM.End(senderID)
		_, _ = ctx.Reply(fmt.Sprintf("Storage for `%d` set to *%d GB* (absolute).", targetID, gb), bot.WithParseMode("Markdown"))
		_, _ = instance.MainBot.SendMessage(targetID, fmt.Sprintf("Your storage limit has been set to *%d GB* by an admin.", gb), bot.WithParseMode("Markdown"))
	} else {
		if len(parts) != 1 || !isDigit(parts[0]) {
			_, _ = ctx.Reply("Send GB number only.")
			return nil
		}
		gb, _ := strconv.ParseInt(parts[0], 10, 64)
		_ = db.UpdateUserStorageLimit(ctx.Context(), targetID, gb)
		AdminFSM.End(senderID)
		_, _ = ctx.Reply(fmt.Sprintf("Storage for `%d` set to *%d GB* (absolute).", targetID, gb), bot.WithParseMode("Markdown"))
		_, _ = instance.MainBot.SendMessage(targetID, fmt.Sprintf("Your storage limit has been set to *%d GB* by an admin.", gb), bot.WithParseMode("Markdown"))
	}
	return nil
}

func AdmAddBonusPrompt(ctx *bot.Context, uid int64) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateAddBonus)
	AdminFSM.SetData(ctx.SenderID(), "target_id", uid)
	_, _ = ctx.Reply(fmt.Sprintf("*Add Bonus to* `%d`\n\nSend GB amount (e.g. `10`):", uid), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmAddBonusGeneral(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateAddBonus)
	AdminFSM.SetData(ctx.SenderID(), "target_id", int64(0))
	_, _ = ctx.Reply("*Add Bonus Storage*\n\nSend: `USER_ID GB`", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmAddBonusDo(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateAddBonus) {
		return nil
	}

	targetID := AdminFSM.GetIntData(senderID, "target_id")
	parts := strings.Fields(strings.TrimSpace(ctx.Text()))

	if targetID == 0 {
		if len(parts) != 2 || !isDigit(parts[0]) || !isDigit(parts[1]) {
			_, _ = ctx.Reply("Format: `USER_ID GB`")
			return nil
		}
		targetID, _ = strconv.ParseInt(parts[0], 10, 64)
		gb, _ := strconv.ParseInt(parts[1], 10, 64)
		_ = db.UpdateUserLimit(ctx.Context(), targetID, gb)
		AdminFSM.End(senderID)
		_, _ = ctx.Reply(fmt.Sprintf("Added *+%d GB* bonus to `%d`.", gb, targetID), bot.WithParseMode("Markdown"))
		_, _ = instance.MainBot.SendMessage(targetID, fmt.Sprintf("*Bonus Alert!* +%d GB storage added by admin!", gb), bot.WithParseMode("Markdown"))
	} else {
		if len(parts) != 1 || !isDigit(parts[0]) {
			_, _ = ctx.Reply("Send GB number only.")
			return nil
		}
		gb, _ := strconv.ParseInt(parts[0], 10, 64)
		_ = db.UpdateUserLimit(ctx.Context(), targetID, gb)
		AdminFSM.End(senderID)
		_, _ = ctx.Reply(fmt.Sprintf("Added *+%d GB* bonus to `%d`.", gb, targetID), bot.WithParseMode("Markdown"))
		_, _ = instance.MainBot.SendMessage(targetID, fmt.Sprintf("*Bonus Alert!* +%d GB storage added by admin!", gb), bot.WithParseMode("Markdown"))
	}
	return nil
}

func AdmPlans(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	plans, _ := db.GetAllPlans(ctx.Context())

	var b strings.Builder
	b.WriteString("*All Plans*\n\n")
	if len(plans) == 0 {
		b.WriteString("No plans yet.")
	} else {
		for _, p := range plans {
			status := "Active"
			if !p.IsActive {
				status = "Inactive"
			}
			b.WriteString(fmt.Sprintf("- *%s* - %d GB - Rs.%.0f - %d days - %s\n",
				p.Name, p.StorageGB, p.PriceINR, p.DurationDays, status))
		}
	}

	markup := bot.NewInlineKeyboard()
	for _, p := range plans {
		if p.IsActive {
			markup.InlineKeyboard = append(markup.InlineKeyboard,
				[]bot.InlineKeyboardButton{
					bot.InlineButton(fmt.Sprintf("Delete: %s", p.Name), fmt.Sprintf("adm_delplan_%d", p.ID)),
				})
		}
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard,
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Create Plan", "adm_create_plan"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back", "admin_panel"),
		})

	_, _ = ctx.Reply(b.String(), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func AdmDelPlan(ctx *bot.Context, planID int) error {
	_ = ctx.AnswerCallback("Plan deactivated", false)
	_ = db.DeletePlan(ctx.Context(), planID)
	return AdmPlans(ctx)
}

func AdmCreatePlanStart(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateCreatePlanName)
	_, _ = ctx.Reply("*Create Plan*\n\n*Step 1/5:* Enter plan *name*:", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_plans")))
	return nil
}

func AdmCreatePlanName(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreatePlanName) {
		return nil
	}
	AdminFSM.SetData(senderID, "plan_name", strings.TrimSpace(ctx.Text()))
	AdminFSM.SetState(senderID, StateCreatePlanPrice)
	_, _ = ctx.Reply("*Step 2/5:* Enter *price* in Rs (e.g. `99`):", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_plans")))
	return nil
}

func AdmCreatePlanPrice(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreatePlanPrice) {
		return nil
	}
	price, ok := parseFloat(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Enter a number.")
		return nil
	}
	AdminFSM.SetData(senderID, "plan_price", price)
	AdminFSM.SetState(senderID, StateCreatePlanStorage)
	_, _ = ctx.Reply("*Step 3/5:* Enter *storage in GB* (e.g. `50`):", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_plans")))
	return nil
}

func AdmCreatePlanStorage(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreatePlanStorage) {
		return nil
	}
	gb, ok := parseInt(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Enter a whole number.")
		return nil
	}
	AdminFSM.SetData(senderID, "plan_storage", gb)
	AdminFSM.SetState(senderID, StateCreatePlanDays)
	_, _ = ctx.Reply("*Step 4/5:* Enter *duration in days* (e.g. `30`):", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_plans")))
	return nil
}

func AdmCreatePlanDays(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreatePlanDays) {
		return nil
	}
	days, ok := parseInt(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Enter a whole number.")
		return nil
	}
	AdminFSM.SetData(senderID, "plan_days", days)
	AdminFSM.SetState(senderID, StateCreatePlanDesc)
	_, _ = ctx.Reply("*Step 5/5:* Enter plan *description* (or `skip`):", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_plans")))
	return nil
}

func AdmCreatePlanDesc(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreatePlanDesc) {
		return nil
	}

	data := AdminFSM.Get(senderID).Data
	if data == nil {
		AdminFSM.End(senderID)
		_, _ = ctx.Reply("Session expired. Please start again.")
		return nil
	}
	planName := safeString(data["plan_name"])
	planPrice, _ := data["plan_price"].(float64)
	planStorage, _ := data["plan_storage"].(int64)
	planDays, _ := data["plan_days"].(int64)
	desc := strings.TrimSpace(ctx.Text())
	if strings.ToLower(desc) == "skip" {
		desc = ""
	}

	planID, err := db.CreatePlan(ctx.Context(), planName, desc, planPrice, planStorage, planDays)
	AdminFSM.End(senderID)
	if err != nil {
		_, _ = ctx.Reply("Failed to create plan.")
		return nil
	}

	text := fmt.Sprintf("*Plan Created!*\n\n*ID:* %d\n*Name:* %s\n*Price:* Rs.%.0f\n*Storage:* %d GB\n*Duration:* %d days", planID, planName, planPrice, planStorage, planDays)
	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_plans")))
	return nil
}

func AdmPayments(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	payments, _ := db.GetPendingPayments(ctx.Context())

	var b strings.Builder
	b.WriteString(fmt.Sprintf("*Pending Payments (%d)*\n\n", len(payments)))
	if len(payments) == 0 {
		b.WriteString("No pending payments.")
	} else {
		for _, p := range payments {
			b.WriteString(fmt.Sprintf("- #%d | %s | %s | Rs.%.0f\n", p.ID, p.FirstName, p.PlanName, p.Amount))
		}
	}

	markup := bot.NewInlineKeyboard()
	for _, p := range payments {
		markup.InlineKeyboard = append(markup.InlineKeyboard,
			[]bot.InlineKeyboardButton{
				bot.InlineButton(fmt.Sprintf("#%d - %s", p.ID, truncate(p.FirstName, 10)),
					fmt.Sprintf("adm_pay_detail_%d", p.ID)),
			})
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard,
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back", "admin_panel"),
		})

	_, _ = ctx.Reply(b.String(), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func AdmPayDetail(ctx *bot.Context, paymentID int) error {
	_ = ctx.AnswerCallback("", false)
	payment, _ := db.GetPayment(ctx.Context(), paymentID)
	if payment == nil {
		_, _ = ctx.Reply("Payment not found.")
		return nil
	}

	text := fmt.Sprintf("*Payment #%d*\n\n*User:* %s (%d)\n*Plan:* %s\n*Amount:* Rs.%.0f\n*Date:* %s\n\nScreenshot attached above (if visible)",
		paymentID, payment.FirstName, payment.UserID, payment.PlanName, payment.Amount,
		payment.CreatedAt)

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Approve", fmt.Sprintf("pay_approve_%d", paymentID)),
			bot.InlineButton("Reject", fmt.Sprintf("pay_reject_%d", paymentID)),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back", "adm_payments"),
		},
	)

	if payment.ScreenshotFileID != "" {
		if instance.MainBot != nil {
			if _, err := instance.MainBot.SendPhoto(ctx.ChatID(), payment.ScreenshotFileID, text, bot.WithPhotoCaption(text)); err != nil {
				log.Printf("[ADMIN] MainBot photo send failed: %v", err)
				_, _ = ctx.Reply(text+"\n\n*Screenshot unavailable - may have expired or from different bot*", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
			} else {
				_, _ = ctx.Reply("*Approve* or *Reject* below:", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
			}
		} else {
			_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
		}
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}

func AdmPayApprove(ctx *bot.Context, paymentID int) error {
	_ = ctx.AnswerCallback("", false)
	_ = db.ApprovePayment(ctx.Context(), paymentID)
	_, _ = ctx.Reply(fmt.Sprintf("Payment #%d approved.", paymentID), bot.WithParseMode("Markdown"))
	return AdmPayments(ctx)
}

func AdmPayReject(ctx *bot.Context, paymentID int) error {
	_ = ctx.AnswerCallback("", false)
	_ = db.RejectPayment(ctx.Context(), paymentID)
	_, _ = ctx.Reply(fmt.Sprintf("Payment #%d rejected.", paymentID), bot.WithParseMode("Markdown"))
	return AdmPayments(ctx)
}

func AdmGifts(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	codes, _ := db.GetAllGiftCodes(ctx.Context())

	var b strings.Builder
	b.WriteString(fmt.Sprintf("*Gift Codes (%d)*\n\n", len(codes)))
	if len(codes) == 0 {
		b.WriteString("No gift codes yet.")
	} else {
		for _, c := range codes {
			status := "Available"
			if c.IsFullyClaimed() {
				status = fmt.Sprintf("Used %d/%d", c.ClaimCount, c.MaxClaims)
			} else if c.IsExpired() {
				status = "Expired"
			}
			b.WriteString(fmt.Sprintf("- `%s` - %d GB - %s\n", c.Code, c.StorageGB, status))
		}
	}

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Create Gift Code", "adm_create_gift"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back", "admin_panel"),
		},
	)

	_, _ = ctx.Reply(b.String(), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func AdmCreateGiftPrompt(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateCreateGift)
	_, _ = ctx.Reply("*Create Gift Code*\n\n*Step 1/3:* Enter GB amount (e.g. `10`):", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_gifts")))
	return nil
}

func AdmCreateGiftDo(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreateGift) {
		return nil
	}
	gb, ok := parseInt(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Send a number (GB).")
		return nil
	}
	AdminFSM.SetData(senderID, "storage_gb", int64(gb))
	AdminFSM.SetState(senderID, StateCreateGiftExpiry)
	_, _ = ctx.Reply(fmt.Sprintf("*Gift Code*\n\nStep 1/3: Storage set to *%d GB*\n*Step 2/3:* Enter expiry in days (`0` for permanent):", gb), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_gifts")))
	return nil
}

func AdmCreateGiftExpiry(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreateGiftExpiry) {
		return nil
	}
	days, ok := parseInt(ctx.Text())
	if !ok {
		_, _ = ctx.Reply("Enter a number for days.")
		return nil
	}
	AdminFSM.SetData(senderID, "expiry_days", int64(days))
	AdminFSM.SetState(senderID, StateCreateGiftClaims)
	expiryText := "Permanent"
	if days > 0 {
		expiryText = fmt.Sprintf("%d days", days)
	}
	_, _ = ctx.Reply(fmt.Sprintf("*Gift Code*\n*Step 3/3:* Enter max claims (default `1`):\n_(Expiry set to %s)_", expiryText), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_gifts")))
	return nil
}

func AdmCreateGiftClaims(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateCreateGiftClaims) {
		return nil
	}

	data := AdminFSM.Get(senderID).Data
	if data == nil {
		AdminFSM.End(senderID)
		_, _ = ctx.Reply("Session expired. Please start again.", bot.WithParseMode("Markdown"))
		return nil
	}

	maxClaims := 1
	if n, ok := parseInt(ctx.Text()); ok && n > 0 {
		maxClaims = int(n)
	}

	gbVal, ok := data["storage_gb"]
	if !ok {
		AdminFSM.End(senderID)
		_, _ = ctx.Reply("Session expired. Please start again.", bot.WithParseMode("Markdown"))
		return nil
	}
	gb := int(gbVal.(int64))

	days := int64(0)
	if daysVal, ok := data["expiry_days"]; ok && daysVal != nil {
		days = daysVal.(int64)
	}

	var expiresAt *time.Time
	if days > 0 {
		t := time.Now().Add(time.Duration(days) * 24 * time.Hour)
		expiresAt = &t
	}

	code := generateGiftCode()
	gift, err := db.CreateGiftCode(ctx.Context(), code, gb, maxClaims, senderID, expiresAt)
	AdminFSM.End(senderID)
	if err != nil {
		_, _ = ctx.Reply("Failed to create gift code.")
		return nil
	}

	expiryText := "Permanent"
	if expiresAt != nil {
		expiryText = expiresAt.Format("2006-01-02")
	}

	text := fmt.Sprintf("*Gift Code Created!*\n\n`%s`\nStorage: %d GB\nExpiry: %s\nMax Claims: %d", code, gift.StorageGB, expiryText, gift.MaxClaims)
	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(backKB("adm_gifts")))
	return nil
}

func AdmBroadcastPrompt(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	AdminFSM.SetState(ctx.SenderID(), StateBroadcast)
	_, _ = ctx.Reply("*Broadcast*\n\nSend the message to broadcast to ALL users:\n_(Supports Markdown formatting)_", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(cancelKB()))
	return nil
}

func AdmBroadcastSend(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !AdminFSM.HasState(senderID, StateBroadcast) {
		return nil
	}
	AdminFSM.End(senderID)

	msg := ctx.Text()
	if msg == "" {
		return nil
	}

	users, _, _ := db.GetAllUsers(ctx.Context(), 1000, 0)
	total := len(users)
	adminChatID := ctx.ChatID()

	go func() {
		sent := 0
		failed := 0

		for i, u := range users {
			if u.IsBanned {
				continue
			}
			_, err := instance.MainBot.SendMessage(u.ID, msg, bot.WithParseMode("Markdown"))
			if err != nil {
				failed++
			} else {
				sent++
			}
			if (i+1)%30 == 0 {
				time.Sleep(1 * time.Second)
			}
		}

		report := fmt.Sprintf("*Broadcast Complete!*\n\n*Sent:* %d\n*Failed:* %d\n*Total:* %d", sent, failed, total)
		if instance.AdminBot != nil {
			_, _ = instance.MainBot.SendMessage(adminChatID, report, bot.WithParseMode("Markdown"))
		}
		log.Printf("Broadcast done. Sent: %d, Failed: %d, Total: %d", sent, failed, total)
	}()

	_, _ = ctx.Reply(fmt.Sprintf("Broadcasting to %d users in background...\nYou'll be notified when done.", total), bot.WithParseMode("Markdown"))
	return nil
}

func AdmCleanup(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("Running cleanup...", false)
	count, _ := db.DeleteExpiredFiles(ctx.Context())
	_, _ = ctx.Reply(fmt.Sprintf("*Cleanup Complete!*\n\nDeleted *%d* expired files.", count), bot.WithParseMode("Markdown"))
	return nil
}

func showUserDetail(ctx *bot.Context, user *models.User) error {
	uname := "N/A"
	if user.Username != nil {
		uname = "@" + *user.Username
	}
	displayName := "N/A"
	if user.Name != nil {
		displayName = *user.Name
	}
	limitStr := fmt.Sprintf("%d GB", user.StorageLimitGB)
	if user.StorageLimitGB == -1 {
		limitStr = "Unlimited"
	}
	storageUsed := utils.FormatBytes(user.StorageUsedBytes)
	percent := user.StoragePercentUsed()
	refCount, _ := db.GetReferralCount(ctx.Context(), user.ID)
	fileCount, _ := db.GetUserFilesCount(ctx.Context(), user.ID)

	text := fmt.Sprintf("*User Details*\n\n*ID:* %d\n*Name:* %s\n*Username:* %s\n*Joined:* %s\n*Banned:* %s\n\n*Storage:* %s / %s *(%.1f%%)*\n*Files:* %d\n*Referrals:* %d",
		user.ID, displayName, uname, user.CreatedAt.Format("2006-01-02"),
		yesNo(user.IsBanned), storageUsed, limitStr, percent, fileCount, refCount)

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Ban", fmt.Sprintf("adm_ban_do_%d", user.ID)),
			bot.InlineButton("Unban", fmt.Sprintf("adm_unban_do_%d", user.ID)),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Set Storage", fmt.Sprintf("adm_setstg_%d", user.ID)),
			bot.InlineButton("Add Bonus", fmt.Sprintf("adm_addbonus_%d", user.ID)),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back", "admin_panel"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func generateGiftCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

func parseUserID(s string) int64 {
	uid, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return uid
}

func parseID(s string) (int64, bool) {
	uid, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return uid, err == nil
}

func parseInt(s string) (int64, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n, err == nil
}

func parseFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f, err == nil
}

func isDigit(s string) bool {
	_, ok := strconv.ParseInt(s, 10, 64)
	return ok == nil
}

func safeString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func backKB(callback string) *bot.InlineKeyboardMarkup {
	return bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back", callback),
		},
	)
}

func cancelKB() *bot.InlineKeyboardMarkup {
	return bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Cancel", "admin_panel"),
		},
	)
}
