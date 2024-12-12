package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

func BindQueue(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		QueueName    string `json:"queue_name"`
		RoutingKey   string `json:"routing_key"`
	}
	if err := c.BodyParser(&request); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("BIND_QUEUE %s %s %s",
		request.ExchangeName,
		request.QueueName,
		request.RoutingKey)
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

func ListBindings(c *fiber.Ctx) error {

	exchangeName := c.Params("exchange")
	if exchangeName == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Exchange name is required"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Exchange name is required",
		})
	}

	command := fmt.Sprintf("LIST_BINDINGS %s", exchangeName)
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
		// 	"bindings": commandResponse.Data,
		// })
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"bindings": commandResponse.Data,
		})
	}
}

func DeleteBinding(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		QueueName    string `json:"queue_name"`
		RoutingKey   string `json:"routing_key"`
	}
	if err := c.BodyParser(&request); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	command := fmt.Sprintf("DELETE_BINDING %s %s %s",
		request.ExchangeName,
		request.QueueName,
		request.RoutingKey)
	response, err := utils.SendCommand(command)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse response"})
	}

	if commandResponse.Status == "ERROR" {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": commandResponse.Message})
	} else {
		// c.JSON(http.StatusOK, gin.H{
		// 	"message": commandResponse.Message,
		// })
		return c.Status(http.StatusOK).JSON(fiber.Map{"message": commandResponse.Message})
	}
}
