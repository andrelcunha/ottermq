package web

import (
	"net"
	"os"
	"time"

	"github.com/andrelcunha/ottermq/web/handlers/api"
	"github.com/andrelcunha/ottermq/web/handlers/api_admin"
	"github.com/andrelcunha/ottermq/web/handlers/webui"
	"github.com/andrelcunha/ottermq/web/middleware"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
)

type WebServer struct {
	brokerAddr        string
	conn              net.Conn
	heartbeatInterval time.Duration
	config            *Config
}

type Config struct {
	BrokerHost        string
	BrokerPort        string
	HeartbeatInterval int
	Username          string
	Password          string
}

func (ws *WebServer) Close() {
	ws.conn.Close()
}

func NewWebServer(config *Config) (*WebServer, error) {
	brokerAddr := config.BrokerHost + ":" + config.BrokerPort
	conn, err := getConnection(brokerAddr)
	if err != nil {
		return nil, err
	}
	return &WebServer{
		brokerAddr:        brokerAddr,
		conn:              conn,
		heartbeatInterval: time.Duration(config.HeartbeatInterval) * time.Second,
		config:            config,
	}, nil
}

func (ws *WebServer) SetupApp(logFile *os.File) *fiber.App {
	ws.configBrokerClient()
	app := ws.configServer()

	ws.AddSwagger(app)

	// Serve static files
	app.Static("/", "./web/static")

	app.Get("/login", webui.LoginPage)
	app.Post("/login", webui.Authenticate)

	ws.AddApi(app)

	// Admin API routes
	apiAdminGrp := app.Group("/api/admin")
	apiAdminGrp.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte("secret")},
	}))
	apiAdminGrp.Get("/users", api_admin.GetUsers)
	apiAdminGrp.Post("/users", api_admin.AddUser)

	ws.AddUI(app)

	return app
}

func (ws *WebServer) AddApi(app *fiber.App) {
	// API routes
	apiGrp := app.Group("/api/")
	apiGrp.Post("/authenticate", api.Authenticate)
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
	apiGrp.Get("/connections", api.ListConnections)
	apiGrp.Post("/login", api_admin.Login)
}

func (ws *WebServer) AddSwagger(app *fiber.App) {
	swaggerCfg := swagger.Config{
		BasePath: "/api/",
		FilePath: "./web/static/docs/swagger.json",
		Path:     "docs",
		Title:    "OtterMQ API",
	}
	swaggerHandler := swagger.New(swaggerCfg)
	app.Use(swaggerHandler)
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
	// webGrp.Get("/settings", webui.Settings)
}
