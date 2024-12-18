package api

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/models"
	_ "github.com/andrelcunha/ottermq/web/models"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

// ListExchanges godoc
// @Summary List all exchanges
// @Description Get a list of all exchanges
// @Tags exchanges
// @Accept json
// @Produce json
// @Success 200 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/exchanges [get]
func ListExchanges(c *fiber.Ctx) error {
	response, err := utils.SendCommand("LIST_EXCHANGES")
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
			"exchanges": commandResponse.Data,
		})
	}
}

// CreateExchange godoc
// @Summary Create a new exchange
// @Description Create a new exchange with the specified name and type
// @Tags exchanges
// @Accept json
// @Produce json
// @Param exchange body models.CreateExchangeRequest true "Exchange to create"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/exchanges [post]
func CreateExchange(c *fiber.Ctx) error {
	var request models.CreateExchangeRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("CREATE_EXCHANGE %s %s", request.ExchangeName, request.ExchangeType)
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

// DeleteExchange godoc
// @Summary Delete an exchange
// @Description Delete an exchange with the specified name
// @Tags exchanges
// @Accept json
// @Produce json
// @Param exchange path string true "Exchange name"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/exchanges/{exchange} [delete]
func DeleteExchange(c *fiber.Ctx) error {
	exchangeName := c.Params("exchange")
	if exchangeName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "exchange name is required",
		})
	}

	command := fmt.Sprintf("DELETE_EXCHANGE %s", exchangeName)
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
