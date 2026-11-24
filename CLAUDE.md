# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o telefile ./cmd/telefile
./telefile
```

The bot runs in **polling mode** by default. If `WEBHOOK_URL` is set in `.env`, it switches to webhook mode on port `PORT` (default 8080). Webhook endpoints are `/webhook` (main bot) and `/admin/webhook` (admin bot).

To set up webhooks with Telegram:
```bash
curl -F "url=https://yourdomain.com/webhook" -F "secret_token=YOUR_SECRET" \
  "https://api.telegram.org/bot$BOT_TOKEN/setWebhook"
```

## Architecture

### Dual-Bot Model
Two independent `bot.Bot` instances share the same DB and scheduler:
- **Main bot** (`cfg.BotToken`) — serves all users via `cmd/telefile/main.go` routes
- **Admin bot** (`cfg.AdminBotToken`) — separate bot for admin commands (`/admin`, `/stats`), protected by `middleware.AdminOnly`

Both bots are stored in `internal/bot/instance/instance.go` as `MainBot` and `AdminBot` package-level variables so other packages (scheduler, admin handlers) can send messages without a ctx reference.

### Storage Model
Files are stored entirely in Telegram. The DB only holds `telegram_file_id` references — no file bytes ever touch the app server. On download, the bot re-sends the file using the stored `file_id`.

### User State (In-Memory Maps)
Several handlers use package-level map variables for conversational state:
- `handlers.PendingUploads` — batch upload tracking (in `upload.go`)
- `handlers.DownloadStates` — download password prompts (in `download.go`)
- `handlers.RedeemStates`, `handlers.ReportStates` — multi-step flows
- `adminhandlers.UserStates`, `adminhandlers.BroadcastStates`, `adminhandlers.GiftCreateStates` — admin flows

These are lost on restart (no persistence). `PendingUploads` also holds inline keyboard `ControlMsgID` to edit the same message as files are added.

### Middleware Chain (Main Bot)
Two middleware run on every update before handlers:
1. `middleware.BanCheck()` — rejects banned users
2. `middleware.EnsureUser()` — auto-creates DB user on first interaction, also extracts `ref_<userid>` from `/start` params for the referral system

### Router Pattern
TeleRiHa uses a callback-key based router. Every inline keyboard button press triggers a `OnCallback("key")` handler. Commands like `/start` are `On("/start")`. The `Default()` handler catches all unhandled text/documents/photos and routes them into the upload flow or one of the state maps.

### Scheduler
`robfig/cron/v3` runs two jobs:
- Every hour: `CleanupJob` deletes expired files from DB
- Daily at 9 AM: `StorageWarningJob` notifies users above 80% storage quota

## Key DB Tables

| Table | Purpose |
|---|---|
| `users` | TG user, storage quota, referral, ban status |
| `files` | file_id, unique link, password hash, expiry, download count |
| `batches` | groups multiple files into one shareable link |
| `gifts` | promo codes with storage GB and claim limits |
| `subscriptions` | plan-based storage upgrades |
| `referrals` | tracks referrer→referred pairs and bonus status |

## Key Conventions

- `models.File.Password` is stored as plain text (not hashed) — the `IsPasswordProtected()` check compares strings directly
- `storage_limit_gb = -1` is the sentinel for "Unlimited" plans
- File links are generated as `uuid.New().String()[:12]` — 12-char alphanumeric, collision risk is low but there is no uniqueness enforcement at the DB level beyond the `UNIQUE` constraint on `files.unique_link`
- `UploadExpiryHandler` in `upload.go` handles both single and batch uploads (when `batchID != nil`, multiple files share one batch link)