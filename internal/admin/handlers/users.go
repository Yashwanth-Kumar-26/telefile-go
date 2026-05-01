package admin

import (
	"fmt"
	"strconv"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// AdminUserLookup looks up a user by ID
func AdminUserLookup(ctx *bot.Context) error {
	senderID := ctx.SenderID()

	if !UserStates[senderID] {
		return nil
	}

	userIDStr := ctx.Text()
	delete(UserStates, senderID)

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		_, _ = ctx.Reply("Invalid user ID format.")
		return nil
	}

	user, err := db.GetUser(ctx.Context(), userID)
	if err != nil {
		_, _ = ctx.Reply("User not found.")
		return nil
	}

	username := "N/A"
	if user.Username != nil {
		username = *user.Username
	}

	text := fmt.Sprintf("[User Profile] ID: %d, Username: @%s, Storage: %s / %d GB, Limit: %d GB, Banned: %v, Created: %s, Last Active: %s",
		userID, username, utils.FormatBytes(int64(user.StorageUsedBytes)), user.StorageLimitGB, user.StorageLimitGB, user.IsBanned,
		user.CreatedAt.Format("2006-01-02 15:04"), user.LastActive.Format("2006-01-02 15:04"))

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton(fmt.Sprintf("Ban User %d", userID), fmt.Sprintf("admin_ban_%d", userID)),
			bot.InlineButton(fmt.Sprintf("Unban User %d", userID), fmt.Sprintf("admin_unban_%d", userID)),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Panel", "admin_panel"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

// AdminBan bans a user
func AdminBan(ctx *bot.Context, userIDStr string) error {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return ctx.AnswerCallback("Invalid user ID", false)
	}

	if err := db.BanUser(ctx.Context(), userID); err != nil {
		return ctx.AnswerCallback("Failed to ban user", false)
	}

	return ctx.AnswerCallback("User banned", false)
}

// AdminUnban unbans a user
func AdminUnban(ctx *bot.Context, userIDStr string) error {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return ctx.AnswerCallback("Invalid user ID", false)
	}

	if err := db.UnbanUser(ctx.Context(), userID); err != nil {
		return ctx.AnswerCallback("Failed to unban user", false)
	}

	return ctx.AnswerCallback("User unbanned", false)
}