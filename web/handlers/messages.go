package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"

	"github.com/gofiber/fiber/v2"
)

// PublishMessage godoc
// @Summary Publish a message to an exchange
// @Description Publish a message to the specified exchange with a routing key
// @Tags messages
// @Accept json
// @Produce json
// @Param message body models.PublishMessageRequest true "Message details"
// @Success 200 {object} models.FiberMap
// @Failure 400 {object} models.FiberMap
// @Failure 500 {object} models.FiberMap
// @Router /api/messages [post]
func PublishMessage(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		RoutingKey   string `json:"routing_key"`
		Message      string `json:"message"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("PUBLISH %s %s %s", request.ExchangeName, request.RoutingKey, request.Message)
	response, err := utils.SendCommand(command)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}
	if commandResponse.Status == "ERROR" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": commandResponse.Message,
		})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": commandResponse.Data,
		})
	}
}

// AckMessage godoc
// @Summary Acknowledge a message
// @Description Acknowledge a message with the specified ID
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} models.FiberMap
// @Failure 400 {object} models.FiberMap
// @Failure 500 {object} models.FiberMap
// @Router /api/messages/{id}/ack [post]
func AckMessage(c *fiber.Ctx) error {
	msgID := c.Params("id")
	if msgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message ID is required",
		})
	}
	command := fmt.Sprintf("ACK %s", msgID)
	response, err := utils.SendCommand(command)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}

	if commandResponse.Status == "ERROR" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": commandResponse.Message,
		})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": commandResponse.Message,
		})
	}
}
