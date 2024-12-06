package web

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/andrelcunha/ottermq/pkg/common"
	"github.com/gin-gonic/gin"
)

type WebServer struct {
	brokerAddr string
}

func NewWebServer(brokerAddr string) *WebServer {
	return &WebServer{
		brokerAddr: brokerAddr,
	}
}

func (ws *WebServer) SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/queues", ws.ListQueues)
	router.POST("/queues", ws.CreateQueue)
	router.POST("/publish", ws.PublishMessage)
	// Add more routes as needed
	return router
}

func (ws *WebServer) SendCommand(command string) (string, error) {
	conn, err := net.Dial("tcp", ":5672")
	if err != nil {
		return "", fmt.Errorf("failed to connect to broker: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	return response, nil
}

func (ws *WebServer) ListQueues(c *gin.Context) {
	response, err := ws.SendCommand("LIST_QUEUES")
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

func (ws *WebServer) CreateQueue(c *gin.Context) {
	var request struct {
		QueueName string `json:"queue_name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("CREATE_QUEUE %s", request.QueueName)
	response, err := ws.SendCommand(command)
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

func (ws *WebServer) PublishMessage(c *gin.Context) {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		RoutingKey   string `json:"routing_key"`
		Message      string `json:"message"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("PUBLISH %s %s %s", request.ExchangeName, request.RoutingKey, request.Message)
	response, err := ws.SendCommand(command)
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
