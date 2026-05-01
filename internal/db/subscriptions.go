package db

import (
	"context"
	"encoding/json"

	"github.com/Yashwanth-Kumar-26/telefile/internal/models"
)

// GetSubscription retrieves a subscription by name
func GetSubscription(ctx context.Context, name string) (*models.Subscription, error) {
	sub := &models.Subscription{}
	var features []byte
	err := Pool.QueryRow(ctx, `
		SELECT id, name, storage_gb, price_inr, features, is_active, created_at
		FROM subscriptions WHERE name = $1
	`, name).Scan(
		&sub.ID, &sub.Name, &sub.StorageGB, &sub.PriceINR,
		&features, &sub.IsActive, &sub.CreatedAt)
	if err != nil {
		return nil, err
	}
	sub.Features = json.RawMessage(features)
	return sub, nil
}

// GetAllSubscriptions retrieves all active subscriptions
func GetAllSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, name, storage_gb, price_inr, features, is_active, created_at
		FROM subscriptions WHERE is_active = true ORDER BY price_inr ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		sub := &models.Subscription{}
		var features []byte
		err := rows.Scan(
			&sub.ID, &sub.Name, &sub.StorageGB, &sub.PriceINR,
			&features, &sub.IsActive, &sub.CreatedAt)
		if err != nil {
			return nil, err
		}
		sub.Features = json.RawMessage(features)
		subs = append(subs, sub)
	}
	return subs, nil
}
