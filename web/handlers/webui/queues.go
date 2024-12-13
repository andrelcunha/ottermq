package webui

import (
	"github.com/gofiber/fiber/v2"
)

func ListQueues(c *fiber.Ctx) error {
	return c.Render("queues", fiber.Map{
		"Title": "Queues",
	})
}
