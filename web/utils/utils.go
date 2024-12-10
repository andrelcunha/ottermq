package utils

import (
	"bufio"
	"fmt"
	"net"
)

func SendCommand(command string) (string, error) {
	conn, err := net.Dial("tcp", ":5672")
	if err != nil {
		return "", fmt.Errorf("failed to connect to broker: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	return response, nil
}
