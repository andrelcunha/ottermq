package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

const (
	AMQP_PROTOCOL_HEADER = "AMQP\x00\x00\x09\x01"
)

type ProtocolHandler struct {
	conn net.Conn
}

func NewProtocolHandler(conn net.Conn) *ProtocolHandler {
	return &ProtocolHandler{conn: conn}
}

func (ph *ProtocolHandler) SendProtocolHeader() error {
	header := []byte(AMQP_PROTOCOL_HEADER)
	_, err := ph.conn.Write(header)
	return err
}

func (ph *ProtocolHandler) ReadProtocolHeader() ([]byte, error) {
	header := make([]byte, 8)
	_, err := io.ReadFull(ph.conn, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (ph *ProtocolHandler) ReadFrame() ([]byte, error) {
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

func (ph *ProtocolHandler) SendFrame(frame []byte) error {
	_, err := ph.conn.Write(frame)
	return err
}

func (ph *ProtocolHandler) ConnectionHandshake() error {
	// read the protocol header from the client
	clientHeader, err := ph.ReadProtocolHeader()
	if err != nil {
		return err
	}

	expectedHeader := []byte(AMQP_PROTOCOL_HEADER)
	if !bytes.Equal(clientHeader, expectedHeader) {
		if err := ph.SendProtocolHeader(); err != nil {
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
	panic("unimplemented")
}

func createConnectionOpenFrame() []byte {
	panic("unimplemented")
}

func createConnectionStartFrame() []byte {
	panic("not implemented")
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

	propertiesBuf := encodeTable(serverProperties)
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

// encodeTable encodes a proper AMQP field table
func encodeTable(table map[string]interface{}) []byte {
	var buf bytes.Buffer

	for key, value := range table {
		// Field name
		buf.WriteByte(byte(len(key)))
		buf.WriteString(key)

		// Field value type and value
		switch v := value.(type) {
		case string:
			buf.WriteByte('S') // Field value type 'S' (string)
			binary.Write(&buf, binary.BigEndian, uint32(len(v)))
			buf.WriteString(v)
		case int:
			buf.WriteByte('I') // Field value type 'I' (int)
			binary.Write(&buf, binary.BigEndian, int32(v))
			// Add cases for other types as needed
		}
	}
	return buf.Bytes()
}

// decodeTable decodes an AMQP field table from a byte slice
func decodeTable(data []byte) (map[string]interface{}, error) {

	table := make(map[string]interface{})
	buf := bytes.NewBuffer(data)

	for buf.Len() > 0 {
		// Read field name
		fieldNameLength, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		fieldName := make([]byte, fieldNameLength)
		_, err = buf.Read(fieldName)
		if err != nil {
			return nil, err
		}

		// Read field value type
		fieldType, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		// Read field value based on the type
		switch fieldType {
		case 'S': // String
			var strLength uint32
			if err := binary.Read(buf, binary.BigEndian, &strLength); err != nil {
				return nil, err
			}

			strValue := make([]byte, strLength)
			_, err := buf.Read(strValue)
			if err != nil {
				return nil, err
			}

			table[string(fieldName)] = string(strValue)

		case 'I': // Integer (simplified, normally long-int should be used)
			var intValue int32
			if err := binary.Read(buf, binary.BigEndian, &intValue); err != nil {
				return nil, err
			}

			table[string(fieldName)] = intValue

		// Add cases for other types as needed

		default:
			return nil, fmt.Errorf("unknown field type: %c", fieldType)
		}
	}
	return table, nil
}

func formatHeader(frameType uint8, channel uint16, payloadSize uint32) []byte {
	header := make([]byte, 7)
	header[0] = frameType
	binary.BigEndian.PutUint16(header[1:3], channel)
	binary.BigEndian.PutUint32(header[3:7], uint32(payloadSize))
	fmt.Printf("header: %x\n", header)
	return header
}
