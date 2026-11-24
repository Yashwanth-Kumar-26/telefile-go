package db

import (
	"context"
	"fmt"
	"log"

	"github.com/Yashwanth-Kumar-26/telefile/internal/bot/instance"
	"github.com/Yashwanth-Kumar-26/teleriha/pkg/bot"
)

type PendingPayment struct {
	ID               int     `json:"id"`
	UserID           int64   `json:"user_id"`
	PlanID           int     `json:"plan_id"`
	PlanName         string  `json:"plan_name"`
	FirstName        string  `json:"first_name"`
	Amount           float64 `json:"amount"`
	ScreenshotFileID string  `json:"screenshot_file_id"`
	CreatedAt        string  `json:"created_at"`
}

func GetPendingPaymentCount(ctx context.Context) (int, error) {
	var count int
	err := Pool.QueryRow(ctx, "SELECT COUNT(*) FROM pending_payments").Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, err
}

func GetPendingPayments(ctx context.Context) ([]*PendingPayment, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, user_id, plan_id, plan_name, COALESCE(first_name, 'User'), COALESCE(amount, 0),
		       COALESCE(screenshot_file_id, ''), COALESCE(created_at::text, '')
		FROM pending_payments ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var payments []*PendingPayment
	for rows.Next() {
		p := &PendingPayment{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.PlanID, &p.PlanName, &p.FirstName, &p.Amount, &p.ScreenshotFileID, &p.CreatedAt); err != nil {
			continue
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func GetPayment(ctx context.Context, paymentID int) (*PendingPayment, error) {
	p := &PendingPayment{}
	err := Pool.QueryRow(ctx, `
		SELECT id, user_id, plan_id, plan_name, COALESCE(first_name, 'User'), COALESCE(amount, 0),
		       COALESCE(screenshot_file_id, ''), COALESCE(created_at::text, '')
		FROM pending_payments WHERE id = $1
	`, paymentID).Scan(&p.ID, &p.UserID, &p.PlanName, &p.FirstName, &p.Amount, &p.ScreenshotFileID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ApprovePayment(ctx context.Context, paymentID int) error {
	payment, err := GetPayment(ctx, paymentID)
	if err != nil {
		log.Printf("[PAYMENT] Error getting payment: %v", err)
		return err
	}

	plan, err := GetSubscriptionByID(ctx, payment.PlanID)
	if err != nil || plan == nil {
		log.Printf("[PAYMENT] Plan not found: plan_id=%d", payment.PlanID)
		return err
	}

	// Update user storage (add plan storage to their limit)
	_, err = Pool.Exec(ctx, `
		UPDATE users SET storage_limit_gb = storage_limit_gb + $2
		WHERE id = $1
	`, payment.UserID, int64(plan.StorageGB))
	if err != nil {
		log.Printf("[PAYMENT] Error updating user storage: %v", err)
		return err
	}

	// Notify user of approval
	if instance.MainBot != nil {
		msg := fmt.Sprintf("Payment Approved!\n\nPlan: %s\nStorage: %s\n\nYour plan has been activated.", payment.PlanName, plan.StorageLabel())
		_, _ = instance.MainBot.SendMessage(payment.UserID, msg, bot.WithParseMode("Markdown"))
	}

	// Delete pending payment
	_, err = Pool.Exec(ctx, "DELETE FROM pending_payments WHERE id = $1", paymentID)
	if err != nil {
		log.Printf("[PAYMENT] Error deleting payment: %v", err)
	}
	return err
}

func RejectPayment(ctx context.Context, paymentID int) error {
	payment, err := GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}
	_, err = Pool.Exec(ctx, "DELETE FROM pending_payments WHERE id = $1", paymentID)
	if err != nil {
		return err
	}
	if instance.MainBot != nil {
		msg := fmt.Sprintf("Payment Rejected.\n\nPlan: %s\nAmount: ₹%.0f\n\nPlease contact support or re-submit with correct payment.", payment.PlanName, payment.Amount)
		_, _ = instance.MainBot.SendMessage(payment.UserID, msg, bot.WithParseMode("Markdown"))
	}
	return nil
}