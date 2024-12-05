package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/andrelcunha/ottermq/internal/broker"
)

func main() {
	log.Println("OtterMq is starting...")
	b := broker.NewBroker()

	go b.Start(":5672")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down OtterMq...")

	// Perform any necessary cleanup or shutdown tasks here
	// Example: b.Shytdown()

	log.Println("Server gracefully stopped.")
}
