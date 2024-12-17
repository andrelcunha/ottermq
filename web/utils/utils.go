package utils

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

var conn net.Conn
var commandResponse = make(chan string)

func SetConn(c net.Conn) {
	conn = c
	go listenForMessages()
}

func SendCommand(command string) (string, error) {

	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	response := <-commandResponse
	return response, nil
}

func listenForMessages() {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Failed to read message: ", err)
			return
		}

		if message == "HEARTBEAT\n" {
			log.Println("Received heartbeat")
			continue
		}

		commandResponse <- message
	}
}

func SendHeartbeat(heartbeatInterval time.Duration) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		_, err := conn.Write([]byte("HEARTBEAT\n"))
		if err != nil {
			log.Println("Failed to send heartbeat:", err)
			return
		}
	}
}
