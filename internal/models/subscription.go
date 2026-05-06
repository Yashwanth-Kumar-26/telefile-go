package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Subscription represents a subscription plan
type Subscription struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	StorageGB    int             `json:"storage_gb"`
	PriceINR     float64         `json:"price_inr"`
	Features     json.RawMessage `json:"features"`
	IsActive     bool            `json:"is_active"`
	DurationDays int64           `json:"duration_days"`
	CreatedAt    time.Time       `json:"created_at"`
}

// MaxFileSize returns the max file size in MB from features
func (s *Subscription) MaxFileSize() int64 {
	if s.Features == nil {
		return 2048
	}
	var features map[string]interface{}
	if err := json.Unmarshal(s.Features, &features); err != nil {
		return 2048
	}
	if maxSize, ok := features["max_file_size"].(float64); ok {
		return int64(maxSize)
	}
	return 2048
}

// HasFeature checks if a feature is enabled
func (s *Subscription) HasFeature(feature string) bool {
	if s.Features == nil {
		return false
	}
	var features map[string]interface{}
	if err := json.Unmarshal(s.Features, &features); err != nil {
		return false
	}
	if val, ok := features[feature].(bool); ok {
		return val
	}
	return false
}

// StorageLabel returns a human-readable storage label
func (s *Subscription) StorageLabel() string {
	if s.StorageGB == -1 {
		return "Unlimited"
	}
	if s.StorageGB >= 1024 {
		return fmt.Sprintf("%d TB", s.StorageGB/1024)
	}
	return fmt.Sprintf("%d GB", s.StorageGB)
}