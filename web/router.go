package web

import (
	"log"
	"os"

	_ "github.com/andrelcunha/ottermq/docs"
	"github.com/andrelcunha/ottermq/web/handlers/api"
	"github.com/andrelcunha/ottermq/web/handlers/webui"

	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/html/v2"
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
	engine := html.New("./web/templates", ".html")

	config := fiber.Config{
		Prefork:               false,
		AppName:               "ottermq-webadmin",
		Views:                 engine,
		ViewsLayout:           "layout",
		DisableStartupMessage: false,
	}
	app := fiber.New(config)

	// Enable CORS
	app.Use(utils.CORSMiddleware())

	// Logger
	file, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()
	app.Use(logger.New(logger.Config{
		Output: file,
	}))

	// API routes
	apiGrp := app.Group("/api")
	apiGrp.Get("/swagger/*", swagger.HandlerDefault)
	apiGrp.Get("/queues", api.ListQueues)
	apiGrp.Post("/queues", api.CreateQueue)
	apiGrp.Delete("/queues/:queue", api.DeleteQueue)
	apiGrp.Post("/queues/:queue/consume", api.ConsumeMessage)
	apiGrp.Get("/queues/:queue/count", api.CountMessages)
	apiGrp.Post("/messages/:id/ack", api.AckMessage)
	apiGrp.Post("/messages", api.PublishMessage)
	apiGrp.Get("/exchanges", api.ListExchanges)
	apiGrp.Post("/exchanges", api.CreateExchange)
	apiGrp.Delete("/exchanges/:exchange", api.DeleteExchange)
	apiGrp.Get("/bindings/:exchange", api.ListBindings)
	apiGrp.Post("/bindings", api.BindQueue)
	apiGrp.Delete("/bindings", api.DeleteBinding)

	// Web Interface Routes
	webGrp := app.Group("/")
	webGrp.Get("/", webui.Dashboard)
	// webGrp.Get("/queues", webui.ListQueues)
	// webGrp.Get("/exchanges", webui.ListExchanges)
	// webGrp.Get("/bindings", webui.ListBindings)
	// webGrp.Get("/settings", webui.Settings)

	// Serve static files
	app.Static("/static", "./static")

	return app
}
