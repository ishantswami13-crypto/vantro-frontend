package whatsapp

import (
	"fmt"
	"html"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Inbound handles Twilio WhatsApp webhook (x-www-form-urlencoded) and returns TwiML XML.
func Inbound(c *fiber.Ctx) error {
	from := c.FormValue("From") // e.g. "whatsapp:+91..."
	body := c.FormValue("Body") // message text

	fmt.Printf("[twilio] from=%s body=%q\n", from, body)

	if strings.TrimSpace(body) == "" {
		return twiml(c, "Send something like: 250 food pizza")
	}

	return twiml(c, "Logged ✅")
}

func twiml(c *fiber.Ctx, msg string) error {
	safe := html.EscapeString(msg)
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Message>%s</Message>
</Response>`, safe)

	c.Set("Content-Type", "application/xml; charset=utf-8")
	return c.Status(fiber.StatusOK).SendString(xml)
}
