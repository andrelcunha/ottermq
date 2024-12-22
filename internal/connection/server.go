package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

type ServerConnectionHandler struct {
	conn net.Conn
}

func NewServerConnectionHandler(conn net.Conn) *ServerConnectionHandler {
	return &ServerConnectionHandler{conn: conn}
}

func (ph *ServerConnectionHandler) ReadFrame() ([]byte, error) {
	// all frames starts with a 7-octet header
	frameHeader := make([]byte, 7)
	_, err := io.ReadFull(ph.conn, frameHeader)
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
	_, err = io.ReadFull(ph.conn, framePayload)
	if err != nil {
		return nil, err
	}

	// frame-end is a 1-octet after the payload
	frameEnd := make([]byte, 1)
	_, err = io.ReadFull(ph.conn, frameEnd)
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

func (ph *ServerConnectionHandler) SendFrame(frame []byte) error {
	_, err := ph.conn.Write(frame)
	return err
}

func (ph *ServerConnectionHandler) ConnectionHandshake() error {
	// read the protocol header from the client
	clientHeader, err := shared.ReadProtocolHeader(ph.conn)
	if err != nil {
		return err
	}

	expectedHeader := []byte(constants.AMQP_PROTOCOL_HEADER)
	if !bytes.Equal(clientHeader, expectedHeader) {
		if err := shared.SendProtocolHeader(ph.conn); err != nil {
			return err
		}
		return fmt.Errorf("invalid protocol header")
	}

	// send connection.start frame
	startFrame := createConnectionStartFrame()
	if err := ph.SendFrame(startFrame); err != nil {
		return err
	}

	// read connecion.start-ok frame
	frame, err := ph.ReadFrame()
	if err != nil {
		return err
	}
	fmt.Printf("Received connection.start-ok: %x\n", frame)

	// send connection.tune frame
	tuneFrame := createConnectionTuneFrame()
	if err := ph.SendFrame(tuneFrame); err != nil {
		return err
	}

	// read connection.tune-ok frame
	frame, err = ph.ReadFrame()
	if err != nil {
		return err
	}
	fmt.Printf("Received connection.tune-ok: %x\n", frame)

	//send connection.open frame
	frame = createConnectionOpenFrame()
	if err := ph.SendFrame(frame); err != nil {
		return err
	}

	// read connection.open-ok frame
	frame, err = ph.ReadFrame()
	if err != nil {
		return err
	}
	fmt.Printf("Received connection.open-ok: %x\n", frame)

	return nil
}

func createConnectionTuneFrame() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{1, 0, 0, 0, 0, 0, 4, 0, 10, 0, 30})
	buf.WriteByte(0xCE)
	return buf.Bytes()
}

func createConnectionOpenFrame() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{1, 0, 0, 0, 0, 0, 4, 0, 10, 0, 40})
	buf.WriteByte(0xCE)
	return buf.Bytes()
}

func createConnectionStartFrame() []byte {
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
	headerBuf := formatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame
}

func formatHeader(frameType uint8, channel uint16, payloadSize uint32) []byte {
	header := make([]byte, 7)
	header[0] = frameType
	binary.BigEndian.PutUint16(header[1:3], channel)
	binary.BigEndian.PutUint32(header[3:7], uint32(payloadSize))
	fmt.Printf("header: %x\n", header)
	return header
}
