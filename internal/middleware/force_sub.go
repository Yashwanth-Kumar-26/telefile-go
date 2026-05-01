package middleware

import (
	"fmt"

	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// ForceSub middleware checks if user has joined required channel
// Note: This middleware is optional and requires the bot to be a member of the channel
func ForceSub(channelID string) func(bot.Handler) bot.Handler {
	return func(next bot.Handler) bot.Handler {
		return func(ctx *bot.Context) error {
			if channelID == "" {
				return next(ctx)
			}

			senderID := ctx.SenderID()
			if senderID == 0 {
				return next(ctx)
			}

			// Check membership using ctx.Bot.GetChatMember
			var member *bot.ChatMember
			var err error
			// channelID may be like "-100123456789" (chat ID) or "@username"
			if len(channelID) > 0 && channelID[0] == '-' {
				// Numeric chat ID (e.g., -100xxx)
				var chatID int64
				fmt.Sscanf(channelID, "%d", &chatID)
				member, err = ctx.Bot.GetChatMember(chatID, senderID)
			} else {
				// Username, skip check
				member, err = ctx.Bot.GetChatMember(0, senderID)
			}
			if err != nil {
				// If we can't check (e.g., bot not in channel), allow through
				return next(ctx)
			}

			// Member statuses that are allowed
			allowedStatuses := []string{"member", "administrator", "creator"}
			for _, status := range allowedStatuses {
				if member.Status == status {
					return next(ctx)
				}
			}

			// Not a member - show join prompt
			joinURL := fmt.Sprintf("https://t.me/%s", channelID[1:]) // Remove @ from channel ID
			markup := bot.NewInlineKeyboard(
				[]bot.InlineKeyboardButton{
					bot.InlineButtonURL("Join Channel", joinURL),
					bot.InlineButton("Done", "force_sub_done"),
				},
			)

			_, _ = ctx.Reply("\xe2\x9a\xa0 Please join our channel first to use this bot.", bot.WithReplyMarkup(markup))
			return nil
		}
	}
}
