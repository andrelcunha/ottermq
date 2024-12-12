package web

import (
	"github.com/andrelcunha/ottermq/web/handlers"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
)

type WebServer struct {
	brokerAddr string
}

func NewWebServer(brokerAddr string) *WebServer {
	return &WebServer{
		brokerAddr: brokerAddr,
	}
}

func (ws *WebServer) SetupApp() *fiber.App {
	config := fiber.Config{
		Prefork:               false,
		AppName:               "ottermq-webadmin",
		DisableStartupMessage: true,
	}
	app := fiber.New(config)

	// Enable CORS
	app.Use(utils.CORSMiddleware())

	api := app.Group("/api")

	api.Get("/queues", handlers.ListQueues)
	api.Post("/queues", handlers.CreateQueue)
	api.Delete("/queues/:queue", handlers.DeleteQueue)
	api.Post("/queues/:queue/consume", handlers.ConsumeMessage)
	api.Get("/queues/:queue/count", handlers.CountMessages)

	api.Post("/messages/:id/ack", handlers.AckMessage)
	api.Post("/messages", handlers.PublishMessage)

	api.Get("/exchanges", handlers.ListExchanges)
	api.Post("/exchanges", handlers.CreateExchange)
	api.Delete("/exchanges/:exchange", handlers.DeleteExchange)

	api.Get("/bindings/:exchange", handlers.ListBindings)
	api.Post("/bindings", handlers.BindQueue)
	api.Delete("/bindings", handlers.DeleteBinding)

	// Add more routes as needed
	return app
}
