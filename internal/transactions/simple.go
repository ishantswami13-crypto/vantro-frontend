package transactions

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ishantswami13-crypto/vantro-backend/internal/points"
)

type SimpleRepo struct {
	Pool *pgxpool.Pool
}

type CreateTxnRequest struct {
	Amount    int64  `json:"amount"`
	Direction string `json:"direction"` // IN or OUT
	Note      string `json:"note"`
}

type Txn struct {
	ID        string `json:"id"`
	Amount    int64  `json:"amount"`
	Direction string `json:"direction"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`
}

func NewSimpleRepo(pool *pgxpool.Pool) *SimpleRepo {
	return &SimpleRepo{Pool: pool}
}

func (r *SimpleRepo) Create(ctx context.Context, userID string, req CreateTxnRequest) (Txn, error) {
	req.Direction = strings.ToUpper(strings.TrimSpace(req.Direction))
	if req.Direction != "IN" && req.Direction != "OUT" {
		return Txn{}, errors.New("direction must be IN or OUT")
	}
	if req.Amount <= 0 {
		return Txn{}, errors.New("amount must be > 0")
	}

	tx, err := r.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Txn{}, err
	}
	defer tx.Rollback(ctx)

	var out Txn
	err = tx.QueryRow(ctx, `
INSERT INTO user_transactions (user_id, amount, direction, note)
VALUES ($1, $2, $3, NULLIF($4,''))
RETURNING id::text, amount, direction, COALESCE(note,''), created_at::text
`, userID, req.Amount, req.Direction, req.Note).Scan(&out.ID, &out.Amount, &out.Direction, &out.Note, &out.CreatedAt)
	if err != nil {
		return Txn{}, err
	}

	// Auto points for OUT
	if req.Direction == "OUT" {
		_, err = points.AwardPointsForTransaction(ctx, r.Pool, userID, &out.ID, req.Amount, "txn_out")
		if err != nil {
			return Txn{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Txn{}, err
	}

	return out, nil
}

func (r *SimpleRepo) List(ctx context.Context, userID string, limit int) ([]Txn, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := r.Pool.Query(ctx, `
SELECT id::text, amount, direction, COALESCE(note,''), created_at::text
FROM user_transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Txn
	for rows.Next() {
		var t Txn
		if err := rows.Scan(&t.ID, &t.Amount, &t.Direction, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
