package reports

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type StatementItem struct {
	Type      string `json:"type"` // income/expense
	ID        string `json:"id"`
	Title     string `json:"title"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Date      string `json:"date"`
	CreatedAt string `json:"created_at"`
}

type StatementResponse struct {
	Currency string          `json:"currency"`
	From     string          `json:"from"`
	To       string          `json:"to"`
	Items    []StatementItem `json:"items"`
}

func (h *Handler) Statement(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	if uidVal == nil {
		uidVal = c.Locals("userID")
	}
	userID, _ := uidVal.(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	if from == "" || to == "" {
		end := time.Now()
		start := end.AddDate(0, 0, -29)
		from = start.Format("2006-01-02")
		to = end.Format("2006-01-02")
	}

	if _, err := time.Parse("2006-01-02", from); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "from must be YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "to must be YYYY-MM-DD")
	}

	ctx := c.UserContext()

	rows, err := h.Pool.Query(ctx, `
SELECT type, id, title, amount, currency, date, created_at
FROM (
  SELECT 'income' AS type,
         id::text AS id,
         client_name AS title,
         amount::bigint AS amount,
         COALESCE(currency,'INR') AS currency,
         received_on::text AS date,
         created_at::text AS created_at
  FROM incomes
  WHERE user_id=$1 AND deleted_at IS NULL AND received_on BETWEEN $2::date AND $3::date

  UNION ALL

  SELECT 'expense' AS type,
         id::text AS id,
         vendor_name AS title,
         amount::bigint AS amount,
         COALESCE(currency,'INR') AS currency,
         spent_on::text AS date,
         created_at::text AS created_at
  FROM expenses
  WHERE user_id=$1 AND deleted_at IS NULL AND spent_on BETWEEN $2::date AND $3::date
) t
ORDER BY date DESC, created_at DESC
LIMIT 1000
`, userID, from, to)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed statement: "+err.Error())
	}
	defer rows.Close()

	var items []StatementItem
	for rows.Next() {
		var it StatementItem
		if err := rows.Scan(&it.Type, &it.ID, &it.Title, &it.Amount, &it.Currency, &it.Date, &it.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "scan statement: "+err.Error())
		}
		items = append(items, it)
	}

	return c.JSON(StatementResponse{
		Currency: "INR",
		From:     from,
		To:       to,
		Items:    items,
	})
}
