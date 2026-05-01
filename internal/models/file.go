package models

import "time"

// File represents a stored file in the database
type File struct {
	ID             string     `json:"id"`
	UserID        int64      `json:"user_id"`
	FileName      string     `json:"file_name"`
	FileSize      int64      `json:"file_size"`
	FileType      *string    `json:"file_type"`
	TelegramFileID string   `json:"telegram_file_id"`
	UniqueLink   string     `json:"unique_link"`
	Password     *string    `json:"password,omitempty"`
	DownloadLimit int        `json:"download_limit"`
	DownloadCount int       `json:"download_count"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	BatchID      *int       `json:"batch_id,omitempty"`
}

// IsExpired checks if the file has expired
func (f *File) IsExpired() bool {
	if f.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*f.ExpiresAt)
}

// IsPasswordProtected checks if the file has password protection
func (f *File) IsPasswordProtected() bool {
	return f.Password != nil && *f.Password != ""
}

// CanDownload checks if the file can still be downloaded
func (f *File) CanDownload() bool {
	if f.IsExpired() {
		return false
	}
	if f.DownloadLimit == -1 {
		return true
	}
	return f.DownloadCount < f.DownloadLimit
}

// RemainingDownloads returns how many downloads are left
func (f *File) RemainingDownloads() int {
	if f.DownloadLimit == -1 {
		return -1 // unlimited
	}
	remaining := f.DownloadLimit - f.DownloadCount
	if remaining < 0 {
		return 0
	}
	return remaining
}