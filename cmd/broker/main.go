package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/andrelcunha/ottermq/config"
	broker "github.com/andrelcunha/ottermq/internal/core"
	"github.com/andrelcunha/ottermq/pkg/persistdb"
)

var (
	version = "0.6.0-alpha"
)

func main() {
	config := &config.Config{
		Port:              "5672",
		Host:              "localhost",
		HeartBeatInterval: 600,
		Username:          "guest",
		Password:          "guest",
	}
	b := broker.NewBroker(config)

	log.Println("OtterMq is starting...")

	// verify if the database file exists
	if _, err := os.Stat("./data/ottermq.db"); os.IsNotExist(err) {
		log.Println("Database file not found. Creating a new one...")
		persistdb.InitDB()
		persistdb.AddDefaultRoles()
		persistdb.AddDefaultPermissions()
		user := persistdb.UserCreateDTO{Username: config.Username, Password: config.Password, RoleID: 1}
		persistdb.AddUser(user)
		persistdb.CloseDB()
	}
	go b.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down OtterMq...")
	b.Shutdown()

	log.Println("Server gracefully stopped.")
}
