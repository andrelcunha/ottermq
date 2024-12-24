package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/broker"
	"github.com/andrelcunha/ottermq/pkg/persistdb"
)

var (
	version = "0.6.0-alpha"
)

func main() {
	config := &config.Config{
		Port:                 "5672",
		Host:                 "localhost",
		Username:             "guest",
		Password:             "guest",
		HeartbeatIntervalMax: 600,
		ChannelMax:           100,
		FrameMax:             131072,
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
	go b.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down OtterMq...")
	b.Shutdown()

	log.Println("Server gracefully stopped.")
}
