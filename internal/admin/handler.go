package admin

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	Pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{Pool: pool}
}

type latestUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type latestTx struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	CreatedAt string `json:"created_at"`
}

type OverviewResponse struct {
	UsersTotal     int64        `json:"users_total"`
	LatestUsers    []latestUser `json:"latest_users"`
	IncomesTotal   int64        `json:"incomes_total"`
	ExpensesTotal  int64        `json:"expenses_total"`
	LatestIncomes  []latestTx   `json:"latest_incomes"`
	LatestExpenses []latestTx   `json:"latest_expenses"`
}

func (h *Handler) Overview(c *fiber.Ctx) error {
	adminKey := os.Getenv("ADMIN_KEY")
	adminKey = strings.TrimSpace(adminKey)
	if adminKey == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "ADMIN_KEY not set on server")
	}

	reqKey := strings.TrimSpace(c.Get("X-Admin-Key"))
	if reqKey == "" || reqKey != adminKey {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := c.UserContext()

	var resp OverviewResponse

	// totals
	if err := h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&resp.UsersTotal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed users_total: "+err.Error())
	}
	if err := h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM incomes`).Scan(&resp.IncomesTotal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed incomes_total: "+err.Error())
	}
	if err := h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM expenses`).Scan(&resp.ExpensesTotal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed expenses_total: "+err.Error())
	}

	// latest users
	{
		rows, err := h.Pool.Query(ctx, `
			SELECT id::text, email, created_at::text
			FROM users
			ORDER BY created_at DESC
			LIMIT 20`)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed latest_users: "+err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			var u latestUser
			if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed scan latest_users: "+err.Error())
			}
			resp.LatestUsers = append(resp.LatestUsers, u)
		}
		if err := rows.Err(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed latest_users rows: "+err.Error())
		}
	}

	// latest incomes
	{
		rows, err := h.Pool.Query(ctx, `
			SELECT id::text, user_id::text, amount, created_at::text
			FROM incomes
			ORDER BY created_at DESC
			LIMIT 20`)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed latest_incomes: "+err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			var t latestTx
			if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.CreatedAt); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed scan latest_incomes: "+err.Error())
			}
			resp.LatestIncomes = append(resp.LatestIncomes, t)
		}
		if err := rows.Err(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed latest_incomes rows: "+err.Error())
		}
	}

	// latest expenses
	{
		rows, err := h.Pool.Query(ctx, `
			SELECT id::text, user_id::text, amount, created_at::text
			FROM expenses
			ORDER BY created_at DESC
			LIMIT 20`)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed latest_expenses: "+err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			var t latestTx
			if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.CreatedAt); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed scan latest_expenses: "+err.Error())
			}
			resp.LatestExpenses = append(resp.LatestExpenses, t)
		}
		if err := rows.Err(); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed latest_expenses rows: "+err.Error())
		}
	}

	return c.JSON(resp)
}
