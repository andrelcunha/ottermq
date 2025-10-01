package api

import (
	"github.com/andrelcunha/ottermq/internal/core/broker"
	"github.com/andrelcunha/ottermq/internal/core/models"
	"github.com/gofiber/fiber/v2"
)

// ListConnections godoc
// @Summary List all connections
// @Description Get a list of all connections
// @Tags connections
// @Accept json
// @Produce json
// @Success 200 {object} models.ConnectionListResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/connections [get]
func ListConnections(c *fiber.Ctx, b *broker.Broker) error {
	connections := b.ManagerApi.ListConnections()
	if connections == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to list connections",
		})
	}
	return c.Status(fiber.StatusOK).JSON(models.ConnectionListResponse{
		Connections: connections,
	})
}
