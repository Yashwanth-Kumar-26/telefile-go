# TeleFile — Telegram File Storage Bot

> Store and share files for free using Telegram itself as storage. No server-side file storage needed.

## Features

### User
- **File Uploads** — Send any file directly to the bot; stored as Telegram `file_id` references
- **Batch Upload** — Upload multiple files under one shared link
- **Password Protection** — Optional per-file passwords
- **Expiry Options** — Set auto-expiry: 1h, 24h, 7d, 30d, or never
- **Unique Download Links** — Share files via `t.me/<bot>?start=file_<link>`
- **Storage Quotas** — Per-user limits; earn bonus GB via referrals
- **Subscription Plans** — Free (20 GB), Basic, Pro, Unlimited
- **Referral System** — Earn 5 GB for each friend who uploads their first file
- **Gift Codes** — Admin-created storage bonus codes with configurable expiry

### Admin
- **Admin Panel** — `/admin` on the admin bot with inline keyboard navigation; powered by a state-machine (FSM) for multi-step flows
- **Statistics Dashboard** — Total users, new today, banned count, total files, daily uploads, total storage used, active subscriptions, pending payments
- **User Lookup** — Search any user by ID or @username; shows full profile with storage, files, referrals, and quick-action buttons
- **Ban / Unban** — Ban or unban users by ID; banned users receive a direct Telegram DM notification; actions can be triggered from user profile or directly by ID
- **Storage Editing** — Set a user's storage limit to any GB value, or add bonus storage via `USER_ID GB` format; users are notified via DM
- **Plan Management** — List all plans (active + inactive), soft-delete plans, create new plans via 5-step FSM (name → price → storage → duration → description)
- **Pending Payments** — View pending UPI/screenshot payments, see screenshot photo inline, approve (activates plan + adds storage) or reject
- **Gift Code Manager** — List all gift codes with status (available/used/expired), 3-step creation (GB → expiry days → max claims), generates random 8-char codes
- **Broadcast** — Send Markdown-formatted messages to all non-banned users with progress reporting
- **Cleanup** — One-click deletion of all expired files from the database

## Tech Stack

| Layer | Technology |
|---|---|
| Framework | TeleRiHa v0.1.0 |
| Database | PostgreSQL (Neon) via `jackc/pgx/v5` |
| Webhook | Go `net/http` |
| Scheduler | `robfig/cron/v3` |
| Config | `joho/godotenv` |

## Quick Start

```bash
# 1. Clone
git clone https://github.com/Yashwanth-26/telefile.git
cd telefile

# 2. Database — create at neon.tech, then migrate
migrate -path ./migrations -database "$DATABASE_URL" up
# or run setup_db.sql manually

# 3. Configure
cp .env.EXAMPLE .env
# Edit .env with your values

# 4. Build & Run
go build -o telefile ./cmd/telefile
./telefile
```

### Environment Variables (`.env`)

```env
BOT_TOKEN=           # Main bot token from @BotFather
ADMIN_BOT_TOKEN=     # Separate admin bot token
DATABASE_URL=        # postgresql://user:pass@host/db?sslmode=require
ADMIN_IDS=           # Comma-separated Telegram user IDs (no spaces)
BASE_STORAGE_GB=20   # Default storage for new users
MAX_FILE_SIZE_MB=4096
WEBHOOK_URL=         # Optional — set to enable webhook mode
PORT=8080
BOT_USERNAME=        # Bot username (without @)
```

### Webhook Mode

If `WEBHOOK_URL` is set, the bot runs a webhook server instead of polling:

```bash
# Set the webhook
curl -F "url=https://yourdomain.com/webhook" \
     -F "secret_token=YOUR_SECRET" \
     "https://api.telegram.org/bot$BOT_TOKEN/setWebhook"

# Also for the admin bot
curl -F "url=https://yourdomain.com/admin/webhook" \
     -F "secret_token=YOUR_ADMIN_SECRET" \
     "https://api.telegram.org/bot$ADMIN_BOT_TOKEN/setWebhook"
```

## Bot Commands

### User
| Command | Description |
|---|---|
| `/start` | Main menu |
| `/myfiles` | Browse and manage your uploads |
| `/plans` | View subscription plans |
| `/referral` | Get your referral link and stats |
| `/redeem` | Redeem a gift code |
| `/report` | Report an issue |
| `/cancel` | Cancel current action |

### Admin (admin bot)
| Command | Description |
|---|---|
| `/admin` | Open admin panel with inline keyboard menu |
| `/stats` | Quick statistics snapshot |
| `/cancel` | Cancel any in-progress FSM flow |

### Admin Panel Callbacks
| Callback | Action |
|---|---|
| `adm_stats` | View global statistics dashboard |
| `adm_lookup` | Enter FSM: send user ID or @username |
| `adm_ban` | Enter FSM: send user ID to ban |
| `adm_unban` | Enter FSM: send user ID to unban |
| `adm_set_storage` | Enter FSM: set or add storage for a user |
| `adm_add_bonus` | Enter FSM: add bonus storage for a user |
| `adm_plans` | List all plans, soft-delete or create new |
| `adm_create_plan` | 5-step FSM: name → price → storage → duration → description |
| `adm_payments` | View pending payment submissions |
| `adm_pay_detail_<id>` | View payment details, see screenshot, approve/reject |
| `pay_approve_<id>` | Approve payment, activate plan, add storage |
| `pay_reject_<id>` | Reject and delete pending payment |
| `adm_gifts` | List gift codes, create new (3-step FSM) |
| `adm_broadcast` | Enter FSM: send broadcast message to all users |
| `adm_cleanup` | Delete all expired files immediately |
| `adm_ban_do_<id>` | Instant ban from user profile button |
| `adm_unban_do_<id>` | Instant unban from user profile button |
| `adm_setstg_<id>` | Enter FSM to set storage for a specific user |
| `adm_addbonus_<id>` | Enter FSM to add bonus to a specific user |

## Architecture

Files are stored **entirely in Telegram** — the database holds only `file_id` references. On download, the bot re-sends the file using the stored ID. This means:
- No file bytes ever touch the app server
- Download links work as long as the file exists in Telegram's cache
- Batch links group multiple files under one shareable URL

Two independent `Bot` instances share the same DB and scheduler:
- **Main bot** (`BOT_TOKEN`) — serves users
- **Admin bot** (`ADMIN_BOT_TOKEN`) — admin panel, separate from user conversations

### Admin FSM

The admin panel uses an in-memory state machine (`AdminFSMType`) to handle multi-step flows without needing database persistence. States are identified by string constants (e.g. `StateLookupUser`, `StateCreatePlanName`). When a state is active, the `Default` handler in `main.go` routes incoming text to the appropriate `Adm*Do` handler. States and FSM data are cleared via `AdminFSM.End(senderID)` or `/cancel`.

### Key Files

```
cmd/telefile/main.go          # Entry point, dual-bot setup, route registration
internal/
  bot/instance/instance.go    # MainBot / AdminBot package-level vars
  bot/handlers/              # User command & callback handlers
    upload.go               # Batch upload state machine
    download.go             # File/batch download, password flow
  admin/
    admin_state.go          # FSM state constants
    handlers/
      admin.go              # All admin handlers + inline FSM implementation
  db/                        # PostgreSQL via pgx/v5
    users.go, files.go, gifts.go,
    subscriptions.go, referrals.go, payments.go
  middleware/               # BanCheck, EnsureUser (user auto-creation)
  scheduler/cleanup.go      # Hourly expired-file cleanup, daily storage warnings
  models/                   # Domain types with helper methods
```

## Database Schema

```
users             — TG user, storage quota, referral code, ban status
files             — file_id, unique link, password hash, expiry, download count
batches           — groups multiple files into one link
gifts             — promo codes (storage GB, claims, expiry)
gift_claims       — per-user gift redemption record
subscriptions     — plan-based storage upgrades (active + soft-deleted)
referrals         — referrer → referred mapping, bonus status
pending_payments  — UPI screenshot payments awaiting admin approval
```

Run `setup_db.sql` to create all tables. Run `migrate` for managed migrations.

## License

MIT
