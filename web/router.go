package web

import (
	"github.com/andrelcunha/ottermq/web/handlers"
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
	router.GET("/queues", handlers.ListQueues)
	router.POST("/queues", handlers.CreateQueue)
	router.DELETE("/queues/:queue", handlers.DeleteQueue)
	router.POST("/queues/:queue/consume", handlers.ConsumeMessage)
	router.GET("/queues/:queue/count", handlers.CountMessages)

	router.POST("/messages/:id/ack", handlers.AckMessage)
	router.POST("/messages", handlers.PublishMessage)

	router.GET("/exchanges", handlers.ListExchanges)
	router.POST("/exchanges", handlers.CreateExchange)

	router.GET("/bindings/:exchange", handlers.ListBindings)
	router.POST("/bindings", handlers.BindQueue)
	router.DELETE("/bindings", handlers.DeleteBinding)

	// Add more routes as needed
	return router
}
