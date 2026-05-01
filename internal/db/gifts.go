package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
)

// GetGiftCode retrieves a gift code
func GetGiftCode(ctx context.Context, code string) (*models.GiftCode, error) {
	gift := &models.GiftCode{}
	err := Pool.QueryRow(ctx, `
		SELECT code, storage_gb, expires_at, max_claims, claim_count, created_by, created_at
		FROM gift_codes WHERE code = $1
	`, code).Scan(
		&gift.Code, &gift.StorageGB, &gift.ExpiresAt,
		&gift.MaxClaims, &gift.ClaimCount, &gift.CreatedBy, &gift.CreatedAt)
	if err != nil {
		return nil, err
	}
	return gift, nil
}

// CreateGiftCode creates a new gift code
func CreateGiftCode(ctx context.Context, code string, storageGB int, maxClaims int, createdBy int64, expiresAt *time.Time) (*models.GiftCode, error) {
	_, err := Pool.Exec(ctx, `
		INSERT INTO gift_codes (code, storage_gb, expires_at, max_claims, claim_count, created_by, created_at)
		VALUES ($1, $2, $3, $4, 0, $5, NOW())
	`, code, storageGB, expiresAt, maxClaims, createdBy)
	if err != nil {
		return nil, err
	}
	return GetGiftCode(ctx, code)
}

// ClaimGiftCode claims a gift code for a user
func ClaimGiftCode(ctx context.Context, code string, userID int64) (*models.GiftCode, error) {
	gift, err := GetGiftCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if !gift.CanBeClaimed() {
		return nil, fmt.Errorf("gift code is not available")
	}

	var claimed bool
	_ = Pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM gift_claims WHERE user_id = $1 AND code = $2)
	`, userID, code).Scan(&claimed)
	if claimed {
		return nil, fmt.Errorf("gift code already claimed by this user")
	}

	tx, err := Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO gift_claims (user_id, code) VALUES ($1, $2)
	`, userID, code)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE gift_codes SET claim_count = claim_count + 1 WHERE code = $1
	`, code)
	if err != nil {
		return nil, err
	}

	// Update user storage
	_, err = tx.Exec(ctx, `
		UPDATE users SET storage_limit_gb = storage_limit_gb + $2, last_active = NOW()
		WHERE id = $1
	`, userID, int64(gift.StorageGB))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return gift, nil
}

// GetGiftCodesByUser retrieves all gift codes created by a user
func GetGiftCodesByUser(ctx context.Context, userID int64) ([]*models.GiftCode, error) {
	rows, err := Pool.Query(ctx, `
		SELECT code, storage_gb, expires_at, max_claims, claim_count, created_by, created_at
		FROM gift_codes WHERE created_by = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gifts []*models.GiftCode
	for rows.Next() {
		gift := &models.GiftCode{}
		err := rows.Scan(
			&gift.Code, &gift.StorageGB, &gift.ExpiresAt,
			&gift.MaxClaims, &gift.ClaimCount, &gift.CreatedBy, &gift.CreatedAt)
		if err != nil {
			return nil, err
		}
		gifts = append(gifts, gift)
	}
	return gifts, nil
}

// GetAllGiftCodes retrieves all gift codes (admin)
func GetAllGiftCodes(ctx context.Context) ([]*models.GiftCode, error) {
	rows, err := Pool.Query(ctx, `
		SELECT code, storage_gb, expires_at, max_claims, claim_count, created_by, created_at
		FROM gift_codes ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gifts []*models.GiftCode
	for rows.Next() {
		gift := &models.GiftCode{}
		err := rows.Scan(
			&gift.Code, &gift.StorageGB, &gift.ExpiresAt,
			&gift.MaxClaims, &gift.ClaimCount, &gift.CreatedBy, &gift.CreatedAt)
		if err != nil {
			return nil, err
		}
		gifts = append(gifts, gift)
	}
	return gifts, nil
}
