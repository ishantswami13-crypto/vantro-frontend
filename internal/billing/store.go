package billing

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	DB *sql.DB
}

func (s *Store) IsActive(ctx context.Context, phone string) (bool, error) {
	const q = `
        SELECT status, current_period_end
        FROM subscriptions
        WHERE user_phone = $1;
    `
	var status string
	var end sql.NullTime
	err := s.DB.QueryRowContext(ctx, q, phone).Scan(&status, &end)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if status != "active" || !end.Valid {
		return false, nil
	}
	return time.Now().Before(end.Time), nil
}

func (s *Store) ActivateFor30Days(ctx context.Context, phone string, paymentLinkID string) error {
	const q = `
        INSERT INTO subscriptions (user_phone, status, current_period_end, razorpay_payment_link_id, updated_at)
        VALUES ($1, 'active', NOW() + INTERVAL '30 days', $2, NOW())
        ON CONFLICT (user_phone) DO UPDATE SET
            status = 'active',
            current_period_end = NOW() + INTERVAL '30 days',
            razorpay_payment_link_id = EXCLUDED.razorpay_payment_link_id,
            updated_at = NOW();
    `
	_, err := s.DB.ExecContext(ctx, q, phone, paymentLinkID)
	return err
}
