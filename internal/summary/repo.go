package summary

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	DB *pgxpool.Pool
}

type Summary struct {
	TotalIncome  int64  `json:"total_income"`
	TotalExpense int64  `json:"total_expense"`
	Net          int64  `json:"net"`
	Currency     string `json:"currency"`
}

func (r Repo) GetByUser(ctx context.Context, userID string, month string) (Summary, error) {
	var income int64
	var expense int64

	// If month provided, filter by YYYY-MM
	if month != "" {
		err := r.DB.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0)::bigint
			FROM incomes
			WHERE user_id = $1
			  AND deleted_at IS NULL
			  AND to_char(received_on, 'YYYY-MM') = $2
		`, userID, month).Scan(&income)
		if err != nil {
			return Summary{}, err
		}

		err = r.DB.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0)::bigint
			FROM expenses
			WHERE user_id = $1
			  AND deleted_at IS NULL
			  AND to_char(spent_on, 'YYYY-MM') = $2
		`, userID, month).Scan(&expense)
		if err != nil {
			return Summary{}, err
		}
	} else {
		// No month = all time
		err := r.DB.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0)::bigint
			FROM incomes
			WHERE user_id = $1
			  AND deleted_at IS NULL
		`, userID).Scan(&income)
		if err != nil {
			return Summary{}, err
		}

		err = r.DB.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0)::bigint
			FROM expenses
			WHERE user_id = $1
			  AND deleted_at IS NULL
		`, userID).Scan(&expense)
		if err != nil {
			return Summary{}, err
		}
	}

	return Summary{
		TotalIncome:  income,
		TotalExpense: expense,
		Net:          income - expense,
		Currency:     "INR",
	}, nil
}
