package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

// BindQueue godoc
// @Summary Bind a queue to an exchange
// @Description Bind a queue to an exchange with the specified routing key
// @Tags bindings
// @Accept json
// @Produce json
// @Param binding body models.BindQueueRequest true "Binding details"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/bindings [post]
func BindQueue(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		QueueName    string `json:"queue_name"`
		RoutingKey   string `json:"routing_key"`
	}
	if err := c.BodyParser(&request); err != nil {
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

// ListBindings godoc
// @Summary List all bindings for an exchange
// @Description Get a list of all bindings for the specified exchange
// @Tags bindings
// @Accept json
// @Produce json
// @Param exchange path string true "Exchange name"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/bindings/{exchange} [get]
func ListBindings(c *fiber.Ctx) error {

	exchangeName := c.Params("exchange")
	if exchangeName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Exchange name is required",
		})
	}

	command := fmt.Sprintf("LIST_BINDINGS %s", exchangeName)
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
			"bindings": commandResponse.Data,
		})
	}
}

// DeleteBinding godoc
// @Summary Delete a binding
// @Description Delete a binding from an exchange to a queue
// @Tags bindings
// @Accept json
// @Produce json
// @Param binding body models.DeleteBindingRequest true "Binding to delete"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/bindings [delete]
func DeleteBinding(c *fiber.Ctx) error {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		QueueName    string `json:"queue_name"`
		RoutingKey   string `json:"routing_key"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	command := fmt.Sprintf("DELETE_BINDING %s %s %s",
		request.ExchangeName,
		request.QueueName,
		request.RoutingKey)
	response, err := utils.SendCommand(command)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse response"})
	}

	if commandResponse.Status == "ERROR" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": commandResponse.Message})
	} else {
		return c.Status(http.StatusOK).JSON(fiber.Map{"message": commandResponse.Message})
	}
}
