package expense

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Expense struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	VendorName string    `json:"vendor_name"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	SpentOn    time.Time `json:"spent_on"`
	Note       *string   `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repo struct {
	DB *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{DB: db}
}

func (r *Repo) AddExpense(
	ctx context.Context,
	userID, vendorName string,
	amount int64,
	spentOn time.Time,
	note *string,
) (string, error) {
	var id string
	err := r.DB.QueryRow(ctx, `
		INSERT INTO expenses (user_id, vendor_name, amount, currency, spent_on, note)
		VALUES ($1, $2, $3, 'INR', $4, $5)
		RETURNING id
	`, userID, vendorName, amount, spentOn, note).Scan(&id)
	return id, err
}

func (r *Repo) ListExpensesByUser(ctx context.Context, userID string) ([]Expense, error) {
	rows, err := r.DB.Query(ctx, `
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
