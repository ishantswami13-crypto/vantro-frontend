package reports

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{Pool: pool}
}

type DayPoint struct {
	Date    string `json:"date"` // YYYY-MM-DD
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
	Balance int64  `json:"balance"`
}

type ReportResponse struct {
	Currency     string     `json:"currency"`
	From         string     `json:"from"`
	To           string     `json:"to"`
	TotalIncome  int64      `json:"total_income"`
	TotalExpense int64      `json:"total_expense"`
	Balance      int64      `json:"balance"`
	Daily        []DayPoint `json:"daily"`
}

func (h *Handler) Get(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	if uidVal == nil {
		uidVal = c.Locals("userID")
	}
	userID, _ := uidVal.(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	from := strings.TrimSpace(c.Query("from")) // YYYY-MM-DD
	to := strings.TrimSpace(c.Query("to"))     // YYYY-MM-DD

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

	var totalIncome int64
	if err := h.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0)
		FROM incomes
		WHERE user_id=$1
		  AND deleted_at IS NULL
		  AND received_on BETWEEN $2::date AND $3::date
	`, userID, from, to).Scan(&totalIncome); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed total income: "+err.Error())
	}

	var totalExpense int64
	if err := h.Pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount),0)
		FROM expenses
		WHERE user_id=$1
		  AND deleted_at IS NULL
		  AND spent_on BETWEEN $2::date AND $3::date
	`, userID, from, to).Scan(&totalExpense); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed total expense: "+err.Error())
	}

	rows, err := h.Pool.Query(ctx, `
WITH days AS (
  SELECT d::date AS day
  FROM generate_series($2::date, $3::date, interval '1 day') AS d
),
inc AS (
  SELECT received_on::date AS day, SUM(amount)::bigint AS income
  FROM incomes
  WHERE user_id=$1 AND deleted_at IS NULL AND received_on BETWEEN $2::date AND $3::date
  GROUP BY 1
),
exp AS (
  SELECT spent_on::date AS day, SUM(amount)::bigint AS expense
  FROM expenses
  WHERE user_id=$1 AND deleted_at IS NULL AND spent_on BETWEEN $2::date AND $3::date
  GROUP BY 1
)
SELECT
  days.day::text,
  COALESCE(inc.income,0)::bigint,
  COALESCE(exp.expense,0)::bigint
FROM days
LEFT JOIN inc ON inc.day = days.day
LEFT JOIN exp ON exp.day = days.day
ORDER BY days.day ASC
`, userID, from, to)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed daily series: "+err.Error())
	}
	defer rows.Close()

	var daily []DayPoint
	var running int64
	for rows.Next() {
		var day string
		var incAmt, expAmt int64
		if err := rows.Scan(&day, &incAmt, &expAmt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed scan daily: "+err.Error())
		}
		running += incAmt - expAmt
		daily = append(daily, DayPoint{
			Date:    day,
			Income:  incAmt,
			Expense: expAmt,
			Balance: running,
		})
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "daily rows error: "+err.Error())
	}

	resp := ReportResponse{
		Currency:     "INR",
		From:         from,
		To:           to,
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Balance:      totalIncome - totalExpense,
		Daily:        daily,
	}

	return c.JSON(resp)
}
