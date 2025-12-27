package billing

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

func CreatePaymentLinkHandler(store *Store, rp *RazorpayClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateLinkRequest
		if err := c.BodyParser(&req); err != nil || req.UserPhone == "" {
			return c.Status(fiber.StatusBadRequest).SendString("user_phone required")
		}

		link, err := rp.CreateMonthlyLink(c.Context(), req.UserPhone)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("razorpay error")
		}

		// Optional: auto-activate. Better to activate via webhook; keep commented.
		// _ = store.ActivateFor30Days(c.Context(), req.UserPhone, link.ID)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"payment_link_id": link.ID,
			"short_url":       link.ShortURL,
			"status":          link.Status,
		})
	}
}

func RazorpayWebhookHandler(store *Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Body()
		sig := c.Get("X-Razorpay-Signature")
		secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")

		if secret == "" || sig == "" || !VerifyWebhookSignature(raw, sig, secret) {
			return c.Status(fiber.StatusUnauthorized).SendString("invalid signature")
		}

		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("bad payload")
		}

		evt, _ := payload["event"].(string)
		if evt != "payment_link.paid" {
			return c.Status(fiber.StatusOK).SendString("ignored")
		}

		phone := ""
		plID := ""

		if p, ok := payload["payload"].(map[string]any); ok {
			if pl, ok := p["payment_link"].(map[string]any); ok {
				if ent, ok := pl["entity"].(map[string]any); ok {
					if id, ok := ent["id"].(string); ok {
						plID = id
					}
					if cust, ok := ent["customer"].(map[string]any); ok {
						if cphone, ok := cust["contact"].(string); ok {
							phone = cphone
						}
					}
				}
			}
		}

		if phone == "" {
			return c.Status(fiber.StatusBadRequest).SendString("phone not found")
		}

		if err := store.ActivateFor30Days(c.Context(), phone, plID); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("db error")
		}

		return c.Status(fiber.StatusOK).SendString("ok")
	}
}
