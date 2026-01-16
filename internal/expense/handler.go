package expense

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/points"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) CreateExpense(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	idemKey := strings.TrimSpace(c.Get("Idempotency-Key"))
	bodyBytes := c.Body()

	var req CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	req.VendorName = strings.TrimSpace(req.VendorName)
	if req.VendorName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "vendor_name required")
	}
	if req.Amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "amount must be greater than zero")
	}

	spentOn, err := time.Parse("2006-01-02", req.SpentOn)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "spent_on must be YYYY-MM-DD")
	}

	ctx := userContext(c)
	requestHash := ""
	if idemKey != "" {
		sum := sha256.Sum256(append([]byte(c.Method()+" "+c.Path()+" "), bodyBytes...))
		requestHash = hex.EncodeToString(sum[:])

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

	exp := &LegacyExpense{
		UserID:     userID,
		VendorName: req.VendorName,
		Amount:     req.Amount,
		Currency:   "INR",
		SpentOn:    spentOn,
		Note:       req.Note,
	}

	id, err := h.Repo.InsertExpense(ctx, exp)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add expense: "+err.Error())
	}

	// Award points for outgoing payments
	if _, err := points.AwardPointsForTransaction(ctx, h.Repo.Pool, userID, nil, exp.Amount, "expense_reward"); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to award points: "+err.Error())
	}

	resp := CreateExpenseResponse{
		ID:      id,
		Message: "expense added",
	}

	if idemKey != "" && requestHash != "" {
		if buf, mErr := json.Marshal(resp); mErr == nil {
			_, _ = h.Repo.Pool.Exec(
				ctx,
				`INSERT INTO idempotency_keys (owner_id, endpoint, idempotency_key, request_hash, response_status, response_body)
                 VALUES ($1, $2, $3, $4, $5, $6)
                 ON CONFLICT (owner_id, idempotency_key) DO NOTHING`,
				userID,
				"/api/expenses",
				idemKey,
				requestHash,
				fiber.StatusCreated,
				string(buf),
			)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) ListExpenses(c *fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}

	items, err := h.Repo.ListExpensesByUser(userContext(c), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list expenses: "+err.Error())
	}

	return c.JSON(items)
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
