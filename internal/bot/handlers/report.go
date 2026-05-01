package handlers

import (
	"fmt"

	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// ReportState tracks report conversation state
var ReportStates = make(map[int64]bool)

// StartReport starts the report flow
func StartReport(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	ReportStates[senderID] = true
	_, _ = ctx.Reply("📋 *Report an Issue*\n\nPlease describe your issue in a few words?")
	return nil
}

// HandleReport handles the issue description
func HandleReport(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if !ReportStates[senderID] {
		return nil
	}

	issue := ctx.Text()
	if issue == "" {
		return nil
	}

	delete(ReportStates, senderID)

	text := fmt.Sprintf(`📋 *Issue Reported*

Thank you! Our team will review your issue.
Issue: %s

We will get back to you soon.`, issue)

	_, err := ctx.Reply(text, bot.WithParseMode("Markdown"))
	return err
}

// CancelReport cancels report flow
func CancelReport(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if ReportStates[senderID] {
		delete(ReportStates, senderID)
		_, _ = ctx.Reply("✅ Report cancelled.")
		return nil
	}
	return nil
}