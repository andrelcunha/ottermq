package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gin-gonic/gin"
)

func ListExchanges(c *gin.Context) {
	response, err := utils.SendCommand("LIST_EXCHANGES")
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
			"exchanges": commandResponse.Data,
		})
	}
}

func CreateExchange(c *gin.Context) {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		ExchangeType string `json:"exchange_type"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("CREATE_EXCHANGE %s %s", request.ExchangeName, request.ExchangeType)
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
