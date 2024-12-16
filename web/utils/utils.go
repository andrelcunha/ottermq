package utils

import (
	"bufio"
	"fmt"
	"net"
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
