package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

func (cc *ClientConnetion) SendProtocolHeader() error {
	header := []byte(constants.AMQP_PROTOCOL_HEADER)
	_, err := cc.conn.Write(header)
	return err
}

func (cc *ClientConnetion) ReadFrame() ([]byte, error) {
	// all frames starts with a 7-octet header
	frameHeader := make([]byte, 7)
	_, err := io.ReadFull(cc.conn, frameHeader)
	if err != nil {
		return nil, err
	}

	// fist octet is the type of frame
	// frameType := binary.BigEndian.Uint16(frameHeader[0:1])

	// 2nd and 3rd octets (short) are the channel number
	// channelNum := binary.BigEndian.Uint16(frameHeader[1:3])

	// 4th to 7th octets (long) are the size of the payload
	payloadSize := binary.BigEndian.Uint32(frameHeader[3:])

	// read the framePayload
	framePayload := make([]byte, payloadSize)
	_, err = io.ReadFull(cc.conn, framePayload)
	if err != nil {
		return nil, err
	}

	// frame-end is a 1-octet after the payload
	frameEnd := make([]byte, 1)
	_, err = io.ReadFull(cc.conn, frameEnd)
	if err != nil {
		return nil, err
	}

	// check if the frame-end is correct (0xCE)
	if frameEnd[0] != 0xCE {
		// return nil, ErrInvalidFrameEnd
		return nil, fmt.Errorf("invalid frame end octet")
	}

	return append(frameHeader, framePayload...), nil
}

func (cc *ClientConnetion) SendFrame(frame []byte) error {
	_, err := cc.conn.Write(frame)
	return err
}

func (cc *ClientConnetion) ConnectionHandshake() error {
	// Send Protocol Header
	if err := cc.SendProtocolHeader(); err != nil {
		log.Fatalf("Failed to send protocol header: %v", err)
	}

	// Read connection.start frame
	frame, err := cc.ReadFrame()
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
	frame, err = cc.ReadFrame()
	if err != nil {
		log.Fatalf("Failed to read connection.tune frame: %v", err)
	}
	fmt.Printf("Received connection.tune: %x\n", frame)

	// Send connection.tune-ok frame
	tuneOkFrame := createConnectionTuneOkFrame()
	if err := cc.SendFrame(tuneOkFrame); err != nil {
		log.Fatalf("Failed to send connection.tune-ok frame: %v", err)
	}

	frame, err = cc.ReadFrame()
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
	headerBuf := FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame

}

func FormatHeader(frameType uint8, channel uint16, payloadSize uint32) []byte {
	header := make([]byte, 7)
	header[0] = frameType
	binary.BigEndian.PutUint16(header[1:3], channel)
	binary.BigEndian.PutUint32(header[3:7], uint32(payloadSize))
	fmt.Printf("header: %x\n", header)
	return header
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
