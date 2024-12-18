package webui

import (
	"github.com/gofiber/fiber/v2"
)

func ListExchanges(c *fiber.Ctx) error {
	// get username from cookie
	username := c.Cookies("username")
	return c.Render("exchanges", fiber.Map{
		"Title":    "Exchange",
		"Username": username,
	})
}
