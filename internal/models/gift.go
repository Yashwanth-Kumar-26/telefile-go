package models

import "time"

// GiftCode represents a gift code in the database
type GiftCode struct {
	Code       string    `json:"code"`
	StorageGB  int       `json:"storage_gb"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	MaxClaims int       `json:"max_claims"`
	ClaimCount int      `json:"claim_count"`
	CreatedBy int64    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the gift code has expired
func (g *GiftCode) IsExpired() bool {
	if g.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*g.ExpiresAt)
}

// IsFullyClaimed checks if the gift code has been claimed all times
func (g *GiftCode) IsFullyClaimed() bool {
	return g.ClaimCount >= g.MaxClaims
}

// CanBeClaimed checks if the gift code can still be claimed
func (g *GiftCode) CanBeClaimed() bool {
	return !g.IsExpired() && !g.IsFullyClaimed()
}