package client

import (
	"bytes"
	"fmt"
	"log"
	"net"

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

func (c *Client) Start() {
	log.Println("OtterMq is starting...")
	addr := fmt.Sprintf("%s:%s", c.config.Host, c.config.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
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
	ClientHandshake(configurations, conn)

}

// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func ClientHandshake(configurations map[string]interface{}, conn net.Conn) error {
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
	startOkResponse, ok := untypedResp.([]byte)
	if !ok {
		err = fmt.Errorf("Type assertion failed")
		return err
	}
	if err := shared.SendFrame(conn, startOkResponse); err != nil {
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
	vhost := configurations["vhost"].(string)
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
