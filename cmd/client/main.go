package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/andrelcunha/ottermq/pkg/connection/client"

	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

var (
	version = "0.6.0-alpha"
)

const (
	PORT      = "5673"
	HOST      = "localhost"
	HEARTBEAT = 10
	USERNAME  = "guest"
	PASSWORD  = "guest"
	VHOST     = "/"
)

func main() {
	config := &shared.ClientConfig{
		Host:              HOST,
		Port:              PORT,
		Username:          USERNAME,
		Password:          PASSWORD,
		Vhost:             VHOST,
		HeartbeatInterval: HEARTBEAT,
	}
	client := client.NewClient(config)
	go func() {
		if err := client.Dial(HOST, PORT); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	}()

	channel, err := client.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	log.Printf("Channel  %d opened successfully\n", channel.Id)

	// if err := channel.Close(); err != nil {
	// 	log.Fatalf("Failed to close channel: %v", err)
	// }
	// log.Printf("Channel %d closed successfully\n", channel.Id)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down OtterMq...")
	client.Shutdown()
}
