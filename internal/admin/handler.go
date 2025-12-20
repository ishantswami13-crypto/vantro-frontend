package admin

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

type UserRow struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	FullName   *string    `json:"full_name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
	IsAdmin    bool       `json:"is_admin"`
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
