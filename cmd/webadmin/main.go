package main

import (
	"log"

	"github.com/andrelcunha/ottermq/web"
)

func main() {
	log.Println("Starting OtterMq Web Admin...")

	brokerAddr := "localhost:5672"
	webServer := web.NewWebServer(brokerAddr)
	router := webServer.SetupRouter()

	err := router.Run(":8081")
	if err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
