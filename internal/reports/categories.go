package reports

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CategoryRow struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
	Count    int64  `json:"count"`
	Type     string `json:"type"` // income or expense
}

type CategoriesResponse struct {
	Currency string        `json:"currency"`
	From     string        `json:"from"`
	To       string        `json:"to"`
	Top      []CategoryRow `json:"top"`
}

func (h *Handler) Categories(c *fiber.Ctx) error {
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
WITH income_top AS (
  SELECT client_name AS category, SUM(amount)::bigint AS total, COUNT(*)::bigint AS count, 'income' AS type
  FROM incomes
  WHERE user_id=$1 AND deleted_at IS NULL AND received_on BETWEEN $2::date AND $3::date
  GROUP BY 1
),
expense_top AS (
  SELECT vendor_name AS category, SUM(amount)::bigint AS total, COUNT(*)::bigint AS count, 'expense' AS type
  FROM expenses
  WHERE user_id=$1 AND deleted_at IS NULL AND spent_on BETWEEN $2::date AND $3::date
  GROUP BY 1
),
all_rows AS (
  SELECT * FROM income_top
  UNION ALL
  SELECT * FROM expense_top
)
SELECT category, total, count, type
FROM all_rows
ORDER BY total DESC
LIMIT 12
`, userID, from, to)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed categories: "+err.Error())
	}
	defer rows.Close()

	var out []CategoryRow
	for rows.Next() {
		var r CategoryRow
		if err := rows.Scan(&r.Category, &r.Total, &r.Count, &r.Type); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "scan categories: "+err.Error())
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "categories rows error: "+err.Error())
	}

	return c.JSON(CategoriesResponse{
		Currency: "INR",
		From:     from,
		To:       to,
		Top:      out,
	})
}
