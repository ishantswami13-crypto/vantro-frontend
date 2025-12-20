package expense

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

func (r *Repository) InsertExpense(ctx context.Context, exp *Expense) (string, error) {
	var id string
	err := r.Pool.QueryRow(
		ctx,
		`INSERT INTO expenses (user_id, vendor_name, amount, currency, spent_on, note)
         VALUES ($1, $2, $3, COALESCE($4,'INR'), $5, $6)
         RETURNING id`,
		exp.UserID,
		exp.VendorName,
		exp.Amount,
		exp.Currency,
		exp.SpentOn,
		exp.Note,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repository) ListExpensesByUser(ctx context.Context, userID string) ([]Expense, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT id, user_id, vendor_name, amount, currency, spent_on, note, created_at
		FROM expenses
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Expense, 0)
	for rows.Next() {
		var e Expense
		if err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.VendorName,
			&e.Amount,
			&e.Currency,
			&e.SpentOn,
			&e.Note,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
