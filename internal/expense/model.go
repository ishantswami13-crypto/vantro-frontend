package expense

import "time"

type Expense struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	VendorName string    `db:"vendor_name" json:"vendor_name"`
	Amount     int64     `db:"amount" json:"amount"`
	Currency   string    `db:"currency" json:"currency"`
	SpentOn    time.Time `db:"spent_on" json:"spent_on"`
	Note       *string   `db:"note" json:"note,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type CreateExpenseRequest struct {
	VendorName string  `json:"vendor_name"`
	Amount     int64   `json:"amount"`
	SpentOn    string  `json:"spent_on"` // YYYY-MM-DD
	Note       *string `json:"note"`
}

type CreateExpenseResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
