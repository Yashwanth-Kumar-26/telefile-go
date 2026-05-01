package middleware

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// BanCheck middleware checks if the user is banned
func BanCheck() func(bot.Handler) bot.Handler {
	return func(next bot.Handler) bot.Handler {
		return func(ctx *bot.Context) error {
			senderID := ctx.SenderID()
			if senderID == 0 {
				return next(ctx)
			}

			user, err := db.GetUser(context.Background(), senderID)
			if err != nil {
				// User doesn't exist yet, allow through
				return next(ctx)
			}

			if user.IsBanned {
				_, _ = ctx.Reply("\xe2\x9b\x94 Your account has been suspended. Contact support.")
				return nil
			}

			return next(ctx)
		}
	}
}

// IsBanned checks if a user is banned
func IsBanned(ctx context.Context, userID int64) (bool, error) {
	user, err := db.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsBanned, nil
}

// EnsureUser middleware ensures the user exists in the database
func EnsureUser(baseStorageGB int64) func(bot.Handler) bot.Handler {
	return func(next bot.Handler) bot.Handler {
		return func(ctx *bot.Context) error {
			senderID := ctx.SenderID()
			if senderID == 0 {
				return next(ctx)
			}

			var username string
			var name string
			if ctx.Sender != nil {
				if ctx.Sender.Username != "" {
					username = ctx.Sender.Username
				}
				if ctx.Sender.FirstName != "" {
					name = ctx.Sender.FirstName
				}
			}

			// Get referred by from start parameter
			var referredBy *int64
			if text := ctx.Text(); text != "" {
				// Format: /start ref_{userid} or /start file_{link}
				parts := strings.SplitN(text, " ", 2)
				if len(parts) > 1 {
					param := parts[1]
					if strings.HasPrefix(param, "ref_") {
						refID, _ := strconv.ParseInt(strings.TrimPrefix(param, "ref_"), 10, 64)
						if refID > 0 {
							referredBy = &refID
						}
					}
				}
			}

			_, err := db.GetOrCreateUser(context.Background(), senderID, username, name, referredBy, baseStorageGB)
			if err != nil {
				log.Printf("Error ensuring user: %v\n", err)
			}

			return next(ctx)
		}
	}
}
