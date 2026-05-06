package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	adminhandlers "github.com/Yashwanth-Kumar-26/telefile/internal/admin/handlers"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// BuyPlanState stores user's buy plan state
var BuyPlanStates = make(map[int64]*BuyPlanStateData)

type BuyPlanStateData struct {
	PlanID   int
	PlanName string
	Amount   float64
}

// ShowBuyPlanMenu displays plan selection with payment instructions
func ShowBuyPlanMenu(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)

	subs, err := db.GetAllSubscriptions(context.Background())
	if err != nil {
		_, _ = ctx.Reply("❌ Failed to load plans. Please try again.")
		return nil
	}

	text := "💳 *Select a Plan to Buy*\n\n"
	text += "📱 *Payment Process:*\n"
	text += "1️⃣ Select your plan below\n"
	text += "2️⃣ Pay via UPI to *telefile@upi*\n"
	text += "3️⃣ Take screenshot of payment\n"
	text += "4️⃣ Send screenshot here\n"
	text += "5️⃣ We'll verify and activate! ✅\n\n"

	markup := bot.NewInlineKeyboard()

	for _, sub := range subs {
		if sub.PriceINR <= 0 {
			continue
		}

		var features PlanFeatures
		if sub.Features != nil {
			json.Unmarshal(sub.Features, &features)
		}
		durationText := ""
		if features.DurationDays > 0 {
			durationText = fmt.Sprintf(" (%d days)", features.DurationDays)
		}

		markup.InlineKeyboard = append(markup.InlineKeyboard,
			[]bot.InlineKeyboardButton{
				bot.InlineButton(
					fmt.Sprintf("💎 %s — ₹%.0f%s", sub.Name, sub.PriceINR, durationText),
					fmt.Sprintf("bp_select_%d", sub.ID),
				),
			})
	}

	markup.InlineKeyboard = append(markup.InlineKeyboard,
		[]bot.InlineKeyboardButton{
			bot.InlineButton("💰 UPI: telefile@upi", "bp_upi"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("⬅️ Back to Plans", "plans"),
		})

	if ctx.CallbackQuery != nil {
		_, _ = ctx.EditMessageText(text, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}

// BuyPlanSelect handles plan selection
func BuyPlanSelect(ctx *bot.Context, planID int) error {
	_ = ctx.AnswerCallback("", false)

	sub, err := db.GetSubscriptionByID(context.Background(), planID)
	if err != nil || sub == nil {
		_, _ = ctx.Reply("❌ Plan not found.")
		return nil
	}

	if sub.PriceINR <= 0 {
		_, _ = ctx.Reply("❌ This plan is free!")
		return nil
	}

	senderID := ctx.SenderID()
	BuyPlanStates[senderID] = &BuyPlanStateData{
		PlanID:   sub.ID,
		PlanName: sub.Name,
		Amount:   sub.PriceINR,
	}

	var features PlanFeatures
	if sub.Features != nil {
		json.Unmarshal(sub.Features, &features)
	}

	durationText := ""
	if features.DurationDays > 0 {
		durationText = fmt.Sprintf("\n⏰ Valid for *%d days*", features.DurationDays)
	}

	text := fmt.Sprintf("💎 *Selected: %s*\n\n", sub.Name)
	text += fmt.Sprintf("💰 Price: *₹%.0f*\n", sub.PriceINR)
	text += fmt.Sprintf("💾 Storage: *%s*%s\n\n", storageLabel(sub.StorageGB), durationText)
	text += "📱 *Payment Instructions:*\n\n"
	text += fmt.Sprintf("1️⃣ Pay *₹%.0f* to UPI ID:\n   `telefile@upi`\n\n", sub.PriceINR)
	text += "2️⃣ Take a screenshot of payment confirmation\n\n"
	text += "3️⃣ Send the screenshot in this chat\n\n"
	text += "⏳ *We'll verify and activate your plan within 24 hours*"

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("🔄 Cancel", "buy_plan"),
		},
	)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.EditMessageText(text, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}

	adminhandlers.AdminFSM.SetState(senderID, adminhandlers.StateBuyPlanUpload)
	log.Printf("[BUY_PLAN] User %d selected plan %s (ID: %d), waiting for screenshot", senderID, sub.Name, planID)

	return nil
}

func storageLabel(gb int) string {
	if gb == -1 {
		return "Unlimited"
	}
	if gb >= 1000 {
		return fmt.Sprintf("%d TB", gb/1000)
	}
	return fmt.Sprintf("%d GB", gb)
}

// BuyPlanHandlePhoto handles payment screenshot upload
func BuyPlanHandlePhoto(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	log.Printf("[BUY_PLAN] Photo received from user %d", senderID)
	
	if !adminhandlers.AdminFSM.HasState(senderID, adminhandlers.StateBuyPlanUpload) {
		log.Printf("[BUY_PLAN] User %d not in buy plan state", senderID)
		return nil
	}

	state, exists := BuyPlanStates[senderID]
	if !exists {
		log.Printf("[BUY_PLAN] User %d no state found", senderID)
		adminhandlers.AdminFSM.End(senderID)
		return nil
	}

	// Get photo file ID
	fileID := ""
	if ctx.Message.Photo != nil && len(ctx.Message.Photo) > 0 {
		fileID = ctx.Message.Photo[len(ctx.Message.Photo)-1].FileID
	} else if ctx.Message.Document != nil {
		fileID = ctx.Message.Document.FileID
	}

	if fileID == "" {
		log.Printf("[BUY_PLAN] User %d no file ID found", senderID)
		_, _ = ctx.Reply("❌ Please send a screenshot of your payment.")
		return nil
	}

	firstName := "User"
	if ctx.Sender != nil && ctx.Sender.FirstName != "" {
		firstName = ctx.Sender.FirstName
	}

	log.Printf("[BUY_PLAN] Creating payment for user %d, plan %s, amount %.0f", senderID, state.PlanName, state.Amount)
	
	paymentID, err := db.CreatePendingPayment(ctx.Context(), senderID, state.PlanID, state.PlanName, state.Amount, fileID, firstName)
	if err != nil {
		log.Printf("[BUY_PLAN] ERROR creating payment: %v", err)
		adminhandlers.AdminFSM.End(senderID)
		_, _ = ctx.Reply(fmt.Sprintf("❌ Failed to submit payment: %v", err))
		return nil
	}

	adminhandlers.AdminFSM.End(senderID)
	delete(BuyPlanStates, senderID)

	text := "✅ *Payment Submitted!*\n\n"
	text += fmt.Sprintf("📋 *Plan:* %s\n", state.PlanName)
	text += fmt.Sprintf("💰 *Amount:* ₹%.0f\n", state.Amount)
	text += fmt.Sprintf("🔢 *Payment ID:* #%d\n\n", paymentID)
	text += "⏳ *Verification in progress...*\n"
	text += "You'll be notified once an admin reviews your payment.\n\n"
	text += "Thank you for your purchase! 🎉"

	markup := bot.NewInlineKeyboard()
	markup.InlineKeyboard = append(markup.InlineKeyboard,
		[]bot.InlineKeyboardButton{
			bot.InlineButton("🏠 Main Menu", "start"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))

	// Notify admins (async, don't block)
	go notifyAdminsPayment(context.Background(), senderID, firstName, state.PlanName, state.Amount, fileID, paymentID)

	return nil
}

func BuyPlanHandleCancel(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	adminhandlers.AdminFSM.End(senderID)
	delete(BuyPlanStates, senderID)
	return ShowBuyPlanMenu(ctx)
}

func notifyAdminsPayment(ctx context.Context, userID int64, firstName, planName string, amount float64, fileID string, paymentID int) {
	admins := []int64{1096262005}

	text := "💳 *New Payment!*\n\n"
	text += fmt.Sprintf("👤 *User:* %s (`%d`)\n", firstName, userID)
	text += fmt.Sprintf("📦 *Plan:* %s\n", planName)
	text += fmt.Sprintf("💰 *Amount:* ₹%.0f\n", amount)
	text += fmt.Sprintf("🔢 *ID:* #%d\n\n", paymentID)
	text += "Check details in admin: /admin"

	// Use MAIN bot since it received the screenshot (file_id works with main bot only)
	for _, adminID := range admins {
		if instance.MainBot != nil {
			_, err := instance.MainBot.SendPhoto(adminID, fileID, text, bot.WithPhotoCaption(text))
			if err != nil {
				log.Printf("[BUY_PLAN] Notify via main bot failed: %v", err)
			} else {
				log.Printf("[BUY_PLAN] Admin %d notified via main bot", adminID)
			}
		}
	}
}

func ShowUPIQR(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)

	text := "💰 *UPI Payment*\n\n"
	text += "📱 *Pay to:* `telefile@upi`\n\n"
	text += "You can use any UPI app:\n"
	text += "• Google Pay\n"
	text += "• PhonePe\n"
	text += "• Paytm\n"
	text += "• BHIM\n\n"
	text += "After payment, take a screenshot and send it here!"

	markup := bot.NewInlineKeyboard()
	markup.InlineKeyboard = append(markup.InlineKeyboard,
		[]bot.InlineKeyboardButton{
			bot.InlineButton("⬅️ Back", "buy_plan"),
		},
	)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.EditMessageText(text, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}
