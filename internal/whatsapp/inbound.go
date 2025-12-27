package whatsapp

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ishantswami13-crypto/vantro-backend/internal/expense"
)

// Twilio sends x-www-form-urlencoded with fields like:
// From=whatsapp:+91xxxx  Body=250 food pizza  ProfileName=... etc.
// We respond with TwiML XML (simple). Twilio accepts MessagingResponse XML.
func InboundHandler(expStore *expense.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		from := c.FormValue("From") // "whatsapp:+91..."
		body := c.FormValue("Body")

		phone := normalizeTwilioWhatsAppFrom(from)
		text := strings.TrimSpace(body)

		if phone == "" || text == "" {
			writeTwiML(c, "Send like: 250 food pizza")
			return nil
		}

		_, err := expStore.AddExpense(c.Context(), expense.AddExpenseRequest{
			UserPhone: phone,
			Text:      text,
			Source:    "whatsapp",
		})
		if err != nil {
			writeTwiML(c, "Couldn't log that. Try: 250 food coffee")
			return nil
		}

		writeTwiML(c, "Logged ✅  |  Text more like: 180 uber  |  99 coffee")
		return nil
	}
}

func normalizeTwilioWhatsAppFrom(from string) string {
	// Twilio format: "whatsapp:+91XXXXXXXXXX"
	from = strings.TrimSpace(from)
	from = strings.TrimPrefix(from, "whatsapp:")
	from = strings.TrimSpace(from)
	return from
}

func writeTwiML(c *fiber.Ctx, msg string) {
	c.Set("Content-Type", "application/xml")
	c.Status(fiber.StatusOK)
	// minimal TwiML
	_ = c.SendString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<Response><Message>` + escapeXML(msg) + `</Message></Response>`)
}

func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`\"`, "&quot;",
		`\'`, "&apos;",
	)
	return replacer.Replace(s)
}
