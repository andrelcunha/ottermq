package main

import (
	"log"

	"github.com/andrelcunha/ottermq/web"
)

func main() {
	log.Println("Starting OtterMq Web Admin...")

	brokerAddr := "localhost:5672"
	webServer := web.NewWebServer(brokerAddr)
	app := webServer.SetupApp()

	log.Fatal(app.Listen(":3000"))

}
