package whatsapp

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func Inbound(c *fiber.Ctx) error {
	from := c.FormValue("From")
	body := c.FormValue("Body")

	log.Printf("[twilio] from=%s body=%s\n", from, body)

	reply := `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Message>
		Thanks! I received: ` + body + `
	</Message>
</Response>`

	c.Set("Content-Type", "text/xml")
	return c.SendString(reply)
}
