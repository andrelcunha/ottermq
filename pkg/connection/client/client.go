package client

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/shared"
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
	ClientHandshake(conn, c.config)

}

func SendFrame(conn net.Conn, frame []byte) error {
	_, err := conn.Write(frame)
	return err
}

// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func ClientHandshake(conn net.Conn, config *shared.ClientConfig) error {
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
	untypedResp, err := shared.ParseFrame(config, frame)
	if err != nil {
		err = fmt.Errorf("Failed to parse connection.start frame: %v", err)
		return err
	}
	startOkResponse, ok := untypedResp.([]byte)
	if !ok {
		err = fmt.Errorf("Type assertion ConnectionStartFrame failed")
		return err
	}
	if err := SendFrame(conn, startOkResponse); err != nil {
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
	untypedResp, err = shared.ParseFrame(config, frame)
	tuneOkResponse, ok := untypedResp.([]byte)
	if !ok {
		err = fmt.Errorf("Type assertion ConnectionTuneOkFrame failed")
		return err
	}
	if err := SendFrame(conn, tuneOkResponse); err != nil {
		log.Fatalf("Failed to send connection.tune-ok frame: %v", err)
	}
	/** end of connection.tune **/

	/** connection.open **/
	openOkFrame := shared.CreateConnectionOpenFrame(config.Vhost)
	if err := SendFrame(conn, openOkFrame); err != nil {
		log.Fatalf("Failed to send connection.open-ok frame: %v", err)
	}

	frame, err = shared.ReadFrame(conn)
	if err != nil {
		log.Fatalf("Failed to read frame: %v", err)
	}
	fmt.Printf("Received connection.open-ok: %x\n", frame)
	_, err = shared.ParseFrame(config, frame)

	// Parse open-ok frame

	fmt.Println("AMQP handshake complete")
	return nil
}

func isAMQPVersionResponse(frame []byte) bool {
	return bytes.Equal(frame[:4], []byte("AMQP"))
}
