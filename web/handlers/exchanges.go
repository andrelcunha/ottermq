package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

func ListExchanges(c *fiber.Ctx) error {
	response, err := utils.SendCommand("LIST_EXCHANGES")
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
		// 	"exchanges": commandResponse.Data,
		// })
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"exchanges": commandResponse.Data,
		})
	}
}

func CreateExchange(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		ExchangeType string `json:"exchange_type"`
	}
	if err := c.ParamsParser(&request); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("CREATE_EXCHANGE %s %s", request.ExchangeName, request.ExchangeType)
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

func DeleteExchange(c *fiber.Ctx) error {
	exchangeName := c.Params("exchange")
	if exchangeName == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "exchange name is required"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "exchange name is required",
		})
	}

	command := fmt.Sprintf("DELETE_EXCHANGE %s", exchangeName)
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
