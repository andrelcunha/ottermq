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

	brokerAddr := "localhost:5672"
	webServer := web.NewWebServer(brokerAddr)
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
