package transactions

import (
	"context"
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{Repo: repo}
}

func getUserID(c *fiber.Ctx) (string, bool) {
	if v := c.Locals("user_id"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s), true
		}
	}
	return "", false
}

func (h *Handler) ListLatest(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := userContext(c)

	items, err := h.Repo.ListLatest(ctx, userID, 50)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load transactions: "+err.Error())
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) GetSummary(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := userContext(c)

	s, err := h.Repo.GetSummary(ctx, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to compute summary: "+err.Error())
	}

	return c.JSON(s)
}

// ExportCSV streams the latest transactions as CSV for the authenticated user.
func (h *Handler) ExportCSV(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	ctx := userContext(c)
	items, err := h.Repo.ListLatest(ctx, userID, 500)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load transactions: "+err.Error())
	}

	c.Set("Content-Type", "text/csv")
	c.Attachment("transactions.csv")

	var b strings.Builder
	w := csv.NewWriter(&b)

	_ = w.Write([]string{"type", "id", "title", "amount", "currency", "date", "created_at"})
	for _, it := range items {
		record := []string{
			it.Type,
			it.ID,
			it.Title,
			strconv.FormatInt(it.Amount, 10),
			it.Currency,
			it.Date,
			it.CreatedAt,
		}
		_ = w.Write(record)
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to build CSV")
	}

	return c.SendString(b.String())
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	if uidVal == nil {
		uidVal = c.Locals("userID")
	}
	userID, _ := uidVal.(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	typ := normalizeType(c.Params("type"))
	if typ == "" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be income or expense")
	}

	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id required")
	}

	var err error
	if typ == "income" {
		err = h.Repo.DeleteIncomeByID(c.UserContext(), userID, id)
	} else {
		err = h.Repo.DeleteExpenseByID(c.UserContext(), userID, id)
	}

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fiber.NewError(fiber.StatusNotFound, "not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "delete failed: "+err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) Undo(c *fiber.Ctx) error {
	uidVal := c.Locals("user_id")
	if uidVal == nil {
		uidVal = c.Locals("userID")
	}
	userID, _ := uidVal.(string)
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	typ := normalizeType(c.Params("type"))
	if typ == "" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be income or expense")
	}

	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id required")
	}

	var err error
	if typ == "income" {
		err = h.Repo.UndoIncomeByID(c.UserContext(), userID, id)
	} else {
		err = h.Repo.UndoExpenseByID(c.UserContext(), userID, id)
	}

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fiber.NewError(fiber.StatusNotFound, "not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "undo failed: "+err.Error())
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func userContext(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}
