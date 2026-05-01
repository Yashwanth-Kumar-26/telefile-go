package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken        string
	AdminBotToken   string
	DatabaseURL     string
	AdminIDs        []int64
	BaseStorageGB   int64
	MaxFileSizeMB   int64
	WebhookURL      string
	Port            string
	WebhookSecret   string
	AdsChannelID    string
	BotUsername     string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		BotToken:       getEnv("BOT_TOKEN", ""),
		AdminBotToken:  getEnv("ADMIN_BOT_TOKEN", ""),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		AdminIDs:       parseAdminIDs(getEnv("ADMIN_IDS", "")),
		BaseStorageGB:  getEnvInt64("BASE_STORAGE_GB", 20),
		MaxFileSizeMB:  getEnvInt64("MAX_FILE_SIZE_MB", 4096),
		WebhookURL:     getEnv("WEBHOOK_URL", ""),
		Port:           getEnv("PORT", "8080"),
		WebhookSecret:  getEnv("WEBHOOK_SECRET", ""),
		AdsChannelID:    getEnv("ADS_CHANNEL_ID", ""),
		BotUsername:     getEnv("BOT_USERNAME", "YourBot"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func parseAdminIDs(s string) []int64 {
	if s == "" {
		return nil
	}
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.ParseInt(part, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}