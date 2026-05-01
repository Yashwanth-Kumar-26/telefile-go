package handlers

import (
	"context"
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/config"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// ShowReferral shows referral information
func ShowReferral(ctx *bot.Context, cfg *config.Config) error {
	senderID := ctx.SenderID()

	refCount, _ := db.GetReferralCount(context.Background(), senderID)

	// Get bot username dynamically
	botUsername := cfg.BotUsername
	referralLink := fmt.Sprintf("https://t.me/%s?start=ref_%d", botUsername, senderID)

	bonusGB := 5

	text := fmt.Sprintf("[Referral Program]\n\nShare your referral link and earn %d GB of free storage for each referral!\n\nYour Referral Link: %s\n\nTotal Referrals: %d\nBonus Earned: %d GB\n\nHow it works:\n1. Share your link with friends\n2. They use /start and sign up\n3. Both get %d GB bonus!\n4. Bonus is added after they upload their first file", bonusGB, referralLink, refCount, refCount*bonusGB, bonusGB)

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButtonURL("Copy Link", referralLink),
			bot.InlineButton("Back to Menu", "start"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}