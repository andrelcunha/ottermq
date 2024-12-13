package webui

import (
	"github.com/gofiber/fiber/v2"
)

func ListExchanges(c *fiber.Ctx) error {
	return c.Render("exchanges", fiber.Map{
		"Title": "Exchange",
	})
}
