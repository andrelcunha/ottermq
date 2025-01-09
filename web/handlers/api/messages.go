package api

import (
	// "encoding/json"
	// "fmt"

	// "github.com/andrelcunha/ottermq/pkg/common/communication/api"
	// "github.com/andrelcunha/ottermq/web/models"
	// "github.com/andrelcunha/ottermq/web/utils"

	"github.com/andrelcunha/ottermq/web/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rabbitmq/amqp091-go"
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
func PublishMessage(c *fiber.Ctx, ch *amqp091.Channel) error {
	var request models.PublishMessageRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	msg := amqp091.Publishing{
		ContentType: "text/plain",
		Body:        []byte(request.Message),
	}
	err := ch.Publish(
		request.ExchangeName,
		request.RoutingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Message published",
	})
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
	// msgID := c.Params("id")
	// if msgID == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "message ID is required",
	// 	})
	// }
	// command := fmt.Sprintf("ACK %s", msgID)
	// response, err := utils.SendCommand(command)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": err.Error(),
	// 	})
	// }

	// var commandResponse api.CommandResponse
	// if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": "failed to parse response",
	// 	})
	// }

	// if commandResponse.Status == "ERROR" {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": commandResponse.Message,
	// 	})
	// } else {
	// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 		"message": commandResponse.Message,
	// 	})
	// }
	return nil // just to make the compiler happy
}
