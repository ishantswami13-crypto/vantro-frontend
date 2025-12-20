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

func (r *Repo) Create(ctx context.Context, userID string, typ string, amount int64, note *string) (string, error) {
	var id string
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO transactions (user_id, type, amount, note)
		 VALUES ($1::uuid, $2, $3, $4)
		 RETURNING id`,
		userID, typ, amount, note,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repo) List(ctx context.Context, userID string, limit int) ([]Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.Pool.Query(ctx,
		`SELECT id, user_id::text, type, amount, note, created_at
		 FROM transactions
		 WHERE user_id = $1::uuid
		 ORDER BY created_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Transaction, 0, limit)
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repo) Summary(ctx context.Context, userID string) (SummaryResponse, error) {
	var income int64
	var expense int64

	err := r.Pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN type='income' THEN amount END), 0)::bigint AS income,
			COALESCE(SUM(CASE WHEN type='expense' THEN amount END), 0)::bigint AS expense
		 FROM transactions
		 WHERE user_id = $1::uuid`,
		userID,
	).Scan(&income, &expense)

	if err != nil {
		return SummaryResponse{}, err
	}

	return SummaryResponse{
		Income:  income,
		Expense: expense,
		Net:     income - expense,
	}, nil
}
