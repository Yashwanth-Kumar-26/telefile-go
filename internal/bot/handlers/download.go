package handlers

import (
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

// DownloadStates tracks download conversation state
var DownloadStates = make(map[int64]string)

// HandleFileDownload handles file download via start param
func HandleFileDownload(ctx *bot.Context, uniqueLink string) error {
	file, err := db.GetFileByLink(ctx.Context(), uniqueLink)
	if err != nil {
		_, _ = ctx.Reply("File not found or link is invalid.")
		return nil
	}

	if file.IsExpired() {
		_, _ = ctx.Reply("This file has expired.")
		return nil
	}

	if !file.CanDownload() {
		_, _ = ctx.Reply("Download limit reached for this file.")
		return nil
	}

	if file.IsPasswordProtected() {
		DownloadStates[ctx.SenderID()] = file.ID
		_, _ = ctx.Reply("This file is password protected. Please enter the password:")
		return nil
	}

	return sendFile(ctx, file)
}

// HandleBatchDownload handles batch download via start param
func HandleBatchDownload(ctx *bot.Context, linkID string) error {
	files, err := db.GetBatchFiles(ctx.Context(), linkID)
	if err != nil || len(files) == 0 {
		_, _ = ctx.Reply("Batch not found or link is invalid.")
		return nil
	}

	first := files[0]
	if first.IsPasswordProtected() {
		DownloadStates[ctx.SenderID()] = "batch_" + linkID
		_, _ = ctx.Reply("This batch is password protected. Please enter the password:")
		return nil
	}

	for _, f := range files {
		_ = sendFile(ctx, f)
	}
	return nil
}

func sendFile(ctx *bot.Context, file *models.File) error {
	caption := fmt.Sprintf("File: %s (%s)", file.FileName, utils.FormatBytes(file.FileSize))
	
	var err error
	fileType := ""
	if file.FileType != nil {
		fileType = *file.FileType
	}

	// Route based on file type
	switch fileType {
	case "image":
		_, err = ctx.Bot.SendPhoto(ctx.ChatID(), file.TelegramFileID, caption)
	case "video":
		_, err = ctx.Bot.SendVideo(ctx.ChatID(), file.TelegramFileID, caption)
	default:
		_, err = ctx.Bot.SendDocument(ctx.ChatID(), file.TelegramFileID, caption)
	}

	if err != nil {
		fmt.Printf("[ERROR] Failed to send file %s (Type: %s): %v\n", file.FileName, fileType, err)
		_, _ = ctx.Reply(fmt.Sprintf("Failed to send file: %v", err))
		return nil
	}
	
	_ = db.IncrementDownload(ctx.Context(), file.ID)
	return nil
}

// HandleDownloadPassword handles password input for downloads
func HandleDownloadPassword(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	state, ok := DownloadStates[senderID]
	if !ok {
		return nil
	}

	password := ctx.Text()

	if len(state) > 6 && state[:6] == "batch_" {
		linkID := state[6:]
		files, err := db.GetBatchFiles(ctx.Context(), linkID)
		if err != nil || len(files) == 0 {
			delete(DownloadStates, senderID)
			return nil
		}
		if files[0].Password == nil || *files[0].Password != password {
			_, _ = ctx.Reply("Incorrect password. Please try again:")
			return nil
		}
		delete(DownloadStates, senderID)
		for _, f := range files {
			_ = sendFile(ctx, f)
		}
		return nil
	}

	file, err := db.GetFile(ctx.Context(), state)
	if err != nil {
		delete(DownloadStates, senderID)
		_, _ = ctx.Reply("File not found.")
		return nil
	}

	if file.Password == nil || *file.Password != password {
		_, _ = ctx.Reply("Incorrect password. Please try again:")
		return nil
	}

	delete(DownloadStates, senderID)
	return sendFile(ctx, file)
}

func CancelDownload(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	if _, ok := DownloadStates[senderID]; ok {
		delete(DownloadStates, senderID)
		_, _ = ctx.Reply("Download cancelled.")
		return nil
	}
	return nil
}
