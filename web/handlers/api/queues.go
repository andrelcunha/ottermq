package api

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/broker"
	"github.com/andrelcunha/ottermq/internal/core/models"
	"github.com/rabbitmq/amqp091-go"

	"github.com/gofiber/fiber/v2"
)

// ListQueues godoc
// @Summary List all queues
// @Description Get a list of all queues
// @Tags queues
// @Accept json
// @Produce json
// @Success 200 {object} models.QueueListResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/queues [get]
func ListQueues(c *fiber.Ctx, b *broker.Broker) error {
	queues := b.ManagerApi.ListQueues()
	if queues == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to list exchanges",
		})
	}
	return c.Status(fiber.StatusOK).JSON(models.QueueListResponse{
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
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/queues [post]
func CreateQueue(c *fiber.Ctx, ch *amqp091.Channel) error {
	var request models.CreateQueueRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body",
		})
	}
	if request.QueueName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "queue name is required",
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
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Queue created successfully",
	})
}

// DeleteQueue godoc
// @Summary Delete a queue
// @Description Delete a queue with the specified name
// @Tags queues
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/queues/{queue} [delete]
func DeleteQueue(c *fiber.Ctx) error {
	panic("not implemented")
	// queueName := c.Params("queue")
	// if queueName == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
	// 		Error: "queue name is required",
	// 	})
	// }

	// command := fmt.Sprintf("DELETE_QUEUE %s", queueName)
	// response, err := utils.SendCommand(command)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
	// 		Error: err.Error(),
	// 	})
	// }

	// var commandResponse api.CommandResponse
	// if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
	// 		Error: "failed to parse response",
	// 	})
	// }

	// if commandResponse.Status == "ERROR" {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
	// 		Error: commandResponse.Message,
	// 	})
	// } else {
	// 	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
	// 		Message: commandResponse.Message,
	// 	})
	// }
}

// GetMessage godoc
// @Summary Consume a message from a queue
// @Description Consume a message from the specified queue
// @Tags queues
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/queues/{queue}/consume [post]
func GetMessage(c *fiber.Ctx, ch *amqp091.Channel) error {
	queueName := c.Params("queue")
	if queueName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "queue name is required",
		})
	}
	fmt.Println("Consuming from queue:", queueName)
	msg, ok, err := ch.Get(queueName, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "no messages in queue",
		})
	}
	fmt.Println("Message received: ", string(msg.Body))
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: string(msg.Body),
	})
}
