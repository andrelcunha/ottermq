package webui

import (
	"github.com/gofiber/fiber/v2"
)

func ListQueues(c *fiber.Ctx) error {
	// get username from cookie
	username := c.Cookies("username")
	return c.Render("queues", fiber.Map{
		"Title":    "Queues",
		"Username": username,
	})
}
