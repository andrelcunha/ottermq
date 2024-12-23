package client

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

type ClientConnetion struct {
	conn net.Conn
}

func NewProtocolHandler(conn net.Conn) *ClientConnetion {
	return &ClientConnetion{conn: conn}
}

func (cc *ClientConnetion) SendFrame(frame []byte) error {
	_, err := cc.conn.Write(frame)
	return err
}

func (cc *ClientConnetion) ConnectionHandshake() error {
	// Send Protocol Header
	if err := shared.SendProtocolHeader(cc.conn); err != nil {
		log.Fatalf("Failed to send protocol header: %v", err)
	}

	// Read connection.start frame
	frame, err := shared.ReadFrame(cc.conn)
	cc.parseConnectionStartOK(frame)

	if err != nil {
		log.Fatalf("Failed to read connection.start frame: %v", err)
	}
	fmt.Printf("Received connection.start: %x\n", frame)

	// Send connection.start-ok frame
	startOkFrame := createConnectionStartOkFrame()
	if err := cc.SendFrame(startOkFrame); err != nil {
		log.Fatalf("Failed to send connection.start-ok frame: %v", err)
	}

	// Read connection.tune frame
	frame, err = shared.ReadFrame(cc.conn)
	if err != nil {
		log.Fatalf("Failed to read connection.tune frame: %v", err)
	}
	fmt.Printf("Received connection.tune: %x\n", frame)

	// Send connection.tune-ok frame
	tuneOkFrame := createConnectionTuneOkFrame()
	if err := cc.SendFrame(tuneOkFrame); err != nil {
		log.Fatalf("Failed to send connection.tune-ok frame: %v", err)
	}

	frame, err = shared.ReadFrame(cc.conn)
	if err != nil {
		log.Fatalf("Failed to read connection.open frame: %v", err)
	}
	fmt.Printf("Received connection.open: %x\n", frame)

	openOkFrame := createOpenOkFrame()
	if err := cc.SendFrame(openOkFrame); err != nil {
		log.Fatalf("Failed to send connection.open-ok frame: %v", err)
	}

	fmt.Println("AMQP handshake complete")
	return nil
}

func createConnectionStartOkFrame() []byte {
	var payloadBuf bytes.Buffer

	serverProperties := map[string]interface{}{
		"product":     "OtterMQ",
		"version":     "0.1.0",
		"platform":    "Linux",
		"information": "https://github.com/andrelcunha/ottermq",
	}

	payloadBuf.WriteByte(0) // version-major
	payloadBuf.WriteByte(9) // version-minor

	propertiesBuf := shared.EncodeTable(serverProperties)
	payloadBuf.Write(propertiesBuf)

	// Payload: Mechanisms
	payloadBuf.Write([]byte{'P', 'L', 'A', 'I', 'N', 0})

	// Payload: Locales
	payloadBuf.Write([]byte("en_US"))

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(constants.TYPE_METHOD) // METHOD frame type
	channelNum := uint16(0)                   // Channel number
	headerBuf := shared.FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame

}

func createConnectionTuneOkFrame() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{1, 0, 0, 0, 0, 0, 4, 0, 10, 0, 31})
	buf.WriteByte(0xCE)
	return buf.Bytes()
}

func createOpenOkFrame() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{1, 0, 0, 0, 0, 0, 4, 0, 10, 0, 41})
	buf.WriteByte(0xCE)
	return buf.Bytes()
}

func (cc *ClientConnetion) parseConnectionStartOK(frame []byte) error {

	if isAMQPVersionResponse(frame) {
		// print the frame and close the connection
		buf := bytes.NewBuffer(frame)
		log.Println("Received:", buf.String())
		// cc.conn.Close()
		return fmt.Errorf("Version not supporte by the server")
	}

	return nil
}

func isAMQPVersionResponse(frame []byte) bool {
	return bytes.Equal(frame[:4], []byte("AMQP"))
}
