package income

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) CreateIncome(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	idemKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	bodyBytes := c.Body()

	var req CreateIncomeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	req.ClientName = strings.TrimSpace(req.ClientName)
	if req.ClientName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "client_name required")
	}

	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}

	receivedOn, err := time.Parse("2006-01-02", req.ReceivedOn)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "received_on must be YYYY-MM-DD")
	}

	ctx := userContext(c)
	requestHash := ""
	if idemKey != "" {
		sum := sha256.Sum256(append([]byte(c.Method()+" "+c.Path()+" "), bodyBytes...))
		requestHash = hex.EncodeToString(sum[:])
		// Check for existing idempotent response
		var status int
		var storedBody string
		err := h.Repo.Pool.QueryRow(
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
	}

	inc := &Income{
		UserID:     userID,
		ClientName: req.ClientName,
		Amount:     req.Amount,
		Currency:   "INR",
		ReceivedOn: receivedOn,
		Note:       req.Note,
	}

	id, err := h.Repo.InsertIncome(ctx, inc)
	if err != nil {
		// IMPORTANT: log the actual DB error so we can fix schema issues fast
		_ = c.App().Config().ErrorHandler(c, err) // optional, but keeps fiber aware
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add income: "+err.Error())
	}

	resp := CreateIncomeResponse{
		ID:      id,
		Message: "income added",
	}

	if idemKey != "" && requestHash != "" {
		if buf, mErr := json.Marshal(resp); mErr == nil {
			_, _ = h.Repo.Pool.Exec(
				ctx,
				`INSERT INTO idempotency_keys (owner_id, endpoint, idempotency_key, request_hash, response_status, response_body)
                 VALUES ($1, $2, $3, $4, $5, $6)
                 ON CONFLICT (owner_id, idempotency_key) DO NOTHING`,
				userID,
				"/api/incomes",
				idemKey,
				requestHash,
				fiber.StatusCreated,
				string(buf),
			)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) ListIncomes(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := userContext(c)
	incomes, err := h.Repo.ListIncomesByUser(ctx, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch incomes")
	}

	return c.JSON(incomes)
}

func extractUserID(c *fiber.Ctx) (string, error) {
	val := c.Locals("user_id")
	if val == nil {
		val = c.Locals("userID")
	}
	if val == nil {
		return "", errors.New("user id missing")
	}
	if uid, ok := val.(string); ok && strings.TrimSpace(uid) != "" {
		return uid, nil
	}
	return "", errors.New("user id missing")
}

func userContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}
