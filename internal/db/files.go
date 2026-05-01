package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
	"github.com/google/uuid"
)

// SaveFile saves a new file record
func SaveFile(ctx context.Context, file *models.File) error {
	if file.ID == "" {
		file.ID = uuid.New().String()
	}
	if file.UniqueLink == "" {
		file.UniqueLink = uuid.New().String()[:12]
	}
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now()
	}

	_, err := Pool.Exec(ctx, `
		INSERT INTO files (id, user_id, file_name, file_size, file_type,
		                  telegram_file_id, unique_link, password,
		                  download_limit, download_count, expires_at, created_at, batch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, file.ID, file.UserID, file.FileName, file.FileSize, file.FileType,
		file.TelegramFileID, file.UniqueLink, file.Password,
		file.DownloadLimit, file.DownloadCount, file.ExpiresAt, file.CreatedAt, file.BatchID)
	return err
}

// GetFile retrieves a file by ID
func GetFile(ctx context.Context, id string) (*models.File, error) {
	file := &models.File{}
	err := Pool.QueryRow(ctx, `
		SELECT id, user_id, file_name, file_size, file_type,
		       telegram_file_id, unique_link, password,
		       download_limit, download_count, expires_at, created_at
		FROM files WHERE id = $1
	`, id).Scan(
		&file.ID, &file.UserID, &file.FileName, &file.FileSize, &file.FileType,
		&file.TelegramFileID, &file.UniqueLink, &file.Password,
		&file.DownloadLimit, &file.DownloadCount, &file.ExpiresAt, &file.CreatedAt)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetFileByLink retrieves a file by unique link
func GetFileByLink(ctx context.Context, link string) (*models.File, error) {
	file := &models.File{}
	err := Pool.QueryRow(ctx, `
		SELECT id, user_id, file_name, file_size, file_type,
		       telegram_file_id, unique_link, password,
		       download_limit, download_count, expires_at, created_at
		FROM files WHERE unique_link = $1
	`, link).Scan(
		&file.ID, &file.UserID, &file.FileName, &file.FileSize, &file.FileType,
		&file.TelegramFileID, &file.UniqueLink, &file.Password,
		&file.DownloadLimit, &file.DownloadCount, &file.ExpiresAt, &file.CreatedAt)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetUserFiles retrieves all files for a user
func GetUserFiles(ctx context.Context, userID int64, limit, offset int) ([]*models.File, int, error) {
	var total int
	err := Pool.QueryRow(ctx, "SELECT COUNT(*) FROM files WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := Pool.Query(ctx, `
		SELECT id, user_id, file_name, file_size, file_type,
		       telegram_file_id, unique_link, password,
		       download_limit, download_count, expires_at, created_at
		FROM files WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		file := &models.File{}
		err := rows.Scan(
			&file.ID, &file.UserID, &file.FileName, &file.FileSize, &file.FileType,
			&file.TelegramFileID, &file.UniqueLink, &file.Password,
			&file.DownloadLimit, &file.DownloadCount, &file.ExpiresAt, &file.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		files = append(files, file)
	}
	return files, total, nil
}

// IncrementDownload increments the download count
func IncrementDownload(ctx context.Context, fileID string) error {
	_, err := Pool.Exec(ctx, `
		UPDATE files SET download_count = download_count + 1 WHERE id = $1
	`, fileID)
	return err
}

// DeleteFile deletes a file
func DeleteFile(ctx context.Context, fileID string) error {
	_, err := Pool.Exec(ctx, "DELETE FROM files WHERE id = $1", fileID)
	return err
}

// DeleteExpiredFiles deletes all expired files
func DeleteExpiredFiles(ctx context.Context) (int64, error) {
	tag, err := Pool.Exec(ctx, `
		DELETE FROM files WHERE expires_at IS NOT NULL AND expires_at < NOW()
	`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// GetTotalStorageUsed returns total storage used across all files
func GetTotalStorageUsed(ctx context.Context) (int64, error) {
	var total int64
	err := Pool.QueryRow(ctx, "SELECT COALESCE(SUM(file_size), 0) FROM files").Scan(&total)
	return total, err
}

// GetFileCount returns total file count
func GetFileCount(ctx context.Context) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, "SELECT COUNT(*) FROM files").Scan(&count)
	return count, err
}

// GetUserFileCount returns file count for a specific user
func GetUserFileCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, "SELECT COUNT(*) FROM files WHERE user_id = $1", userID).Scan(&count)
	return count, err
}

// CreateBatch creates a new batch record
func CreateBatch(ctx context.Context, userID int64) (int, error) {
	var id int
	err := Pool.QueryRow(ctx, "INSERT INTO batches (user_id) VALUES ($1) RETURNING id", userID).Scan(&id)
	return id, err
}

// SetBatchLink sets the unique link for a batch
func SetBatchLink(ctx context.Context, batchID int, linkID string) error {
	_, err := Pool.Exec(ctx, "UPDATE batches SET unique_link_id = $1 WHERE id = $2", linkID, batchID)
	return err
}

// GetBatchFiles returns all files in a batch
func GetBatchFiles(ctx context.Context, linkID string) ([]*models.File, error) {
	rows, err := Pool.Query(ctx, `
		SELECT f.id, f.user_id, f.file_name, f.file_size, f.telegram_file_id, f.unique_link, f.password, f.download_limit, f.expires_at, f.created_at
		FROM files f
		JOIN batches b ON f.batch_id = b.id
		WHERE b.unique_link_id = $1
	`, linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		f := &models.File{}
		err := rows.Scan(&f.ID, &f.UserID, &f.FileName, &f.FileSize, &f.TelegramFileID, &f.UniqueLink, &f.Password, &f.DownloadLimit, &f.ExpiresAt, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// FileBelongsToUser checks if a file belongs to a user
func FileBelongsToUser(ctx context.Context, fileID string, userID int64) (bool, error) {
	var exists bool
	err := Pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM files WHERE id = $1 AND user_id = $2)
	`, fileID, userID).Scan(&exists)
	return exists, err
}

// ErrFileNotFound is returned when a file is not found
type ErrFileNotFound struct{ ID string }

func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("file not found: %s", e.ID)
}
