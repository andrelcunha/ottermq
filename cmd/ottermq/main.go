package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/broker"
	"github.com/andrelcunha/ottermq/internal/core/persistdb"
	"github.com/andrelcunha/ottermq/web"
)

var (
	VERSION = ""
)

const (
	PORT      = "5672"
	HOST      = ""
	HEARTBEAT = 60
	USERNAME  = "guest"
	PASSWORD  = "guest"
	VHOST     = "/"
)

// @title OtterMQ API
// @version 1.0
// @description API documentation for OtterMQ broker
// @host localhost:3000
// @BasePath /api/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
		Ssl:                  false,
		Version:              VERSION,
	}

	ctx, cancel := context.WithCancel(context.Background())
	b := broker.NewBroker(config, ctx, cancel)

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
		err := b.Start()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
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
		err := app.Listen(":3000")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}()

	// Handle OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down OtterMq...")
	cancel()
	b.ShuttingDown.Store(true)

	// Broadcast connection close to all channels
	b.BroadcastConnectionClose()
	log.Println("Waiting for active connections to close...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		b.ActiveConns.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All connections closed gracefully.")
	case <-shutdownCtx.Done():
		log.Println("Timeout reached. Forcing shutdown.")
	}

	b.Shutdown()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown web server: %v", err)
	}
	log.Println("Server gracefully stopped.")
	os.Exit(0) // if came so far it means the server has stopped gracefully
}
