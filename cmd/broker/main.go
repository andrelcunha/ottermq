package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/broker"
)

func main() {
	config := &config.Config{
		Port:              "5672",
		Host:              "localhost",
		HeartBeatInterval: 600,
	}
	b := broker.NewBroker(config)

	log.Println("OtterMq is starting...")
	go b.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down OtterMq...")
	b.Shutdown()

	log.Println("Server gracefully stopped.")
}
