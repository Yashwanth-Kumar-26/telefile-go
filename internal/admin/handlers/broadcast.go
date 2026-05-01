package admin

import (
	"fmt"
	"log"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// StartBroadcast starts the broadcast flow
func StartBroadcast(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	BroadcastStates[senderID] = true

	_, _ = ctx.Reply("\xf0\x9f\x93\xa2 *Broadcast Message*\n\nSend the message you want to broadcast to all users:")
	return nil
}

// HandleBroadcast handles the broadcast message
func HandleBroadcast(ctx *bot.Context) error {
	senderID := ctx.SenderID()

	if !BroadcastStates[senderID] {
		return nil
	}

	message := ctx.Text()
	if message == "" {
		return nil
	}

	delete(BroadcastStates, senderID)

	go func() {
		offset := 0
		limit := 100
		totalSent := 0

		for {
			users, _, err := db.GetAllUsers(ctx.Context(), limit, offset)
			if err != nil || len(users) == 0 {
				break
			}

			for _, user := range users {
				if instance.MainBot != nil {
					_, err := instance.MainBot.SendMessage(user.ID, message)
					if err != nil {
						log.Printf("Failed to broadcast to %d: %v", user.ID, err)
					} else {
						totalSent++
					}
				}
				// Avoid hitting rate limits
				time.Sleep(50 * time.Millisecond)
			}

			if len(users) < limit {
				break
			}
			offset += limit
		}
		log.Printf("Broadcast finished. Sent to %d users.", totalSent)
	}()

	_, _ = ctx.Reply(fmt.Sprintf("✅ Broadcast started! Sending to users in background..."))
	return nil
}

// AdminCleanup runs manual cleanup
func AdminCleanup(ctx *bot.Context) error {
	count, err := db.DeleteExpiredFiles(ctx.Context())
	if err != nil {
		return ctx.AnswerCallback("Cleanup failed: "+err.Error(), true)
	}
	return ctx.AnswerCallback(fmt.Sprintf("Cleanup finished. Deleted %d files.", count), true)
}

// Also export the states for main.go
func GetUserStates() map[int64]bool { return UserStates }
func GetBroadcastStates() map[int64]bool { return BroadcastStates }
