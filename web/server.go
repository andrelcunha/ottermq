package web

import (
	"os"

	"github.com/andrelcunha/ottermq/internal/core/broker"
	_ "github.com/andrelcunha/ottermq/web/docs"
	"github.com/andrelcunha/ottermq/web/handlers/api"
	"github.com/andrelcunha/ottermq/web/handlers/api_admin"
	"github.com/andrelcunha/ottermq/web/handlers/webui"
	"github.com/andrelcunha/ottermq/web/middleware"

	"github.com/gofiber/swagger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/rabbitmq/amqp091-go"
)

type WebServer struct {
	// brokerAddr        string
	// heartbeatInterval time.Duration
	config  *Config
	Broker  *broker.Broker
	Client  *amqp091.Connection
	Channel *amqp091.Channel
}

type Config struct {
	BrokerHost string
	BrokerPort string
	Username   string
	Password   string
	JwtKey     string
}

func (ws *WebServer) Close() {
	ws.Channel.Close()
	ws.Client.Close()
}

func NewWebServer(config *Config, broker *broker.Broker, conn *amqp091.Connection) (*WebServer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &WebServer{
		config:  config,
		Broker:  broker,
		Client:  conn,
		Channel: ch,
	}, nil
}

func (ws *WebServer) SetupApp(logFile *os.File) *fiber.App {
	// connectionSting := fmt.Sprintf("amqp://%s:%s@%s:%s/", ws.config.Username, ws.config.Password, ws.config.BrokerHost, ws.config.BrokerPort)
	// conn, err := getBrokerClient(connectionSting)

	// ws.Client = conn
	app := ws.configServer(logFile)

	app.Get("/docs/*", swagger.HandlerDefault)

	// Serve static files
	app.Static("/", "./web/static")

	app.Get("/login", webui.LoginPage)
	app.Post("/login", webui.Authenticate)

	ws.AddApi(app)

	ws.AddAdminApi(app)

	ws.AddUI(app)

	return app
}

func (ws *WebServer) AddApi(app *fiber.App) {
	// Public API routes
	app.Post("/api/login", api_admin.Login)

	// Protected API routes
	apiGrp := app.Group("/api")
	apiGrp.Get("/queues", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.ListQueues(c, ws.Broker)
	})
	apiGrp.Post("/queues", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.CreateQueue(c, ws.Channel)
	})
	apiGrp.Delete("/queues/:queue", middleware.JwtMiddleware(ws.config.JwtKey), api.DeleteQueue)
	apiGrp.Post("/queues/:queue/consume", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.GetMessage(c, ws.Channel)
	})
	apiGrp.Post("/messages/:id/ack", middleware.JwtMiddleware(ws.config.JwtKey), api.AckMessage)
	apiGrp.Post("/messages", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.PublishMessage(c, ws.Channel)
	})

	apiGrp.Get("/exchanges", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.ListExchanges(c, ws.Broker)
	})
	apiGrp.Post("/exchanges", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.CreateExchange(c, ws.Broker)
	})

	apiGrp.Delete("/exchanges/:exchange", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.DeleteExchange(c, ws.Channel)
	})
	apiGrp.Get("/bindings/:exchange", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.ListBindings(c, ws.Broker)
	})
	apiGrp.Post("/bindings", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.BindQueue(c, ws.Channel)
	})
	apiGrp.Delete("/bindings", middleware.JwtMiddleware(ws.config.JwtKey), api.DeleteBinding)
	apiGrp.Get("/connections", middleware.JwtMiddleware(ws.config.JwtKey), func(c *fiber.Ctx) error {
		return api.ListConnections(c, ws.Broker)
	})
}

func (ws *WebServer) AddUI(app *fiber.App) {
	// Web Interface Routes
	webGrp := app.Group("/", middleware.AuthRequired)
	webGrp.Get("/", webui.Dashboard)
	webGrp.Get("/logout", webui.Logout)
	webGrp.Get("/overview", webui.Dashboard)
	webGrp.Get("/connections", webui.ListConnections)
	webGrp.Get("/exchanges", webui.ListExchanges)
	webGrp.Get("/queues", webui.ListQueues)
}

func (ws *WebServer) AddAdminApi(app *fiber.App) {
	// Admin API routes
	apiAdminGrp := app.Group("/api/admin")
	apiAdminGrp.Use(middleware.JwtMiddleware(ws.config.JwtKey))
	apiAdminGrp.Get("/users", api_admin.GetUsers)
	apiAdminGrp.Post("/users", api_admin.AddUser)
}

func (ws *WebServer) configServer(logFile *os.File) *fiber.App {
	engine := html.New("./web/templates", ".html")

	config := fiber.Config{

		Prefork:               false,
		AppName:               "ottermq-webadmin",
		Views:                 engine,
		ViewsLayout:           "layout",
		DisableStartupMessage: true,
	}
	app := fiber.New(config)

	// Enable CORS
	app.Use(middleware.CORSMiddleware())

	app.Use(logger.New(logger.Config{
		Output: logFile,
	}))
	return app
}
