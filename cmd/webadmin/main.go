package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrelcunha/ottermq/web"
)

func main() {
	// Logger
	file, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	log.Println("Starting OtterMq Web Admin...")

	config := web.Config{
		BrokerHost:        "localhost",
		BrokerPort:        "5672",
		HeartbeatInterval: 600,
	}

	webServer, err := web.NewWebServer(&config)
	if err != nil {
		log.Fatalf("failed to connect to broker: %v", err)
	}
	defer webServer.Close()
	app := webServer.SetupApp(file)

	//Handle gracefull shutdown
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-channel
		file.Close()
		app.Shutdown()
	}()

	app.Listen(":3000")
}
