package income

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

func (r *Repository) InsertIncome(ctx context.Context, inc *Income) (string, error) {
	var id string
	err := r.Pool.QueryRow(
		ctx,
		`INSERT INTO incomes (user_id, client_name, amount, currency, received_on, note)
         VALUES ($1, $2, $3, COALESCE($4, 'INR'), $5, $6)
         RETURNING id`,
		inc.UserID,
		inc.ClientName,
		inc.Amount,
		inc.Currency,
		inc.ReceivedOn,
		inc.Note,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *Repository) ListIncomesByUser(ctx context.Context, userID string) ([]Income, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT id, user_id, client_name, amount, currency, received_on, note, created_at
		 FROM incomes
		 WHERE user_id = $1
		 ORDER BY received_on DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incomes []Income
	for rows.Next() {
		var inc Income
		if err := rows.Scan(
			&inc.ID,
			&inc.UserID,
			&inc.ClientName,
			&inc.Amount,
			&inc.Currency,
			&inc.ReceivedOn,
			&inc.Note,
			&inc.CreatedAt,
		); err != nil {
			return nil, err
		}
		incomes = append(incomes, inc)
	}

	return incomes, nil
}
