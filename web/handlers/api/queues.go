package api

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/broker"
	dtos "github.com/andrelcunha/ottermq/internal/core/models"
	"github.com/andrelcunha/ottermq/web/models"
	"github.com/rabbitmq/amqp091-go"

	"github.com/gofiber/fiber/v2"
)

// ListQueues godoc
// @Summary List all queues
// @Description Get a list of all queues
// @Tags queues
// @Accept json
// @Produce json
// @Success 200 {object} dto.QueueListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/queues [get]
func ListQueues(c *fiber.Ctx, b *broker.Broker) error {
	queues := b.ManagerApi.ListQueues()
	if queues == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{
			Error: "failed to list exchanges",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dtos.QueueListResponse{
		Queues: queues,
	})
}

// CreateQueue godoc
// @Summary Create a new queue
// @Description Create a new queue with the specified name
// @Tags queues
// @Accept json
// @Produce json
// @Param queue body models.CreateQueueRequest true "Queue to create"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/queues [post]
func CreateQueue(c *fiber.Ctx, ch *amqp091.Channel) error {
	var request models.CreateQueueRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if request.QueueName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "queue name is required",
		})
	}

	_, err := ch.QueueDeclare(
		request.QueueName,
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Queue created successfully",
	})
}

// DeleteQueue godoc
// @Summary Delete a queue
// @Description Delete a queue with the specified name
// @Tags queues
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/queues/{queue} [delete]
func DeleteQueue(c *fiber.Ctx) error {
	// queueName := c.Params("queue")
	// if queueName == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "queue name is required",
	// 	})
	// }

	// command := fmt.Sprintf("DELETE_QUEUE %s", queueName)
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
	return nil // just to make it compile
}

// GetMessage godoc
// @Summary Consume a message from a queue
// @Description Consume a message from the specified queue
// @Tags queues
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/queues/{queue}/consume [post]
func GetMessage(c *fiber.Ctx, ch *amqp091.Channel) error {
	queueName := c.Params("queue")
	if queueName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "queue name is required",
		})
	}
	fmt.Println("Consuming from queue:", queueName)
	msg, ok, err := ch.Get(queueName, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"data": "",
		})
	}
	fmt.Println("Message received: ", string(msg.Body))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": string(msg.Body),
	})
}
