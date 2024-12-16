package webui

import "github.com/gofiber/fiber/v2"

func ListConnections(c *fiber.Ctx) error {
	return c.Render("connections", fiber.Map{
		"Title": "Connections",
	})
}
