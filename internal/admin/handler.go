package admin

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB       *pgxpool.Pool
	AdminKey string
}

func NewHandler(db *pgxpool.Pool, adminKey string) *Handler {
	return &Handler{DB: db, AdminKey: strings.TrimSpace(adminKey)}
}

type UserRow struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	FullName   *string    `json:"full_name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	IsAdmin    bool       `json:"is_admin"`
}

type overviewUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type overviewTxn struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount"`
	CreatedAt string `json:"created_at"`
}

type overviewResponse struct {
	UsersTotal     int64          `json:"users_total"`
	LatestUsers    []overviewUser `json:"latest_users"`
	IncomesTotal   int64          `json:"incomes_total"`
	ExpensesTotal  int64          `json:"expenses_total"`
	LatestIncomes  []overviewTxn  `json:"latest_incomes"`
	LatestExpenses []overviewTxn  `json:"latest_expenses"`
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, `
		SELECT id::text, email, NULLIF(full_name,''), created_at, last_seen_at, is_admin
		FROM users
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list users: "+err.Error())
	}
	defer rows.Close()

	out := make([]UserRow, 0, 200)
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.CreatedAt, &u.LastSeenAt, &u.IsAdmin); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read users: "+err.Error())
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list users: "+err.Error())
	}

	return c.JSON(out)
}

func (h *Handler) Stats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var total int64
	var active7d int64

	err := h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "stats error: "+err.Error())
	}

	err = h.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE last_seen_at IS NOT NULL
		  AND last_seen_at >= NOW() - INTERVAL '7 days'
	`).Scan(&active7d)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "stats error: "+err.Error())
	}

	return c.JSON(fiber.Map{
		"total_users": total,
		"active_7d":   active7d,
	})
}

// Overview returns admin-only analytics without exposing sensitive fields.
func (h *Handler) Overview(c *fiber.Ctx) error {
	keyHeader := strings.TrimSpace(c.Get("X-Admin-Key"))
	if keyHeader == "" || keyHeader != strings.TrimSpace(h.AdminKey) || keyHeader == "" || h.AdminKey == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var resp overviewResponse

	if err := h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&resp.UsersTotal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load users_total: "+err.Error())
	}

	rowsUsers, err := h.DB.Query(ctx, `
		SELECT id::text, email, created_at::text
		FROM users
		ORDER BY created_at DESC
		LIMIT 20
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load latest users: "+err.Error())
	}
	defer rowsUsers.Close()
	for rowsUsers.Next() {
		var u overviewUser
		if err := rowsUsers.Scan(&u.ID, &u.Email, &u.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read latest users: "+err.Error())
		}
		resp.LatestUsers = append(resp.LatestUsers, u)
	}
	if err := rowsUsers.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read latest users: "+err.Error())
	}

	if err := h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM incomes`).Scan(&resp.IncomesTotal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load incomes_total: "+err.Error())
	}
	if err := h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM expenses`).Scan(&resp.ExpensesTotal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load expenses_total: "+err.Error())
	}

	rowsIncome, err := h.DB.Query(ctx, `
		SELECT id::text, user_id::text, amount, created_at::text
		FROM incomes
		ORDER BY created_at DESC
		LIMIT 20
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load latest incomes: "+err.Error())
	}
	defer rowsIncome.Close()
	for rowsIncome.Next() {
		var it overviewTxn
		if err := rowsIncome.Scan(&it.ID, &it.UserID, &it.Amount, &it.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read latest incomes: "+err.Error())
		}
		resp.LatestIncomes = append(resp.LatestIncomes, it)
	}
	if err := rowsIncome.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read latest incomes: "+err.Error())
	}

	rowsExpense, err := h.DB.Query(ctx, `
		SELECT id::text, user_id::text, amount, created_at::text
		FROM expenses
		ORDER BY created_at DESC
		LIMIT 20
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load latest expenses: "+err.Error())
	}
	defer rowsExpense.Close()
	for rowsExpense.Next() {
		var it overviewTxn
		if err := rowsExpense.Scan(&it.ID, &it.UserID, &it.Amount, &it.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to read latest expenses: "+err.Error())
		}
		resp.LatestExpenses = append(resp.LatestExpenses, it)
	}
	if err := rowsExpense.Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to read latest expenses: "+err.Error())
	}

	return c.JSON(resp)
}
