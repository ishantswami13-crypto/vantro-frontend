package http

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHandler struct {
	DB *pgxpool.Pool
}

func NewTransactionHandler(db *pgxpool.Pool) *TransactionHandler {
	return &TransactionHandler{DB: db}
}

func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	userID, err := userUUID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}

	if err := c.BodyParser(&body); err != nil {
		return fiber.ErrBadRequest
	}

	businessID, err := h.ensureBusiness(c, userID)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	_, err = h.DB.Exec(
		c.Context(),
		`INSERT INTO transactions (user_id, business_id, type, amount, note)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, businessID, body.Type, body.Amount, body.Description,
	)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *TransactionHandler) List(c *fiber.Ctx) error {
	userID, err := userUUID(c)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	rows, err := h.DB.Query(
		c.Context(),
		`SELECT id, type, amount::float8, COALESCE(note, ''), created_at
		 FROM transactions
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	defer rows.Close()

	var result []fiber.Map
	for rows.Next() {
		var (
			id        int64
			ttype     string
			desc      string
			amount    float64
			createdAt time.Time
		)

		if err := rows.Scan(&id, &ttype, &amount, &desc, &createdAt); err != nil {
			return fiber.ErrInternalServerError
		}

		result = append(result, fiber.Map{
			"id":          id,
			"type":        ttype,
			"amount":      amount,
			"description": desc,
			"created_at":  createdAt.Format(time.RFC3339),
		})
	}

	if err := rows.Err(); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(result)
}

func (h *TransactionHandler) ensureBusiness(c *fiber.Ctx, userID uuid.UUID) (int64, error) {
	var bizID int64
	err := h.DB.QueryRow(
		c.Context(),
		`SELECT id
		 FROM businesses
		 WHERE owner_user_id = $1
		 ORDER BY created_at ASC
		 LIMIT 1`,
		userID,
	).Scan(&bizID)
	if err == nil {
		return bizID, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		err = h.DB.QueryRow(
			c.Context(),
			`INSERT INTO businesses (owner_user_id, name, currency)
			 VALUES ($1, 'Default Business', 'INR')
			 RETURNING id`,
			userID,
		).Scan(&bizID)
	}

	return bizID, err
}

func userUUID(c *fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals("user_id")
	if val == nil {
		val = c.Locals("userID")
	}

	switch v := val.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(strings.TrimSpace(v))
	default:
		return uuid.Nil, errors.New("user id missing")
	}
}
