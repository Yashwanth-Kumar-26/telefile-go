package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/telefile/internal/config"
	"github.com/Yashwanth-Kumar-26/telefile/internal/db"
	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
	"github.com/google/uuid"
)

type FileInfo struct {
	FileID   string
	FileName string
	FileSize int64
	FileType string
}

type UploadState struct {
	Files         []FileInfo
	BatchMode     bool
	Password      string
	ExpiresIn     time.Duration
	MaxDownloads  int // -1 = unlimited, 0 = not set yet
	ControlMsgID  int64
	AskPhase      string // "expiry" | "password" | "downloads"
}

var PendingUploads = make(map[int64]*UploadState)

func HandleUpload(ctx *bot.Context, cfg *config.Config) error {
	senderID := ctx.SenderID()

	banned, _ := db.IsBanned(context.Background(), senderID)
	if banned {
		_, _ = ctx.Reply("Your account has been suspended.")
		return nil
	}

	user, err := db.GetUser(context.Background(), senderID)
	if err != nil {
		_, _ = ctx.Reply("Please use /start first.")
		return nil
	}

	state, ok := PendingUploads[senderID]
	if !ok {
		state = &UploadState{Files: []FileInfo{}, AskPhase: "expiry"}
		PendingUploads[senderID] = state
	}

	var info *FileInfo
	if ctx.Message != nil {
		if ctx.Message.Document != nil {
			info = &FileInfo{
				FileID:   ctx.Message.Document.FileID,
				FileName: ctx.Message.Document.FileName,
				FileSize: ctx.Message.Document.FileSize,
				FileType: "document",
			}
		} else if ctx.Message.Photo != nil {
			photo := ctx.Message.Photo
			info = &FileInfo{
				FileID:   photo[len(photo)-1].FileID,
				FileName: fmt.Sprintf("photo_%d.jpg", time.Now().Unix()),
				FileSize: photo[len(photo)-1].FileSize,
				FileType: "image",
			}
		}
	}

	if info == nil {
		return nil
	}

	maxSize := cfg.MaxFileSizeMB * 1024 * 1024
	if info.FileSize > maxSize {
		_, _ = ctx.Reply(fmt.Sprintf("File too large. Maximum size is %d MB.", cfg.MaxFileSizeMB))
		return nil
	}

	totalPendingSize := int64(0)
	for _, f := range state.Files {
		totalPendingSize += f.FileSize
	}

	if !user.HasAvailableStorage(totalPendingSize + info.FileSize) {
		_, _ = ctx.Reply("Storage limit exceeded.")
		return nil
	}

	state.Files = append(state.Files, *info)

	var fileList strings.Builder
	for i, f := range state.Files {
		fileList.WriteString(fmt.Sprintf("%d. `%s` (%s)\n", i+1, f.FileName, utils.FormatBytes(f.FileSize)))
	}

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton(fmt.Sprintf("Done (%d files) ✅", len(state.Files)), "upload_batch_done"),
			bot.InlineButton("Cancel ❌", "upload_cancel"),
		},
	)

	msg := fmt.Sprintf("📦 *Batch Upload in Progress*\n\n*Files Collected:*\n%s\nTotal: %d files.\n\nSend more files or click **Done**.", 
		fileList.String(), len(state.Files))

	if state.ControlMsgID != 0 {
		_, err := ctx.Bot.EditMessageText(ctx.ChatID(), state.ControlMsgID, msg, 
			bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
		if err == nil {
			return nil
		}
	}

	res, _ := ctx.Reply(msg, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	if res != nil {
		state.ControlMsgID = res.MessageID
	}
	state.BatchMode = true
	return nil
}

func StartBatch(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	state, ok := PendingUploads[senderID]
	if !ok {
		state = &UploadState{Files: []FileInfo{}}
		PendingUploads[senderID] = state
	}

	state.BatchMode = true
	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton(fmt.Sprintf("Done (%d files) ✅", len(state.Files)), "upload_batch_done"),
			bot.InlineButton("Cancel ❌", "upload_cancel"),
		},
	)
	
	msg := fmt.Sprintf("📦 *Batch Mode Activated!*\n\nCurrently have %d files.\n\nPlease send all the files you want to include in this batch, then click **Done**.", len(state.Files))
	
	var res *bot.Message
	if ctx.CallbackQuery != nil {
		_ = ctx.AnswerCallback("", false)
		res, _ = ctx.EditMessageText(msg, bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
	} else {
		res, _ = ctx.Reply(msg, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	}

	if res != nil {
		state.ControlMsgID = res.MessageID
	}
	return nil
}

func UploadPasswordYes(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	_, _ = ctx.Reply("Please enter a password for your upload:")
	return nil
}

func UploadPasswordNo(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("", false)
	senderID := ctx.SenderID()
	state, ok := PendingUploads[senderID]
	if !ok {
		log.Printf("[ERROR] UploadPasswordNo: No state found for user %d", senderID)
		return nil
	}
	state.Password = ""
	return showExpiryOptions(ctx)
}

func HandlePasswordInput(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	state, ok := PendingUploads[senderID]
	if !ok || len(state.Files) == 0 {
		return nil
	}

	if state.Password == "" && ctx.Text() != "" {
		state.Password = ctx.Text()
		return showExpiryOptions(ctx)
	}
	return nil
}

func showExpiryOptions(ctx *bot.Context) error {
	senderID := ctx.SenderID()
	state, ok := PendingUploads[senderID]
	if !ok {
		log.Printf("[ERROR] showExpiryOptions: No state found for user %d", senderID)
		return nil
	}

	markup := bot.NewInlineKeyboard(
		[]bot.InlineKeyboardButton{
			bot.InlineButton("1 Hour", "upload_exp_1h"),
			bot.InlineButton("24 Hours", "upload_exp_24h"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("7 Days", "upload_exp_7d"),
			bot.InlineButton("30 Days", "upload_exp_30d"),
		},
		[]bot.InlineKeyboardButton{
			bot.InlineButton("Never", "upload_exp_never"),
		},
	)
	
	msg := fmt.Sprintf("⏳ *Upload Ready (%d files)*\n\nWhen should your upload expire?", len(state.Files))
	
	if state.ControlMsgID != 0 {
		_, err := ctx.Bot.EditMessageText(ctx.ChatID(), state.ControlMsgID, msg, 
			bot.WithNewParseMode("Markdown"), bot.WithNewReplyMarkup(markup))
		if err == nil {
			return nil
		}
	}
	
	_, _ = ctx.Reply(msg, bot.WithParseMode("Markdown"), bot.WithReplyMarkup(markup))
	return nil
}

func UploadExpiryHandler(ctx *bot.Context, cfg *config.Config, hours int) error {
	_ = ctx.AnswerCallback("", false)
	senderID := ctx.SenderID()
	state, ok := PendingUploads[senderID]
	if !ok || len(state.Files) == 0 {
		log.Printf("[ERROR] UploadExpiryHandler: No state found for user %d", senderID)
		return nil
	}

	var expiresAt *time.Time
	if hours > 0 {
		t := time.Now().Add(time.Duration(hours) * time.Hour)
		expiresAt = &t
	}

	var batchID *int
	batchLinkID := ""
	if len(state.Files) > 1 {
		id, err := db.CreateBatch(context.Background(), senderID)
		if err == nil {
			batchID = &id
			batchLinkID = uuid.New().String()[:12]
			_ = db.SetBatchLink(context.Background(), id, batchLinkID)
		}
	}

	var savedFiles []*models.File
	totalSize := int64(0)

	for _, fInfo := range state.Files {
		uniqueLink := uuid.New().String()[:12]
		file := &models.File{
			ID:             uuid.New().String(),
			UserID:         senderID,
			FileName:       fInfo.FileName,
			FileSize:       fInfo.FileSize,
			TelegramFileID: fInfo.FileID,
			UniqueLink:     uniqueLink,
			DownloadLimit:  -1,
			ExpiresAt:      expiresAt,
			CreatedAt:      time.Now(),
			BatchID:        batchID,
		}
		if fInfo.FileType != "" {
			file.FileType = &fInfo.FileType
		}
		if state.Password != "" {
			file.Password = &state.Password
		}

		_ = db.SaveFile(context.Background(), file)
		_ = db.UpdateUserStorage(context.Background(), senderID, fInfo.FileSize)
		totalSize += fInfo.FileSize
		savedFiles = append(savedFiles, file)
	}

	delete(PendingUploads, senderID)

	var shareLink string
	var resultText string
	if batchID != nil {
		shareLink = fmt.Sprintf("https://t.me/%s?start=batch_%s", cfg.BotUsername, batchLinkID)
		resultText = fmt.Sprintf("✅ *Batch Upload Complete!*\n\nFiles: %d\nTotal Size: %s\nLink: `%s`", len(state.Files), utils.FormatBytes(totalSize), shareLink)
	} else {
		file := savedFiles[0]
		shareLink = fmt.Sprintf("https://t.me/%s?start=file_%s", cfg.BotUsername, file.UniqueLink)
		resultText = fmt.Sprintf("✅ *Upload Complete!*\n\nFile: `%s`\nSize: %s\nLink: `%s`", file.FileName, utils.FormatBytes(file.FileSize), shareLink)
	}

	if state.ControlMsgID != 0 {
		_, _ = ctx.Bot.EditMessageText(ctx.ChatID(), state.ControlMsgID, resultText, bot.WithNewParseMode("Markdown"))
	} else {
		_, _ = ctx.Reply(resultText, bot.WithParseMode("Markdown"))
	}

	for _, f := range savedFiles {
		_, _ = ctx.Bot.SendDocument(ctx.ChatID(), f.TelegramFileID, fmt.Sprintf("File: %s (%s)", f.FileName, utils.FormatBytes(f.FileSize)))
	}

	fileCount, _ := db.GetUserFileCount(context.Background(), senderID)
	if fileCount <= len(state.Files) { 
		user, _ := db.GetUser(context.Background(), senderID)
		if user != nil && user.ReferredBy != nil {
			bonusGB := int64(5)
			_ = db.UpdateUserLimit(context.Background(), *user.ReferredBy, bonusGB)
			_ = db.UpdateUserLimit(context.Background(), senderID, bonusGB)
			if instance.MainBot != nil {
				_, _ = instance.MainBot.SendMessage(*user.ReferredBy, fmt.Sprintf("🎉 Referral reward! You earned %d GB storage.", bonusGB))
			}
		}
	}

	return nil
}

func CancelUpload(ctx *bot.Context) error {
	_ = ctx.AnswerCallback("Upload Cancelled", true)
	delete(PendingUploads, ctx.SenderID())
	_, _ = ctx.Reply("Upload cancelled.")
	return nil
}
