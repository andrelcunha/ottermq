package main

// import "github.com/andrelcunha/ottermq/cmd/cli/client"
import (
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/client"
)

func main() {
	// client.Execute()
	conn, err := net.Dial("tcp", "localhost:5672")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	handler := client.NewProtocolHandler(conn)
	handler.ConnectionHandshake()
}
