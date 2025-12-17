package income

import "time"

type Income struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	ClientName string    `db:"client_name" json:"client_name"`
	Amount     int64     `db:"amount" json:"amount"`
	Currency   string    `db:"currency" json:"currency"`
	ReceivedOn time.Time `db:"received_on" json:"received_on"`
	Note       *string   `db:"note" json:"note,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type CreateIncomeRequest struct {
	ClientName string  `json:"client_name"`
	Amount     int64   `json:"amount"`
	ReceivedOn string  `json:"received_on"`
	Note       *string `json:"note"`
}

type CreateIncomeResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
