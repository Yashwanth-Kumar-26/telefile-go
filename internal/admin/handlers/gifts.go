package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// GiftCodeState for creating gift codes
type GiftCodeState struct {
	StorageGB int
	MaxClaims int
	ExpiresIn time.Duration
}

// GiftStates stores gift code creation state
var GiftStates = make(map[int64]*GiftCodeState)

// ShowGifts shows gift code management
func ShowGifts(ctx *bot.Context) error {
	senderID := ctx.SenderID()

	gifts, err := db.GetGiftCodesByUser(ctx.Context(), senderID)
	if err != nil {
		return ctx.AnswerCallback("Failed to load gift codes", false)
	}

	text := "[Gift Codes]\n\n"

	if len(gifts) == 0 {
		text += "No gift codes created yet."
	} else {
		for _, gift := range gifts {
			status := "Available"
			if gift.IsExpired() {
				status = "Expired"
			} else if gift.IsFullyClaimed() {
				status = "Fully Claimed"
			}

			expires := "Never"
			if gift.ExpiresAt != nil {
				expires = gift.ExpiresAt.Format("2006-01-02")
			}

			text += fmt.Sprintf("`%s`:\n- Storage: %d GB\n- Claims: %d/%d\n- Expires: %s\n- Status: %s\n\n",
				gift.Code, gift.StorageGB,
				gift.ClaimCount, gift.MaxClaims,
				expires, status)
		}
	}

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("+ Create Gift Code", "admin_gift_create"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Panel", "admin_panel"),
		},
	)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

// StartGiftCreate starts creating a gift code
func StartGiftCreate(ctx *bot.Context) error {
	GiftStates[ctx.SenderID()] = &GiftCodeState{}

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("10 GB", "admin_gift_10gb"),
			bot.InlineButton("50 GB", "admin_gift_50gb"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("100 GB", "admin_gift_100gb"),
			bot.InlineButton("500 GB", "admin_gift_500gb"),
		},
	)

	_, _ = ctx.Reply("Create Gift Code - Select storage amount:", bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

// HandleGiftStorage handles storage selection
func HandleGiftStorage(ctx *bot.Context, gb int) error {
	state, ok := GiftStates[ctx.SenderID()]
	if !ok {
		return ctx.AnswerCallback("No active gift creation", false)
	}

	state.StorageGB = gb

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("1", "admin_gift_claims_1"),
			bot.InlineButton("10", "admin_gift_claims_10"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("100", "admin_gift_claims_100"),
			bot.InlineButton("Unlimited", "admin_gift_claims_unlimited"),
		},
	)

	_, _ = ctx.Reply(fmt.Sprintf("Select max claims for %d GB:", gb), bot.WithReplyMarkup(markup))
	return nil
}

// HandleGiftClaims handles max claims selection
func HandleGiftClaims(ctx *bot.Context, claims int) error {
	state, ok := GiftStates[ctx.SenderID()]
	if !ok {
		return ctx.AnswerCallback("No active gift creation", false)
	}

	state.MaxClaims = claims

	code := generateCode()

	gift, err := db.CreateGiftCode(ctx.Context(), code, state.StorageGB, state.MaxClaims, ctx.SenderID(), nil)
	if err != nil {
		delete(GiftStates, ctx.SenderID())
		return ctx.AnswerCallback("Failed to create gift code", false)
	}

	delete(GiftStates, ctx.SenderID())

	text := fmt.Sprintf("Gift Code Created! Code: `%s`, Storage: %d GB, Max Claims: %d", code, gift.StorageGB, gift.MaxClaims)

	_, _ = ctx.Reply(text, bot.WithParseMode("Markdown"))
	return nil
}

func generateCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}