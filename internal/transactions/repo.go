package transactions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	Pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{Pool: pool}
}

func (r *Repo) ListLatest(ctx context.Context, userID string, limit int) ([]TxItem, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	rows, err := r.Pool.Query(ctx, `
SELECT
  'income' AS type,
  id::text,
  client_name AS title,
  amount,
  currency,
  received_on::text AS date,
  created_at::text
FROM incomes
WHERE user_id = $1
  AND deleted_at IS NULL

UNION ALL

SELECT
  'expense' AS type,
  id::text,
  vendor_name AS title,
  amount,
  currency,
  spent_on::text AS date,
  created_at::text
FROM expenses
WHERE user_id = $1
  AND deleted_at IS NULL

ORDER BY created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TxItem, 0, limit)
	for rows.Next() {
		var it TxItem
		if err := rows.Scan(&it.Type, &it.ID, &it.Title, &it.Amount, &it.Currency, &it.Date, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repo) GetSummary(ctx context.Context, userID string) (SummaryResponse, error) {
	var income int64
	var expense int64

	err := r.Pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(amount), 0)::bigint AS income
		FROM incomes
		WHERE user_id = $1
		  AND deleted_at IS NULL
	`, userID).Scan(&income)
	if err != nil {
		return SummaryResponse{}, err
	}

	err = r.Pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(amount), 0)::bigint AS expense
		FROM expenses
		WHERE user_id = $1
		  AND deleted_at IS NULL
	`, userID).Scan(&expense)
	if err != nil {
		return SummaryResponse{}, err
	}

	return SummaryResponse{
		Income:   income,
		Expense:  expense,
		Balance:  income - expense,
		Currency: "INR",
	}, nil
}
