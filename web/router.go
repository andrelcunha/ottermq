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
	router.GET("/exchanges", ws.ListExchanges)
	router.POST("/exchanges", ws.CreateExchange)
	router.POST("/bindings/list", ws.ListBindings)
	router.POST("/bindings", ws.BindQueue)
	router.POST("/consume", ws.ConsumeMessage)
	router.POST("/ack", ws.AckMessage)
	router.POST("/delete", ws.DeleteQueue)
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
			"data": commandResponse.Data,
		})
	}
}

func (ws *WebServer) CreateExchange(c *gin.Context) {
	var request struct {
		ExchangeName string `json:"exchange_name"`
		ExchangeType string `json:"exchange_type"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("CREATE_EXCHANGE %s %s", request.ExchangeName, request.ExchangeType)
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

func (ws *WebServer) ListExchanges(c *gin.Context) {
	response, err := ws.SendCommand("LIST_EXCHANGES")
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

func (ws *WebServer) BindQueue(c *gin.Context) {
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

func (ws *WebServer) ListBindings(c *gin.Context) {
	var request struct {
		ExchangeName string `json:"exchange_name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("LIST_BINDINGS %s", request.ExchangeName)
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
			"bindings": commandResponse.Data,
		})
	}
}

func (ws *WebServer) ConsumeMessage(c *gin.Context) {
	var request struct {
		QueueName string `json:"queue_name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("CONSUME %s", request.QueueName)
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
			"data": commandResponse.Data,
		})
	}
}

func (ws *WebServer) AckMessage(c *gin.Context) {
	var request struct {
		MsgID string `json:"msg_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("ACK %s", request.MsgID)
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

func (ws *WebServer) DeleteQueue(c *gin.Context) {
	var request struct {
		QueueName string `json:"queue_name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	command := fmt.Sprintf("DELETE_QUEUE %s", request.QueueName)
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
