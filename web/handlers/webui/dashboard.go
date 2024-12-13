package webui

import (
	"github.com/gofiber/fiber/v2"
)

func Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", fiber.Map{
		"Title":    "Dashboard",
		"Greeting": "Welcome to the admin dashboard!",
	})
}
