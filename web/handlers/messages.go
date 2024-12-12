package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"

	"github.com/gofiber/fiber/v2"
)

func PublishMessage(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		RoutingKey   string `json:"routing_key"`
		Message      string `json:"message"`
	}
	if err := c.BodyParser(&request); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("PUBLISH %s %s %s", request.ExchangeName, request.RoutingKey, request.Message)
	response, err := utils.SendCommand(command)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}
	if commandResponse.Status == "ERROR" {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": commandResponse.Message,
		})
	} else {
		// c.JSON(http.StatusOK, gin.H{
		// 	"data": commandResponse.Data,
		// })
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": commandResponse.Data,
		})
	}
}

func AckMessage(c *fiber.Ctx) error {
	msgID := c.Params("id")
	if msgID == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "message ID is required"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "message ID is required",
		})
	}
	command := fmt.Sprintf("ACK %s", msgID)
	response, err := utils.SendCommand(command)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse response",
		})
	}

	if commandResponse.Status == "ERROR" {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": commandResponse.Message,
		})
	} else {
		// c.JSON(http.StatusOK, gin.H{
		// 	"message": commandResponse.Message,
		// })
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": commandResponse.Message,
		})
	}
}
