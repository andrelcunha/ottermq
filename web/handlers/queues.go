package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

func ListQueues(c *fiber.Ctx) error {
	response, err := utils.SendCommand("LIST_QUEUES")
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
		// 	"queues": commandResponse.Data,
		// })
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"queues": commandResponse.Data,
		})
	}
}

func CreateQueue(c *fiber.Ctx) error {
	var request struct {
		QueueName string `json:"queue_name"`
	}
	if err := c.ParamsParser(&request); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	command := fmt.Sprintf("CREATE_QUEUE %s", request.QueueName)
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

func DeleteQueue(c *fiber.Ctx) error {
	queueName := c.Params("queue")
	if queueName == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "queue name is required",
		})
	}

	command := fmt.Sprintf("DELETE_QUEUE %s", queueName)
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

func ConsumeMessage(c *fiber.Ctx) error {
	queueName := c.Params("queue")
	if queueName == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "queue name is required",
		})
	}
	command := fmt.Sprintf("CONSUME %s", queueName)
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

func CountMessages(c *fiber.Ctx) error {
	queueName := c.Params("queue")
	if queueName == "" {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "queue name is required",
		})
	}

	command := fmt.Sprintf("COUNT_MESSAGES %s", queueName)
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
	// c.JSON(http.StatusOK, gin.H{"data": commandResponse.Data})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": commandResponse.Data,
	})
}
