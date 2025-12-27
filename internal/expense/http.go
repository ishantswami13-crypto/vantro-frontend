package expense

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

// Fiber-friendly wrappers
func AddExpenseHandler(store *Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AddExpenseRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
		}

		e, err := store.AddExpense(c.Context(), req)
		if err != nil {
			if err == ErrBadRequest {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server error"})
		}
		return c.Status(fiber.StatusCreated).JSON(e)
	}
}

func ListExpensesHandler(store *Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		phone := c.Query("phone")
		limitStr := c.Query("limit")
		limit := 50
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil {
				limit = v
			}
		}

		items, err := store.ListExpenses(c.Context(), ListExpensesParams{
			UserPhone: phone,
			Limit:     limit,
		})
		if err != nil {
			if err == ErrBadRequest {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "phone required"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server error"})
		}
		return c.JSON(items)
	}
}

func MonthlySummaryHandler(store *Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		phone := c.Query("phone")
		yearStr := c.Query("year")
		monthStr := c.Query("month")

		if phone == "" || yearStr == "" || monthStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "phone, year, month required"})
		}

		year, err1 := strconv.Atoi(yearStr)
		month, err2 := strconv.Atoi(monthStr)
		if err1 != nil || err2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year/month invalid"})
		}

		sum, err := store.MonthlySummary(c.Context(), phone, year, month)
		if err != nil {
			if err == ErrBadRequest {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server error"})
		}
		return c.JSON(sum)
	}
}
