package transactions

import "time"

type Transaction struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"` // income | expense
	Amount    int64     `json:"amount" db:"amount"`
	Note      *string   `json:"note,omitempty" db:"note"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateRequest struct {
	Type   string  `json:"type"`   // income | expense
	Amount int64   `json:"amount"` // >= 0 (use rupees or paise consistently)
	Note   *string `json:"note"`
}

type SummaryResponse struct {
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	Net     int64 `json:"net"`
}
