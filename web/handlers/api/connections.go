package api

import (
	// "encoding/json"

	// "github.com/andrelcunha/ottermq/pkg/common/communication/api"
	// "github.com/andrelcunha/ottermq/web/utils"
	"github.com/andrelcunha/ottermq/internal/core/broker"
	"github.com/gofiber/fiber/v2"
)

// ListConnections godoc
// @Summary List all connections
// @Description Get a list of all connections
// @Tags connections
// @Accept json
// @Produce json
// @Success 200 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/connections [get]
func ListConnections(c *fiber.Ctx, b *broker.Broker) error {
	connections := broker.ListConnections(b)
	if connections == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list connections",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"connections": connections,
	})
}
