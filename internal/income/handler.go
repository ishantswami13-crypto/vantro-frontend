package income

import (
	"context"
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

	return c.Status(fiber.StatusCreated).JSON(CreateIncomeResponse{
		ID:      id,
		Message: "income added",
	})
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
