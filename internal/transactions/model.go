package transactions

// TxItem represents a unified transaction (income or expense) for listing.
type TxItem struct {
	Type      string `json:"type"` // "income" | "expense"
	ID        string `json:"id"`
	Title     string `json:"title"` // client_name or vendor_name
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Date      string `json:"date"` // YYYY-MM-DD
	CreatedAt string `json:"created_at"`
}

type SummaryResponse struct {
	Income   int64  `json:"income"`
	Expense  int64  `json:"expense"`
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
}
