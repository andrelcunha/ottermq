package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/broker"
	"github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/andrelcunha/ottermq/web"
)

var (
	version = "0.6.0-alpha"
)

const (
	PORT      = "5672"
	HOST      = "localhost"
	HEARTBEAT = 10
	USERNAME  = "guest"
	PASSWORD  = "guest"
	VHOST     = "/"
)

func main() {
	config := &config.Config{
		Port:                 PORT,
		Host:                 HOST,
		Username:             USERNAME,
		Password:             PASSWORD,
		HeartbeatIntervalMax: HEARTBEAT,
		ChannelMax:           5,
		FrameMax:             131072,
	}

	b := broker.NewBroker(config)

	log.Println("OtterMq is starting...")

	// Verify if the database file exists
	if _, err := os.Stat("./data/ottermq.db"); os.IsNotExist(err) {
		log.Println("Database file not found. Creating a new one...")
		persistdb.InitDB()
		persistdb.AddDefaultRoles()
		persistdb.AddDefaultPermissions()
		user := persistdb.UserCreateDTO{Username: config.Username, Password: config.Password, RoleID: 1}
		persistdb.AddUser(user)
		persistdb.CloseDB()
	}
	persistdb.OpenDB()
	user, err := persistdb.GetUserByUsername(config.Username)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	if user.RoleID != 1 {
		log.Fatalf("User is not an admin")
	}
	persistdb.CloseDB()
	b.VHosts["/"].Users[user.Username] = &user

	// Start the broker in a goroutine
	go func() {
		b.Start()
	}()

	// Initialize the web admin server
	webConfig := &web.Config{
		BrokerHost: HOST,
		BrokerPort: PORT,
		// HeartbeatInterval: 60,
		Username: USERNAME,
		Password: PASSWORD,
		JwtKey:   "secret",
	}
	conn, err := web.GetBrokerClient(webConfig)
	if err != nil {
		log.Fatalf("failed to connect to broker: %v", err)
	}
	defer conn.Close()

	webServer, err := web.NewWebServer(webConfig, b, conn)
	if err != nil {
		log.Fatalf("failed to connect to broker: %v", err)
	}
	defer webServer.Close()
	app := webServer.SetupApp(os.Stdout) // Using os.Stdout for logging

	// Start the web admin server in a goroutine
	go func() {
		log.Fatal(app.Listen(":3000"))
	}()

	// Handle OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down OtterMq...")
	b.Shutdown()
	app.Shutdown()
	log.Println("Server gracefully stopped.")
}
