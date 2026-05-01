package db

import (
	"context"

	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
)

// CreateReferral creates a new referral
func CreateReferral(ctx context.Context, referrerID, referredID int64, bonusGB int) (*models.Referral, error) {
	ref := &models.Referral{
		ReferrerID: referrerID,
		ReferredID: referredID,
		BonusGB:    bonusGB,
	}

	_, err := Pool.Exec(ctx, `
		INSERT INTO referrals (referrer_id, referred_id, bonus_gb, created_at)
		VALUES ($1, $2, $3, NOW())
	`, ref.ReferrerID, ref.ReferredID, ref.BonusGB)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// GetReferralByReferred retrieves referral by referred user
func GetReferralByReferred(ctx context.Context, referredID int64) (*models.Referral, error) {
	ref := &models.Referral{}
	err := Pool.QueryRow(ctx, `
		SELECT id, referrer_id, referred_id, bonus_gb, is_redeemed, created_at
		FROM referrals WHERE referred_id = $1
	`, referredID).Scan(
		&ref.ID, &ref.ReferrerID, &ref.ReferredID, &ref.BonusGB,
		&ref.IsRedeemed, &ref.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// MarkReferralRedeemed marks a referral as redeemed
func MarkReferralRedeemed(ctx context.Context, referredID int64) error {
	_, err := Pool.Exec(ctx, `
		UPDATE referrals SET is_redeemed = true WHERE referred_id = $1
	`, referredID)
	return err
}

// GetReferralsByUser retrieves all referrals for a user
func GetReferralsByUser(ctx context.Context, userID int64) ([]*models.Referral, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, referrer_id, referred_id, bonus_gb, is_redeemed, created_at
		FROM referrals WHERE referrer_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*models.Referral
	for rows.Next() {
		ref := &models.Referral{}
		err := rows.Scan(
			&ref.ID, &ref.ReferrerID, &ref.ReferredID, &ref.BonusGB,
			&ref.IsRedeemed, &ref.CreatedAt)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// GetReferralCount returns the number of successful referrals
func GetReferralCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM referrals WHERE referrer_id = $1 AND is_redeemed = true
	`, userID).Scan(&count)
	return count, err
}
