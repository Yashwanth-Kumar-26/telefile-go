package models

import "time"

// Referral represents a referral in the database
type Referral struct {
	ID         int       `json:"id"`
	ReferrerID int64    `json:"referrer_id"`
	ReferredID int64    `json:"referred_id"`
	BonusGB   int      `json:"bonus_gb"`
	IsRedeemed bool     `json:"is_redeemed"`
	CreatedAt time.Time `json:"created_at"`
}