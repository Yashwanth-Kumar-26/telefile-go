package middleware

import (
	"github.com/Yashwanth-Kumar-26/telefile/internal/config"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// AdminOnly middleware checks if the user is an admin
func AdminOnly(cfg *config.Config) func(bot.Handler) bot.Handler {
	return func(next bot.Handler) bot.Handler {
		return func(ctx *bot.Context) error {
			senderID := ctx.SenderID()
			if senderID == 0 {
				_, _ = ctx.Reply("Unable to identify user.")
				return nil
			}

			for _, adminID := range cfg.AdminIDs {
				if senderID == adminID {
					return next(ctx)
				}
			}

			_, _ = ctx.Reply("Access denied. Admin only.")
			return nil
		}
	}
}

// IsAdmin checks if a user ID is an admin
func IsAdmin(userID int64, cfg *config.Config) bool {
	for _, adminID := range cfg.AdminIDs {
		if userID == adminID {
			return true
		}
	}
	return false
}
