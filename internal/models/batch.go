package models

import "time"

type Batch struct {
	ID           int       `json:"id"`
	UserID       int64     `json:"user_id"`
	UniqueLinkID string    `json:"unique_link_id"`
	CreatedAt    time.Time `json:"created_at"`
}
