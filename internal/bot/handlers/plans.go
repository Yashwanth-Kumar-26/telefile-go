package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

type PlanFeatures struct {
	Description    string `json:"description"`
	DurationDays   int    `json:"duration_days"`
	MaxFileSize    int    `json:"max_file_size"`
	PrioritySupport bool  `json:"priority_support"`
	AdFree         bool   `json:"ad_free"`
	BatchFiles     int    `json:"batch_files"`
}

func ShowPlans(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)

	subs, err := db.GetAllSubscriptions(context.Background())
	if err != nil {
		_, _ = ctx.Reply("❌ Failed to load plans.")
		return nil
	}

	text := "📦 *TeleFile Plans*\n\n"
	text += "Choose a plan that fits your needs:\n\n"

	for _, sub := range subs {
		price := fmt.Sprintf("₹%.0f", sub.PriceINR)
		if sub.PriceINR == 0 {
			price = "Free"
		}

		storage := "Unlimited"
		if sub.StorageGB > 0 {
			if sub.StorageGB >= 1000 {
				storage = fmt.Sprintf("%d TB", sub.StorageGB/1000)
			} else {
				storage = fmt.Sprintf("%d GB", sub.StorageGB)
			}
		}

		var features PlanFeatures
		if sub.Features != nil {
			json.Unmarshal(sub.Features, &features)
		}

		durationText := ""
		if features.DurationDays > 0 {
			durationText = fmt.Sprintf(" • %d days", features.DurationDays)
		}

		text += fmt.Sprintf("🌟 *%s*\n", sub.Name)
		text += fmt.Sprintf("💰 *%s*/month%s\n", price, durationText)
		text += fmt.Sprintf("💾 %s storage\n", storage)

		if features.Description != "" {
			text += fmt.Sprintf("\n_%s_\n", features.Description)
		}

		featList := []string{}
		if features.AdFree {
			featList = append(featList, "✓ Ad-free")
		}
		if features.PrioritySupport {
			featList = append(featList, "✓ Priority Support")
		}
		if features.BatchFiles != 0 {
			if features.BatchFiles == -1 {
				featList = append(featList, "✓ Unlimited batches")
			} else {
				featList = append(featList, fmt.Sprintf("✓ %d batch files", features.BatchFiles))
			}
		}
		if features.MaxFileSize > 2048 {
			featList = append(featList, fmt.Sprintf("✓ %d MB max file", features.MaxFileSize))
		}

		for _, f := range featList {
			text += fmt.Sprintf("  %s\n", f)
		}
		text += "\n"
	}

	hasPaidPlans := false
	for _, sub := range subs {
		if sub.PriceINR > 0 {
			hasPaidPlans = true
			break
		}
	}

	var markup *bot.InlineKeyboardMarkup
	if hasPaidPlans {
		markup = bot.NewInlineKeyboard(
			[]bot.InlineKeyboardButton{
				bot.InlineButton("💳 Buy Plan", "buy_plan"),
			},
			[]bot.InlineKeyboardButton{
				bot.InlineButton("🏠 Main Menu", "start"),
			},
		)
	} else {
		markup = bot.NewInlineKeyboard(
			[]bot.InlineKeyboardButton{
				bot.InlineButton("🏠 Main Menu", "start"),
			},
		)
	}

	if ctx.CallbackQuery != nil {
		_, _ = ctx.EditMessageText(text, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}
	return nil
}
