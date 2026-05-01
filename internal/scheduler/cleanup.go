package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron

// CleanupJob runs cleanup every hour
type CleanupJob struct{}

func (c *CleanupJob) Run() {
	ctx := context.Background()

	count, err := db.DeleteExpiredFiles(ctx)
	if err != nil {
		log.Printf("Cleanup error: %v", err)
		return
	}

	if count > 0 {
		log.Printf("Deleted %d expired files", count)
	}
}

// StorageWarningJob runs daily to check storage warnings
type StorageWarningJob struct{}

func (s *StorageWarningJob) Run() {
	users, err := db.GetHighStorageUsers(context.Background(), 80.0)
	if err != nil {
		log.Printf("Error getting high storage users: %v", err)
		return
	}

	for _, user := range users {
		percent := user.StoragePercentUsed()
		log.Printf("Warning: User %d is using %.1f%% storage", user.ID, percent)

		if instance.MainBot != nil {
			msg := fmt.Sprintf("⚠️ *Storage Warning*\n\nYou are using %.1f%% of your storage limit. Please delete some files or upgrade your plan to avoid upload failures.", percent)
			_, err := instance.MainBot.SendMessage(user.ID, msg, bot.WithParseMode("Markdown"))
			if err != nil {
				log.Printf("Failed to send storage warning to %d: %v", user.ID, err)
			}
		}
	}
}

// StartScheduler starts the cron scheduler
func StartScheduler() {
	cronScheduler = cron.New()

	// Run cleanup every hour
	_, err := cronScheduler.AddJob("@every 1h", &CleanupJob{})
	if err != nil {
		log.Printf("Failed to add cleanup job: %v", err)
		return
	}

	// Run storage warning daily at 9 AM
	_, err = cronScheduler.AddFunc("0 9 * * *", func() {
		(&StorageWarningJob{}).Run()
	})
	if err != nil {
		log.Printf("Failed to add storage warning job: %v", err)
		return
	}

	cronScheduler.Start()
	log.Println("Scheduler started")
}

// StopScheduler stops the cron scheduler
func StopScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
	}
}