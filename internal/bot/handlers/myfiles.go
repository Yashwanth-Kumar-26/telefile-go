package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// MyFilesState tracks pagination state
type MyFilesState struct {
	Page    int
	Total   int
	PerPage int
}

var myFilesStates = make(map[int64]*MyFilesState)

const perPage = 5

// MyFiles shows user's files
func MyFiles(ctx *bot.Context) error {
	senderID := ctx.SenderID()

	files, total, err := db.GetUserFiles(ctx.Context(), senderID, perPage, 0)
	if err != nil {
		_, _ = ctx.Reply("Failed to fetch files.")
		return nil
	}

	if len(files) == 0 {
		_, _ = ctx.Reply("You haven't uploaded any files yet. Send a file to upload!")
		return nil
	}

	myFilesStates[senderID] = &MyFilesState{
		Page:    1,
		Total:   total,
		PerPage: perPage,
	}

	return showFilesPage(ctx, senderID, files, 1, total)
}

func showFilesPage(ctx *bot.Context, userID int64, files []*models.File, page, total int) error {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[Your Files] (Page %d/%d)\n\n", page, (total+perPage-1)/perPage))

	for i, file := range files {
		num := (page-1)*perPage + i + 1
		expiry := "Never"
		if file.ExpiresAt != nil {
			expiry = timeUntil(*file.ExpiresAt)
		}

		status := "OK"
		if file.IsExpired() {
			status = "Expired"
		}

		protected := ""
		if file.IsPasswordProtected() {
			protected = " [Protected]"
		}

		sb.WriteString(fmt.Sprintf("%d. %s%s\n", num, file.FileName, protected))
		sb.WriteString(fmt.Sprintf("   Size: %s | Downloads: %d | Expiry: %s | Status: %s\n\n",
			utils.FormatBytes(file.FileSize),
			file.DownloadCount,
			expiry,
			status,
		))
	}

	totalPages := (total + perPage - 1) / perPage

	markup := bot.NewInlineKeyboard()

	if page > 1 {
		markup.InlineKeyboard = append(markup.InlineKeyboard,
			[]bot.InlineKeyboardButton{
				bot.InlineButton("Prev", fmt.Sprintf("myfiles_%d", page-1)),
			})
	}

	if page < totalPages {
		markup.InlineKeyboard = append(markup.InlineKeyboard,
			[]bot.InlineKeyboardButton{
				bot.InlineButton("Next", fmt.Sprintf("myfiles_%d", page+1)),
			})
	}

	markup.InlineKeyboard = append(markup.InlineKeyboard,
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Back to Menu", "start"),
		})

	_, _ = ctx.Reply(sb.String(), bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

// HandleMyFilesPage handles pagination callback
func HandleMyFilesPage(ctx *bot.Context, pageStr string) error {
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	senderID := ctx.SenderID()
	state, ok := myFilesStates[senderID]
	if !ok {
		state = &MyFilesState{PerPage: perPage}
		myFilesStates[senderID] = state
	}

	state.Page = page
	files, total, err := db.GetUserFiles(ctx.Context(), senderID, perPage, (page-1)*perPage)
	if err != nil {
		_, _ = ctx.Reply("Failed to fetch files.")
		return nil
	}

	return showFilesPage(ctx, senderID, files, page, total)
}


