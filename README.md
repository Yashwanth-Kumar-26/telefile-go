# TeleFile - Telegram File Storage Bot
> Store and share files for free using Telegram itself as storage.

## Features

- **File Uploads** - Send any file directly to the bot
- **Password Protection** - Optional per-file passwords
- **Expiry Options** - Set auto-expiry: 1h, 24h, 7d, 30d, or never
- **Unique Download Links** - Share files with unique links
- **Storage Quotas** - Per-user storage limits based on plan
- **Subscription Plans** - Free, Basic, Pro, Unlimited
- **Referral System** - Earn 5 GB for each friend
- **Gift Codes** - Admin-created storage bonus codes
- **Admin Panel** - Full bot management dashboard

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Framework | TeleRiHa v0.1.0 |
| Database | PostgreSQL (Neon) via pgx/v5 |
| Webhook | Go net/http |
| Scheduler | robfig/cron/v3 |
| Config | godotenv |

## Quick Start

```bash
# 1. Clone
git clone https://github.com/Yashwanth-26/telefile.git
cd telefile

# 2. Database - create at neon.tech, then migrate
migrate -path ./migrations -database "$DATABASE_URL" up

# 3. Configure
cp .env.EXAMPLE .env
# Edit .env with your values

# 4. Run
go build -o telefile ./cmd/telefile
./telefile

# Or with webhooks - set in .env first
curl -F "url=https://yourdomain.com/webhook" \
     -F "secret_token=YOUR_SECRET" \
     "https://api.telegram.org/bot$BOT_TOKEN/setWebhook"
```

### Environment Variables (.env)

```env
BOT_TOKEN=your_bot_token_from_botfather
ADMIN_BOT_TOKEN=separate_admin_bot_token
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
ADMIN_IDS=123456789,987654321
BASE_STORAGE_GB=20
MAX_FILE_SIZE_MB=4096
WEBHOOK_URL=https://yourdomain.com
PORT=8080
```

## Bot Commands

### User
- `/start` - Main menu
- `/myfiles` - Your uploaded files
- `/plans` - Subscription plans
- `/referral` - Referral link & stats
- `/redeem` - Redeem gift code
- `/report` - Report issue
- `/cancel` - Cancel conversation

### Admin
- `/admin` - Admin panel
- `/stats` - Bot statistics

## Architecture

This bot uses Telegram itself as file storage by storing only `file_id` references - actual files are never stored on disk. Files are forwarded on download using Telegram's Bot API.

## License

MIT