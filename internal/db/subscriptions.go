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
	populateDurationDays(sub, features)
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
		populateDurationDays(sub, features)
		subs = append(subs, sub)
	}
	return subs, nil
}

// GetActiveSubscriptionCount returns count of active subscriptions
func GetActiveSubscriptionCount(ctx context.Context) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM subscriptions WHERE is_active = true
	`).Scan(&count)
	return count, err
}

// GetAllPlans returns all plans (active and inactive)
func GetAllPlans(ctx context.Context) ([]*models.Subscription, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, name, storage_gb, price_inr, features, is_active, created_at
		FROM subscriptions ORDER BY price_inr ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.Subscription
	for rows.Next() {
		p := &models.Subscription{}
		var features []byte
		err := rows.Scan(&p.ID, &p.Name, &p.StorageGB, &p.PriceINR, &features, &p.IsActive, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		if features != nil {
			p.Features = features
			populateDurationDays(p, features)
		}
		plans = append(plans, p)
	}
	return plans, nil
}

// CreatePlan creates a new plan and returns its ID
func CreatePlan(ctx context.Context, name, description string, price float64, storageGB int64, durationDays int64) (int, error) {
	// Build features JSON with duration
	features := map[string]interface{}{
		"description":   description,
		"duration_days": durationDays,
	}
	featuresJSON, _ := json.Marshal(features)

	var id int
	err := Pool.QueryRow(ctx, `
		INSERT INTO subscriptions (name, features, storage_gb, price_inr, is_active, created_at)
		VALUES ($1, $2, $3, $4, true, NOW())
		RETURNING id
	`, name, string(featuresJSON), storageGB, price).Scan(&id)
	return id, err
}

// GetSubscriptionByID retrieves a subscription by ID
func GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
	sub := &models.Subscription{}
	var features []byte
	err := Pool.QueryRow(ctx, `
		SELECT id, name, storage_gb, price_inr, features, is_active, created_at
		FROM subscriptions WHERE id = $1 AND is_active = true
	`, id).Scan(
		&sub.ID, &sub.Name, &sub.StorageGB, &sub.PriceINR,
		&features, &sub.IsActive, &sub.CreatedAt)
	if err != nil {
		return nil, err
	}
	sub.Features = json.RawMessage(features)
	populateDurationDays(sub, features)
	return sub, nil
}

// CreatePendingPayment creates a new pending payment record
func CreatePendingPayment(ctx context.Context, userID int64, planID int, planName string, amount float64, screenshotFileID, firstName string) (int, error) {
	var id int
	err := Pool.QueryRow(ctx, `
		INSERT INTO pending_payments (user_id, plan_id, plan_name, amount, screenshot_file_id, first_name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', NOW())
		RETURNING id
	`, userID, planID, planName, amount, screenshotFileID, firstName).Scan(&id)
	return id, err
}

// populateDurationDays extracts duration_days from features JSON into the struct field
func populateDurationDays(sub *models.Subscription, features []byte) {
	if len(features) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(features, &data); err == nil {
			if days, ok := data["duration_days"].(float64); ok {
				sub.DurationDays = int64(days)
			}
		}
	}
}

// DeletePlan soft-deletes a plan by setting is_active to false
func DeletePlan(ctx context.Context, planID int) error {
	_, err := Pool.Exec(ctx, "UPDATE subscriptions SET is_active = false WHERE id = $1", planID)
	return err
}
