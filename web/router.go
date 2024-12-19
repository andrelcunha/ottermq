package web

import (
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/andrelcunha/ottermq/docs"
	"github.com/andrelcunha/ottermq/web/handlers/api"
	"github.com/andrelcunha/ottermq/web/handlers/api_admin"
	"github.com/andrelcunha/ottermq/web/handlers/webui"
	"github.com/andrelcunha/ottermq/web/middleware"

	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/html/v2"
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
	conn, err := net.Dial("tcp", brokerAddr)
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

	app.Use(logger.New(logger.Config{
		// Output: logFile,
	}))

	// Pass the connection to the utils package
	utils.SetConn(ws.conn)
	auth_command := fmt.Sprintf("AUTH %s %s",
		ws.config.Username,
		ws.config.Password,
	)
	utils.SendCommand(auth_command)

	// set heartbeat
	go utils.SendHeartbeat(ws.heartbeatInterval)

	app.Get("/login", webui.LoginPage)
	app.Post("/login", webui.Authenticate)

	// API routes
	apiGrp := app.Group("/api")
	apiGrp.Get("/swagger/*", swagger.HandlerDefault)

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

	apiAdminGrp := apiGrp.Group("/admin")
	apiAdminGrp.Get("/users", api_admin.GetUsers)
	apiAdminGrp.Post("/users", api_admin.AddUser)

	// Web Interface Routes
	webGrp := app.Group("/", middleware.AuthRequired)
	webGrp.Get("/", webui.Dashboard)
	webGrp.Get("/logout", webui.Logout)
	webGrp.Get("/overview", webui.Dashboard)
	webGrp.Get("/connections", webui.ListConnections)
	webGrp.Get("/exchanges", webui.ListExchanges)
	webGrp.Get("/queues", webui.ListQueues)
	// webGrp.Get("/settings", webui.Settings)

	// Serve static files
	app.Static("/", "./web/static")

	return app
}
