package http

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BusinessHandler struct {
	DB *pgxpool.Pool
}

func NewBusinessHandler(db *pgxpool.Pool) *BusinessHandler {
	return &BusinessHandler{DB: db}
}

type createBusinessReq struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

func (h *BusinessHandler) Create(c *fiber.Ctx) error {
	userIDAny := c.Locals("user_id")
	userID, ok := userIDAny.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	var req createBusinessReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name required")
	}
	if strings.TrimSpace(req.Currency) == "" {
		req.Currency = "INR"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int64
	err := h.DB.QueryRow(ctx,
		`INSERT INTO businesses (owner_user_id, name, currency) VALUES ($1,$2,$3) RETURNING id`,
		userID, req.Name, req.Currency,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create business")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *BusinessHandler) List(c *fiber.Ctx) error {
	userIDAny := c.Locals("user_id")
	userID, ok := userIDAny.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx,
		`SELECT id, name, currency, created_at FROM businesses WHERE owner_user_id=$1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list businesses")
	}
	defer rows.Close()

	type outBiz struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Currency  string `json:"currency"`
		CreatedAt string `json:"created_at"`
	}

	out := []outBiz{}
	for rows.Next() {
		var b outBiz
		var t time.Time
		if err := rows.Scan(&b.ID, &b.Name, &b.Currency, &t); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not read businesses")
		}
		b.CreatedAt = t.Format(time.RFC3339)
		out = append(out, b)
	}
	return c.JSON(out)
}
