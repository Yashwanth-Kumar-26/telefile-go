package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// Start handles the /start command
func Start(ctx *bot.Context) error {
	updateType := "unknown"
	if ctx.CallbackQuery != nil {
		updateType = "callback"
	} else if ctx.Message != nil {
		updateType = "message"
	}
	log.Printf("[DEBUG] Start called - Type: %s, Text: %q", updateType, ctx.Text())

	param := parseStartParam(ctx.Text())
	log.Printf("[DEBUG] Start param: %q", param)

	// Handle file download link
	if strings.HasPrefix(param, "file_") {
		return HandleFileDownload(ctx, strings.TrimPrefix(param, "file_"))
	}

	// Handle batch download link
	if strings.HasPrefix(param, "batch_") {
		return HandleBatchDownload(ctx, strings.TrimPrefix(param, "batch_"))
	}

	// Handle referral - handled in EnsureUser middleware



	welcomeMsg := fmt.Sprintf("👋 *Welcome back to TeleFile!*\n\nStore & share files with advanced features. 🔒\nPassword protection, expiry, & more!\n\nUse the buttons below to get started:")

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Upload File 📤", "upload_help"),
			bot.InlineButton("Batch Upload 📦", "upload_batch"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("My Files 📂", "myfiles"),
			bot.InlineButton("Plans 💎", "plans"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Referral 👥", "referral"),
			bot.InlineButton("My Stats 📊", "my_stats"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Redeem Code 🎁", "redeem"),
			bot.InlineButton("Report Issue 🐛", "report"),
		},
	)

	if ctx.CallbackQuery != nil {
		log.Printf("[DEBUG] Start - using EditMessageText")
		_, _ = ctx.EditMessageText(welcomeMsg, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		log.Printf("[DEBUG] Start - using Reply")
		_, _ = ctx.Reply(welcomeMsg, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}

// ShowStats shows user statistics
func ShowStats(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	user, err := db.GetUser(context.Background(), senderID)
	if err != nil {
		return nil
	}

	fileCount, _ := db.GetUserFileCount(context.Background(), senderID)
	refCount, _ := db.GetReferralCount(context.Background(), senderID)

	storageUsed := user.StorageUsedBytes
	storageLimit := user.StorageLimitGB * 1024 * 1024 * 1024
	if user.StorageLimitGB == -1 {
		storageLimit = 0 // Infinite
	}

	progress := utils.MakeProgressBar(storageUsed, storageLimit)
	usedStr := utils.FormatBytes(storageUsed)
	limitStr := "Unlimited"
	if user.StorageLimitGB != -1 {
		limitStr = utils.FormatBytes(storageLimit)
	}

	text := fmt.Sprintf("📊 *Your Statistics*\n\n*Name:* %s\n*ID:* `%d` \n\n*Files:* %d\n*Storage Used:* %s\n*Storage Limit:* %s\n%s\n\n*Referrals:* %d",
		user.DisplayName(), senderID, fileCount, usedStr, limitStr, progress, refCount)

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Menu 🏠", "start"),
		},
	)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.EditMessageText(text, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}

// Help shows help information
func Help(ctx *bot.Context) error {
	text := "[TeleFile Help]\n\n*How to use:*\n1. Send any file to upload it\n2. Set optional password & expiry\n3. Share the generated link\n\n*Commands:*\n/start - Main menu\n/myfiles - View your files\n/plans - View subscription plans\n/referral - Get your referral link\n/redeem - Redeem a gift code\n/report - Report an issue\n/cancel - Cancel current action\n\n*File Limits:*\n- Max file size depends on your plan\n- Files expire based on your settings\n- Links work for anyone with the password"

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Menu", "start"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func parseStartParam(text string) string {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
