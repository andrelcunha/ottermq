package api

import (
	"github.com/andrelcunha/ottermq/internal/core/broker"
	dtos "github.com/andrelcunha/ottermq/internal/core/models"
	"github.com/andrelcunha/ottermq/web/models"

	"github.com/gofiber/fiber/v2"
	"github.com/rabbitmq/amqp091-go"
)

// ListExchanges godoc
// @Summary List all exchanges
// @Description Get a list of all exchanges
// @Tags exchanges
// @Accept json
// @Produce json
// @Success 200 {object} dtos.ExchangeListResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /api/exchanges [get]
func ListExchanges(c *fiber.Ctx, b *broker.Broker) error {
	exchanges := b.ManagerApi.ListExchanges()
	if exchanges == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{
			Error: "failed to list exchanges",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dtos.ExchangeListResponse{
		Exchanges: exchanges,
	})
}

// CreateExchange godoc
// @Summary Create a new exchange
// @Description Create a new exchange with the specified name and type
// @Tags exchanges
// @Accept json
// @Produce json
// @Param exchange body models.CreateExchangeRequest true "Exchange to create"
// @Success 200 {object} dtos.SuccessResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /api/exchanges [post]
func CreateExchange(c *fiber.Ctx, b *broker.Broker) error {
	// func CreateExchange(c *fiber.Ctx, ch *amqp091.Channel) error {
	// var request models.CreateExchangeRequest
	// if err := c.BodyParser(&request); err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": err.Error(),
	// 	})
	// }

	// err := ch.ExchangeDeclare(
	// 	request.ExchangeName,
	// 	request.ExchangeType,
	// 	true,  // durable
	// 	false, // auto-deleted
	// 	false, // internal
	// 	false, // no-wait
	// 	nil,   // arguments
	// )
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": err.Error(),
	// 	})
	// }
	// return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 	"message": "Exchange created successfully",
	// })
	// verify if this exchange already exists

	var request models.CreateExchangeRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error: "failed to create exchange: " + err.Error(),
		})
	}
	exchangeDto := dtos.ExchangeDTO{
		VHostName: "/", // Assuming a default vhost for simplicity
		Name:      request.ExchangeName,
		Type:      request.ExchangeType,
	}

	err := b.ManagerApi.CreateExchange(exchangeDto)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{
			Error: "failed to create exchange: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Message: "Exchange created successfully",
	})
}

// DeleteExchange godoc
// @Summary Delete an exchange
// @Description Delete an exchange with the specified name
// @Tags exchanges
// @Accept json
// @Produce json
// @Param exchange path string true "Exchange name"
// @Success 200 {object} dtos.SuccessResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /api/exchanges/{exchange} [delete]
func DeleteExchange(c *fiber.Ctx, ch *amqp091.Channel) error {
	exchangeName := c.Params("exchange")
	if exchangeName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error: "Exchange name is required",
		})
	}

	err := ch.ExchangeDelete(exchangeName, false, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{
			Error: "failed to delete exchange: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Message: "Exchange deleted successfully",
	})
}
