package models

import "time"

// User represents a Telegram user in the database
type User struct {
	ID              int64      `json:"id"`
	Username        *string   `json:"username"`
	Name            *string   `json:"name"`
	StorageLimitGB  int64     `json:"storage_limit_gb"`
	StorageUsedBytes int64   `json:"storage_used_bytes"`
	IsBanned        bool      `json:"is_banned"`
	ReferralCode    *string   `json:"referral_code"`
	ReferredBy      *int64    `json:"referred_by"`
	CreatedAt       time.Time `json:"created_at"`
	LastActive       time.Time `json:"last_active"`
}

// HasAvailableStorage checks if user has storage space for a file
func (u *User) HasAvailableStorage(fileSize int64) bool {
	if u.StorageLimitGB == -1 { // Unlimited plan
		return true
	}
	maxBytes := u.StorageLimitGB * 1024 * 1024 * 1024
	return (u.StorageUsedBytes + fileSize) <= maxBytes
}

// StoragePercentUsed returns the percentage of storage used
func (u *User) StoragePercentUsed() float64 {
	if u.StorageLimitGB == -1 {
		return 0
	}
	maxBytes := float64(u.StorageLimitGB * 1024 * 1024 * 1024)
	return float64(u.StorageUsedBytes) / maxBytes * 100
}

// StorageUsedGB returns used storage in GB
func (u *User) StorageUsedGB() float64 {
	return float64(u.StorageUsedBytes) / (1024 * 1024 * 1024)
}

// StorageLimitGBF returns limit in GB (float)
func (u *User) StorageLimitGBF() float64 {
	if u.StorageLimitGB == -1 {
		return -1
	}
	return float64(u.StorageLimitGB)
}

// DisplayName returns a readable name for the user
func (u *User) DisplayName() string {
	if u.Name != nil && *u.Name != "" {
		return *u.Name
	}
	if u.Username != nil && *u.Username != "" {
		return "@" + *u.Username
	}
	return "User"
}