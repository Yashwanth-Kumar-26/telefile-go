package admin

import (
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// UserStates for admin user search
var UserStates = make(map[int64]bool)
var BroadcastStates = make(map[int64]bool)
var GiftCreateStates = make(map[int64]bool)

// AdminPanel shows the admin panel
func AdminPanel(ctx *bot.Context) error {
	text := "[Admin] Choose an option:"

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Bot Stats", "admin_stats"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("User Management", "admin_users"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Gift Codes", "admin_gifts"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Broadcast", "admin_broadcast"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Cleanup", "admin_cleanup"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

// AdminStats shows bot statistics
func AdminStats(ctx *bot.Context) error {
	userCount, _ := db.GetUserCount(ctx.Context())
	fileCount, _ := db.GetFileCount(ctx.Context())
	totalStorage, _ := db.GetTotalStorageUsed(ctx.Context())

	text := fmt.Sprintf("[Bot Statistics] Total Users: %d, Total Files: %d, Total Storage: %s", userCount, fileCount, utils.FormatBytes(totalStorage))

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"))
	return nil
}

// AdminUsers shows user management
func AdminUsers(ctx *bot.Context) error {
	text := "Enter a user ID to lookup, ban, or unban:"

	UserStates[ctx.SenderID()] = true

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Panel", "admin_panel"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

