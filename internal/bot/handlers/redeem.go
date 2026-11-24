package handlers

import (
	"context"
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// RedeemState tracks redeem conversation state
var RedeemStates = make(map[int64]bool)

// StartRedeem starts the redeem flow
func StartRedeem(ctx *bot.Context) error {
	RedeemStates[ctx.SenderID()] = true
	_, _ = ctx.Reply("Redeem Gift Code - Please enter your gift code:")
	return nil
}

// HandleRedeemCode handles gift code input
func HandleRedeemCode(ctx *bot.Context) error {
	senderID := ctx.SenderID()

	if !RedeemStates[senderID] {
		return nil
	}

	code := ctx.Text()
	if code == "" {
		return nil
	}

	if len(code) < 4 {
		_, _ = ctx.Reply("Invalid code format. Please try again:")
		return nil
	}

	delete(RedeemStates, senderID)

	gift, err := db.ClaimGiftCode(context.Background(), code, senderID)
	if err != nil {
		_, _ = ctx.Reply(fmt.Sprintf("Failed to claim: %v", err))
		return nil
	}
	// ClaimGiftCode already adds storage to user limit (gifts.go:78-84)

	_, _ = ctx.Reply(fmt.Sprintf("Gift Code Redeemed! + %d GB of storage added to your account!", gift.StorageGB), bot.WithParseMode("Markdown"))
	return nil
}

// CancelRedeem cancels redeem flow
func CancelRedeem(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if RedeemStates[senderID] {
		delete(RedeemStates, senderID)
		_, _ = ctx.Reply("Redeem cancelled.")
		return nil
	}
	return nil
}