package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionsHandler struct {
	DB *pgxpool.Pool
}

type createTxnReq struct {
	Type       string `json:"type"`   // "income" | "expense"
	Amount     int64  `json:"amount"` // paise, > 0
	Note       string `json:"note"`
	BusinessID int64  `json:"business_id"`
}

func NewTransactionsHandler(db *pgxpool.Pool) *TransactionsHandler {
	return &TransactionsHandler{DB: db}
}

// NOTE: this assumes your auth middleware sets user_id in c.Locals("user_id").
// If your code uses a different key, replace "user_id" with your actual one.
func (h *TransactionsHandler) Create(c *fiber.Ctx) error {
	userIDAny := c.Locals("user_id")
	userID, ok := userIDAny.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	idemKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	bodyBytes := c.Body()

	var req createTxnReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}
	if req.Type != "income" && req.Type != "expense" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be income or expense")
	}
	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be positive (paise)")
	}
	if req.BusinessID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "business_id required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestHash := ""
	if idemKey != "" {
		sum := sha256.Sum256(append([]byte(c.Method()+" "+c.Path()+" "), bodyBytes...))
		requestHash = hex.EncodeToString(sum[:])

		var status int
		var storedBody string
		err := h.DB.QueryRow(
			ctx,
			`SELECT response_status, response_body
             FROM idempotency_keys
             WHERE owner_id = $1 AND idempotency_key = $2`,
			userID, idemKey,
		).Scan(&status, &storedBody)
		if err == nil {
			c.Status(status)
			c.Type("application/json")
			return c.SendString(storedBody)
		}
		if err != nil && err != pgx.ErrNoRows {
			// On unexpected error, continue without idempotent shortcut.
		}
	}

	var belongs bool
	if err := h.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM businesses WHERE id=$1 AND owner_user_id=$2)`,
		req.BusinessID, userID,
	).Scan(&belongs); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not validate business")
	}
	if !belongs {
		return fiber.NewError(fiber.StatusBadRequest, "invalid business_id")
	}

	var id int64
	err := h.DB.QueryRow(ctx,
		`INSERT INTO transactions (user_id, business_id, type, amount, note)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		userID, req.BusinessID, req.Type, req.Amount, req.Note,
	).Scan(&id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create transaction")
	}

	resp := fiber.Map{
		"id": id,
	}

	if idemKey != "" && requestHash != "" {
		if buf, mErr := json.Marshal(resp); mErr == nil {
			_, _ = h.DB.Exec(
				ctx,
				`INSERT INTO idempotency_keys (owner_id, endpoint, idempotency_key, request_hash, response_status, response_body)
                 VALUES ($1, $2, $3, $4, $5, $6)
                 ON CONFLICT (owner_id, idempotency_key) DO NOTHING`,
				userID,
				"/api/transactions",
				idemKey,
				requestHash,
				fiber.StatusCreated,
				string(buf),
			)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *TransactionsHandler) Summary(c *fiber.Ctx) error {
	userIDAny := c.Locals("user_id")
	userID, ok := userIDAny.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	businessID, err := parseBusinessID(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var income int64
	var expense int64

	err = h.DB.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type='income'  THEN amount END)::bigint, 0) AS income,
			COALESCE(SUM(CASE WHEN type='expense' THEN amount END)::bigint, 0) AS expense
		FROM transactions
		WHERE user_id = $1 AND business_id = $2
	`, userID, businessID).Scan(&income, &expense)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not compute summary")
	}

	return c.JSON(fiber.Map{
		"income":  income,
		"expense": expense,
		"net":     income - expense,
	})
}

func (h *TransactionsHandler) List(c *fiber.Ctx) error {
	userIDAny := c.Locals("user_id")
	userID, ok := userIDAny.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	businessID, err := parseBusinessID(c)
	if err != nil {
		return err
	}

	type txn struct {
		ID        int64  `json:"id"`
		Type      string `json:"type"`
		Amount    int64  `json:"amount"`
		Note      string `json:"note"`
		CreatedAt string `json:"created_at"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := h.DB.Query(ctx, `
		SELECT id, type, amount, COALESCE(note,''), created_at
		FROM transactions
		WHERE user_id = $1 AND business_id = $2
		ORDER BY created_at DESC
		LIMIT 50
	`, userID, businessID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not fetch transactions")
	}
	defer rows.Close()

	out := make([]txn, 0, 50)
	for rows.Next() {
		var t txn
		var created time.Time
		if err := rows.Scan(&t.ID, &t.Type, &t.Amount, &t.Note, &created); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not read transactions")
		}
		t.CreatedAt = created.Format(time.RFC3339)
		out = append(out, t)
	}

	return c.JSON(out)
}

func parseBusinessID(c *fiber.Ctx) (int64, error) {
	raw := strings.TrimSpace(c.Query("business_id"))
	if raw == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "business_id is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid business_id")
	}
	return id, nil
}
