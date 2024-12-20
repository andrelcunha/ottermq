package web

import (
	"fmt"

	"github.com/andrelcunha/ottermq/web/middleware"
	"github.com/andrelcunha/ottermq/web/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

func (ws *WebServer) configServer() *fiber.App {
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
		// Output: logFile,
	}))
	return app
}

func (ws *WebServer) configBrokerClient() {
	// Pass the connection to the utils package
	utils.SetConn(ws.conn)
	auth_command := fmt.Sprintf("AUTH %s %s",
		ws.config.Username,
		ws.config.Password,
	)
	utils.SendCommand(auth_command)

	// set heartbeat
	go utils.SendHeartbeat(ws.heartbeatInterval)
}
