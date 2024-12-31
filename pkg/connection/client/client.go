package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

var (
	version = "0.6.0-alpha"
)

const (
	platform = "golang"
	product  = "OtterMQ Client"
)

type Client struct {
	conn   net.Conn
	config *shared.ClientConfig
}

func NewClient(config *shared.ClientConfig) *Client {
	return &Client{
		config: config,
	}
}

func (c *Client) Dial(host, port string) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("ERROR: %s\n", err.Error())
	}
	c.conn = conn
	defer conn.Close()

	capabilities := map[string]interface{}{
		"basic.nack":             true,
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"publisher_confirms":     true,
	}

	clientProperties := map[string]interface{}{
		"capabilities": capabilities,
		"product":      product,
		"version":      version,
		"platform":     platform,
	}
	configurations := map[string]interface{}{
		"username":         c.config.Username,
		"password":         c.config.Password,
		"vhost":            c.config.Vhost,
		"clientProperties": clientProperties,
		"mechanism":        "PLAIN",
		"locale":           "en_US",
	}
	if err := ClientHandshake(&configurations, conn); err != nil {
		return err
	}

	go c.sendHeartbeats()

	go c.receiveLoop(conn)
	// return nil
	select {}
}

func (c *Client) receiveLoop(conn net.Conn) {
	fmt.Println("Starting receive loop")
	for {
		frame, err := shared.ReadFrame(conn)
		fmt.Printf("Received frame: %x\n", frame)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Connection timeout: %v", err)
			}
			if err == io.EOF {
				log.Printf("Connection closed by server: %v", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading frame: %v", err)
			return
		}
		log.Printf("[DEBUG] Received frame: %x", frame)

		_, err = shared.ParseFrame(nil, frame)
		if err != nil {
			log.Printf("Error parsing frame: %v", err)
			// return
		}
	}
}

// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func ClientHandshake(configurations *map[string]interface{}, conn net.Conn) error {
	log.Printf("Starting connection handshake")
	// Send Protocol Header
	if err := shared.SendProtocolHeader(conn); err != nil {
		log.Fatalf("Failed to send protocol header: %v", err)
	}

	/** connection.start **/
	// Read connection.start frame
	frame, err := shared.ReadFrame(conn)
	if err != nil {
		err = fmt.Errorf("Failed to read connection.start frame: %v", err)
		return err
	}
	if isAMQPVersionResponse(frame) {
		log.Printf("Received AMQP version response: %s", frame)
		return nil
	}

	//TODO: Refactor - remove processing from ParseFrame
	// This processing is concerned to the client only.
	untypedResp, err := shared.ParseFrame(configurations, frame)
	if err != nil {
		err = fmt.Errorf("Failed to parse connection.start frame: %v", err)
		return err
	}
	startResponse, ok := untypedResp.(*shared.ConnectionStartFrame)
	if !ok {
		err = fmt.Errorf("Type assertion failed")
		return err
	}

	startOkRequest, err := shared.CreateConnectionStartOkPayload(configurations, startResponse)
	if err != nil {
		err = fmt.Errorf("Failed to create connection.start-ok frame: %v", err)
		return err
	}
	startOkFrame := shared.CreateConnectionStartOkFrame(&startOkRequest)
	if err := shared.SendFrame(conn, startOkFrame); err != nil {
		err = fmt.Errorf("Failed to send connection.start-ok frame: %v", err)
		return err
	}

	/** connection.tune **/
	// Read connection.tune frame
	frame, err = shared.ReadFrame(conn)
	if err != nil {
		err = fmt.Errorf("Failed to read connection.tune frame: %v", err)
		return err
	}
	untypedResp, err = shared.ParseFrame(configurations, frame)
	tuneOkResponse, ok := untypedResp.([]byte)
	if !ok {
		err = fmt.Errorf("Type assertion ConnectionTuneOkFrame failed")
		return err
	}
	if err := shared.SendFrame(conn, tuneOkResponse); err != nil {
		log.Fatalf("Failed to send connection.tune-ok frame: %v", err)
	}
	/** end of connection.tune **/

	/** connection.open **/
	vhost := (*configurations)["vhost"].(string)
	openOkFrame := shared.CreateConnectionOpenFrame(vhost)
	if err := shared.SendFrame(conn, openOkFrame); err != nil {
		log.Fatalf("Failed to send connection.open-ok frame: %v", err)
	}

	frame, err = shared.ReadFrame(conn)
	if err != nil {
		log.Fatalf("Failed to read frame: %v", err)
	}
	fmt.Printf("Received connection.open-ok: %x\n", frame)
	_, err = shared.ParseFrame(configurations, frame)

	// Parse open-ok frame

	fmt.Println("AMQP handshake complete")
	return nil
}

func isAMQPVersionResponse(frame []byte) bool {
	return bytes.Equal(frame[:4], []byte("AMQP"))
}

func (c *Client) sendHeartbeats() {
	// heartbeatInterval := 5 // You can set this to the appropriate interval
	heartbeatInterval := c.config.HeartbeatInterval
	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			heartbeatFrame := shared.CreateHeartbeatFrame()
			err := shared.SendFrame(c.conn, heartbeatFrame)
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Println("Heartbeat sent")

			// Add a case for stopping the heartbeat loop if needed
			// case <-stopHeartbeatChan:
			//     return
		}
	}
}
