package amqp

import (
	"bytes"
	"encoding/binary"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

type RequestMethodMessage struct {
	Channel  uint16
	ClassID  uint16
	MethodID uint16
	Content  interface{}
}

type ResponseMethodMessage struct {
	Channel  uint16
	ClassID  uint16
	MethodID uint16
	Content  ContentList
}

type ContentList struct {
	KeyValuePairs []KeyValue
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func (msg ResponseMethodMessage) FormatMethodFrame() []byte {
	var payloadBuf bytes.Buffer
	class := msg.ClassID
	method := msg.MethodID
	channelNum := msg.Channel

	binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	methodPayload := formatMethodPayload(msg.Content)
	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(constants.TYPE_METHOD) // METHOD frame type
	headerBuf := FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame
}

func formatMethodPayload(content ContentList) []byte {
	var payloadBuf bytes.Buffer
	for _, kv := range content.KeyValuePairs {
		if kv.Key == INT_OCTET {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint8))
		} else if kv.Key == INT_SHORT {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint16))
		} else if kv.Key == INT_LONG {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint32))
		} else if kv.Key == INT_LONG_LONG {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint64))
		} else if kv.Key == BIT {
			if kv.Value.(bool) {
				payloadBuf.WriteByte(1)
			} else {
				payloadBuf.WriteByte(0)
			}
		} else if kv.Key == STRING_SHORT {
			payloadBuf.Write(EncodeShortStr(kv.Value.(string)))
		} else if kv.Key == STRING_LONG {
			payloadBuf.Write(EncodeLongStr(kv.Value.([]byte)))
		} else if kv.Key == TIMESTAMP {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(int64))
		} else if kv.Key == TABLE {
			encodedTable := EncodeTable(kv.Value.(map[string]interface{}))
			payloadBuf.Write(EncodeLongStr(encodedTable))
		}
	}
	return payloadBuf.Bytes()
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
