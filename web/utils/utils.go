package utils

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var conn net.Conn

func SetConn(c net.Conn) {
	conn = c
}

func SendCommand(command string) (string, error) {

	_, err := conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	return response, nil
}

func SendHeartbeat(heartbeatInterval time.Duration) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		_, err := conn.Write([]byte("HEARTBEAT\n"))
		if err != nil {
			fmt.Println("Failed to send heartbeat:", err)
			return
		}
	}

	// // Get heartbeat from server
	// buffer := make([]byte, 1024)
	// for {
	// 	n, err := conn.Read(buffer)
	// 	if err != nil {
	// 		fmt.Println("Connection closed:", err)
	// 		return
	// 	}
	// 	message := string(buffer[:n])
	// 	fmt.Println("Received:", message)
	// }
}
