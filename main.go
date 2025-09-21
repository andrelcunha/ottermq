package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	// Determine the directory of the running binary
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	executableDir := filepath.Dir(executablePath)
	dataDir := filepath.Join(executableDir, "data")

	// Ensure the data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Println("Data directory not found. Creating a new one...")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
	}

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
	log.Println("OtterMQ version ", version)
	log.Println("Broker is starting...")

	// Verify if the database file exists
	dbPath := filepath.Join(dataDir, "ottermq.db")
	persistdb.SetDbPath(dbPath)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
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
	// open "server.log" for appending
	logfile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	app := webServer.SetupApp(logfile) // Using os.Stdout for logging

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
