package webui

import "github.com/gofiber/fiber/v2"

func ListConnections(c *fiber.Ctx) error {
	// get username from cookie
	username := c.Cookies("username")
	return c.Render("connections", fiber.Map{
		"Title":    "Connections",
		"Username": username,
	})
}
