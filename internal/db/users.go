package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
	"github.com/google/uuid"
)

func GetUser(ctx context.Context, userID int64) (*models.User, error) {
	user := &models.User{}
	err := Pool.QueryRow(ctx, `
		SELECT id, username, name, storage_limit_gb, storage_used_bytes,
		       is_banned, referral_code, referred_by, created_at, last_active
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.Name,
		&user.StorageLimitGB, &user.StorageUsedBytes,
		&user.IsBanned, &user.ReferralCode, &user.ReferredBy,
		&user.CreatedAt, &user.LastActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetOrCreateUser(ctx context.Context, userID int64, username string, name string, referredBy *int64, storageGB int64) (*models.User, error) {
	user, err := GetUser(ctx, userID)
	if err == nil {
		return user, nil
	}

	referralCode, _ := generateReferralCode()
	newUser := &models.User{
		ID:               userID,
		Username:         strPtr(username),
		Name:             strPtr(name),
		StorageLimitGB:   storageGB,
		StorageUsedBytes: 0,
		IsBanned:         false,
		ReferralCode:     &referralCode,
		ReferredBy:       referredBy,
		CreatedAt:        time.Now(),
		LastActive:       time.Now(),
	}

	_, err = Pool.Exec(ctx, `
		INSERT INTO users (id, username, name, storage_limit_gb, storage_used_bytes,
		                  is_banned, referral_code, referred_by, created_at, last_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING
	`, newUser.ID, newUser.Username, newUser.Name, newUser.StorageLimitGB, newUser.StorageUsedBytes,
		newUser.IsBanned, newUser.ReferralCode, newUser.ReferredBy, newUser.CreatedAt, newUser.LastActive)
	if err != nil {
		return nil, err
	}

	return GetUser(ctx, userID)
}

func UpdateUserStorage(ctx context.Context, userID int64, bytes int64) error {
	_, err := Pool.Exec(ctx, `
		UPDATE users SET storage_used_bytes = storage_used_bytes + $2, last_active = NOW()
		WHERE id = $1
	`, userID, bytes)
	return err
}

func UpdateUserLimit(ctx context.Context, userID int64, storageGB int64) error {
	_, err := Pool.Exec(ctx, `
		UPDATE users SET storage_limit_gb = storage_limit_gb + $2, last_active = NOW()
		WHERE id = $1
	`, userID, storageGB)
	return err
}

func BanUser(ctx context.Context, userID int64) error {
	_, err := Pool.Exec(ctx, `UPDATE users SET is_banned = true WHERE id = $1`, userID)
	return err
}

func UnbanUser(ctx context.Context, userID int64) error {
	_, err := Pool.Exec(ctx, `UPDATE users SET is_banned = false WHERE id = $1`, userID)
	return err
}

func GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	var total int
	err := Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := Pool.Query(ctx, `
		SELECT id, username, name, storage_limit_gb, storage_used_bytes,
		       is_banned, referral_code, referred_by, created_at, last_active
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Name,
			&user.StorageLimitGB, &user.StorageUsedBytes,
			&user.IsBanned, &user.ReferralCode, &user.ReferredBy,
			&user.CreatedAt, &user.LastActive,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, nil
}

func GetUserCount(ctx context.Context) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func UpdateLastActive(ctx context.Context, userID int64) error {
	_, err := Pool.Exec(ctx, `UPDATE users SET last_active = NOW() WHERE id = $1`, userID)
	return err
}

func GetHighStorageUsers(ctx context.Context, thresholdPercent float64) ([]*models.User, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, username, name, storage_limit_gb, storage_used_bytes,
		       is_banned, referral_code, referred_by, created_at, last_active
		FROM users
		WHERE (storage_used_bytes::float / (storage_limit_gb * 1024 * 1024 * 1024)) > $1
		AND is_banned = false
	`, thresholdPercent/100.0)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Name,
			&user.StorageLimitGB, &user.StorageUsedBytes,
			&user.IsBanned, &user.ReferralCode, &user.ReferredBy,
			&user.CreatedAt, &user.LastActive,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func IsBanned(ctx context.Context, userID int64) (bool, error) {
	user, err := GetUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsBanned, nil
}

func generateReferralCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return uuid.New().String()[:8], nil
	}
	return hex.EncodeToString(b), nil
}

func strPtr(s string) *string { return &s }
