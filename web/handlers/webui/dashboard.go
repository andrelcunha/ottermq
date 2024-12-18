package webui

import (
	"github.com/gofiber/fiber/v2"
)

func Dashboard(c *fiber.Ctx) error {
	// get username from cookie
	username := c.Cookies("username")
	return c.Render("dashboard", fiber.Map{
		"Title":    "Dashboard",
		"Username": username,
		"Greeting": "Welcome to the admin dashboard!",
	})
}
