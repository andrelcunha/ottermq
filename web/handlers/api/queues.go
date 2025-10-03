package api

import (
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
// @Failure 401 {object} models.UnauthorizedErrorResponse "Missing or invalid JWT token"
// @Failure 500 {object} models.ErrorResponse "Failed to list queues"
// @Router /queues [get]
// @Security BearerAuth
func ListQueues(c *fiber.Ctx, b *broker.Broker) error {
	queues, err := b.ManagerApi.ListQueues()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list queues: " + err.Error(),
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
// @Failure 401 {object} models.UnauthorizedErrorResponse "Missing or invalid JWT token"
// @Failure 500 {object} models.ErrorResponse
// @Router /queues [post]
// @Security BearerAuth
func CreateQueue(c *fiber.Ctx, ch *amqp091.Channel) error {
	var request models.CreateQueueRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
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
// @Success 204 {object} nil
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.UnauthorizedErrorResponse "Missing or invalid JWT token"
// @Failure 500 {object} models.ErrorResponse
// @Router /queues/{queue} [delete]
// @Security BearerAuth
func DeleteQueue(c *fiber.Ctx, b *broker.Broker) error {
	queueName := c.Params("queue")
	if queueName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "queue name is required",
		})
	}

	// Use default vhost
	err := b.ManagerApi.DeleteQueue(broker.DefaultVHost, queueName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetMessage godoc
// @Summary Consume a message from a queue
// @Description Consume a message from the specified queue
// @Tags queues
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse "Queue name is required"
// @Failure 401 {object} models.UnauthorizedErrorResponse "Missing or invalid JWT token"
// @Failure 404 {object} models.ErrorResponse "No messages in queue"
// @Failure 500 {object} models.ErrorResponse
// @Router /queues/{queue}/consume [post]
// @Security BearerAuth
func GetMessage(c *fiber.Ctx, ch *amqp091.Channel) error {
	queueName := c.Params("queue")
	if queueName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Queue name is required",
		})
	}
	msg, ok, err := ch.Get(queueName, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "No messages in queue",
		})
	}
	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: string(msg.Body),
	})
}
