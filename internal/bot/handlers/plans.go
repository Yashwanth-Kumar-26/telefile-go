package handlers

import (
	"context"
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// ShowPlans shows subscription plans
func ShowPlans(ctx *bot.Context) error {
	subs, err := db.GetAllSubscriptions(context.Background())
	if err != nil {
		_, _ = ctx.Reply("Failed to load plans.")
		return nil
	}

	text := "[Subscription Plans]\n\n"

	for _, sub := range subs {
		price := "Free"
		if sub.PriceINR > 0 {
			price = fmt.Sprintf("Rs.%.0f/month", sub.PriceINR)
		}
		if sub.StorageGB == -1 {
			price = fmt.Sprintf("Rs.%.0f/lifetime", sub.PriceINR)
		}

		storage := "Unlimited"
		if sub.StorageGB > 0 {
			storage = fmt.Sprintf("%d GB", sub.StorageGB)
		}

		text += fmt.Sprintf("*%s*\n", sub.Name)
		text += fmt.Sprintf("Storage: %s\n", storage)
		text += fmt.Sprintf("Price: %s\n\n", price)
	}

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Menu", "start"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}