package shared

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

// type FieldTable
type ClientConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Vhost    string
}

type AMQP_Key struct {
	Key  string
	Type string
}

type AMQP_Type struct {
}

func SendProtocolHeader(conn net.Conn) error {
	header := []byte(constants.AMQP_PROTOCOL_HEADER)
	_, err := conn.Write(header)
	return err
}

func ReadProtocolHeader(conn net.Conn) ([]byte, error) {
	header := make([]byte, 8)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func ReadHeader(conn net.Conn) ([]byte, error) {
	header := make([]byte, 7)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// encodeTable encodes a proper AMQP field table
func EncodeTable(table map[string]interface{}) []byte {
	var buf bytes.Buffer

	for key, value := range table {
		// Field name
		buf.Write(EncodeShortStr(key))

		// Field value type and value
		switch v := value.(type) {
		case string:
			buf.WriteByte('S') // Field value type 'S' (string)
			buf.Write(EncodeLongStr([]byte(v)))

		case int:
			buf.WriteByte('I') // Field value type 'I' (int)
			binary.Write(&buf, binary.BigEndian, int32(v))
			// Add cases for other types as needed

		// In the case map[string]interface:
		case map[string]interface{}:
			// Recursively encode the nested map
			buf.WriteByte('F') // Field value type 'F' (field table)
			encodedTable := EncodeTable(v)
			buf.Write(EncodeLongStr(encodedTable))

		case bool:
			buf.WriteByte('t')
			if v {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}

		default:
			buf.WriteByte('U')
		}
	}
	return buf.Bytes()
}

// decodeTable decodes an AMQP field table from a byte slice
func DecodeTable(data []byte) (map[string]interface{}, error) {

	table := make(map[string]interface{})
	buf := bytes.NewReader(data)

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

		case 'F':
			var strLength uint32
			if err := binary.Read(buf, binary.BigEndian, &strLength); err != nil {
				return nil, err
			}

			strValue := make([]byte, strLength)
			_, err := buf.Read(strValue)
			if err != nil {
				return nil, err
			}

			value, err := DecodeTable(strValue)
			if err != nil {
				return nil, err
			}
			table[string(fieldName)] = value

		case 't':
			value, err := DecodeBoolean(buf)
			if err != nil {
				return nil, err
			}
			table[string(fieldName)] = value

		// Add cases for other types as needed

		default:
			return nil, fmt.Errorf("unknown field type: %c", fieldType)
		}
	}
	return table, nil
}

func ReadFrame(conn net.Conn) ([]byte, error) {
	// all frames starts with a 7-octet header
	frameHeader := make([]byte, 7)
	_, err := io.ReadFull(conn, frameHeader)
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
	_, err = io.ReadFull(conn, framePayload)
	if err != nil {
		return nil, err
	}

	// frame-end is a 1-octet after the payload
	frameEnd := make([]byte, 1)
	_, err = io.ReadFull(conn, frameEnd)
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

func FormatHeader(frameType uint8, channel uint16, payloadSize uint32) []byte {
	header := make([]byte, 7)
	header[0] = frameType
	binary.BigEndian.PutUint16(header[1:3], channel)
	binary.BigEndian.PutUint32(header[3:7], uint32(payloadSize))
	return header
}

func EncodeLongStr(data []byte) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

func EncodeShortStr(data string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(len(data)))
	buf.WriteString(data)
	return buf.Bytes()
}

func EncodeSecurityPlain(securityStr string) []byte {
	// Concatenate username, null byte, and password
	// securityStr := username + "\x00" + password
	// Replace spaces with null bytes
	encodedStr := strings.ReplaceAll(securityStr, " ", "\x00")
	// Encode length as a uint32 and append the encoded string
	length := uint32(len(encodedStr))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, length)
	buf.WriteString(encodedStr)
	return buf.Bytes()
}

func DecodeLongStr(buf *bytes.Reader) (string, error) {
	var strLen uint32
	err := binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return "", err
	}

	strData := make([]byte, strLen)
	_, err = buf.Read(strData)
	if err != nil {
		return "", err
	}

	return string(strData), nil
}

func DecodeShortStr(buf *bytes.Reader) (string, error) {
	var strLen uint8
	err := binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return "", err
	}

	strData := make([]byte, strLen)
	_, err = buf.Read(strData)
	if err != nil {
		return "", err
	}

	return string(strData), nil
}

func DecodeShortInt(buf *bytes.Reader) (uint16, error) {
	var value uint16
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func DecodeLongInt(buf *bytes.Reader) (uint32, error) {
	var value uint32
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func DecodeBoolean(buf *bytes.Reader) (bool, error) {
	var value uint8
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

func DecodeSecurityPlain(buf *bytes.Reader) (string, error) {
	var strLen uint32
	err := binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return "", err
	}

	if uint32(buf.Len()) < strLen {
		fmt.Printf("Rached EOF.  buf.Len(): %d\n", buf.Len())
		return "", io.EOF
	}

	// Read each byte and replace 0x00 with a space
	strData := make([]byte, strLen)
	for i := uint32(0); i < strLen; i++ {
		b, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if b == 0x00 {
			strData[i] = ' '
		} else {
			strData[i] = b
		}
	}

	return string(strData), nil
}

func ParseFrame(config *ClientConfig, frame []byte) (interface{}, error) {
	if len(frame) < 7 {
		return nil, fmt.Errorf("frame too short")
	}

	frameType := frame[0]
	channel := binary.BigEndian.Uint16(frame[1:3])
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if len(frame) < int(7+payloadSize) {
		return nil, fmt.Errorf("frame too short")
	}
	payload := frame[7:]

	switch frameType {
	case byte(constants.TYPE_METHOD):
		fmt.Printf("Received METHOD frame on channel %d\n", channel)
		return parseMethodFrame(config, channel, payload)
	case byte(constants.TYPE_HEARTBEAT):
		err := processHeartbeat(channel)
		return frame, err
	default:
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func processHeartbeat(channel uint16) error {
	// TODO: Implement heartbeat processing
	fmt.Printf("Received HEARTBEAT frame on channel %d\n", channel)
	return nil
}

func parseMethodFrame(config *ClientConfig, channel uint16, payload []byte) (interface{}, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short")
	}

	classID := binary.BigEndian.Uint16(payload[0:2])
	methodID := binary.BigEndian.Uint16(payload[2:4])
	methodPayload := payload[4:]

	switch classID {
	case uint16(constants.CONNECTION):
		fmt.Printf("Received CONNECTION frame on channel %d\n", channel)
		return parseConnectionMethod(config, methodID, methodPayload)
	default:
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}
