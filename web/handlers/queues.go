package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gin-gonic/gin"
)

func ListQueues(c *gin.Context) {
	response, err := utils.SendCommand("LIST_QUEUES")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	if commandResponse.Status == "ERROR" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"queues": commandResponse.Data,
		})
	}
}

func CreateQueue(c *gin.Context) {
	var request struct {
		QueueName string `json:"queue_name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("CREATE_QUEUE %s", request.QueueName)
	response, err := utils.SendCommand(command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	if commandResponse.Status == "ERROR" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": commandResponse.Message,
		})
	}
}

func DeleteQueue(c *gin.Context) {
	queueName := c.Param("queue")
	if queueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return
	}

	command := fmt.Sprintf("DELETE_QUEUE %s", queueName)
	response, err := utils.SendCommand(command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	if commandResponse.Status == "ERROR" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": commandResponse.Message,
		})
	}
}

func ConsumeMessage(c *gin.Context) {
	queueName := c.Param("queue")
	if queueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return
	}
	command := fmt.Sprintf("CONSUME %s", queueName)
	response, err := utils.SendCommand(command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	if commandResponse.Status == "ERROR" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": commandResponse.Message})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": commandResponse.Data,
		})
	}
}

func CountMessages(c *gin.Context) {
	queueName := c.Param("queue")
	if queueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return
	}

	command := fmt.Sprintf("COUNT_MESSAGES %s", queueName)
	response, err := utils.SendCommand(command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var commandResponse common.CommandResponse
	if err := json.Unmarshal([]byte(response), &commandResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": commandResponse.Data})
}
