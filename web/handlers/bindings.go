package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gin-gonic/gin"
)

func BindQueue(c *gin.Context) {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		QueueName    string `json:"queue_name"`
		RoutingKey   string `json:"routing_key"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("BIND_QUEUE %s %s %s",
		request.ExchangeName,
		request.QueueName,
		request.RoutingKey)
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

func ListBindings(c *gin.Context) {
	exchangeName := c.Param("exchange")
	if exchangeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exchange name is required"})
		return
	}

	command := fmt.Sprintf("LIST_BINDINGS %s", exchangeName)
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
			"bindings": commandResponse.Data,
		})
	}
}

func DeleteBinding(c *gin.Context) {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		QueueName    string `json:"queue_name"`
		RoutingKey   string `json:"routing_key"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("DELETE_BINDING %s %s %s",
		request.ExchangeName,
		request.QueueName,
		request.RoutingKey)
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
