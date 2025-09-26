package amqp

import (
	"bytes"
	"encoding/binary"
	"time"
)

type BasicPublishMessage struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
}

type BasicGetMessage struct {
	Reserved1 uint16
	Queue     string
	NoAck     bool
}

type BasicGetOk struct {
	DeliveryTag  uint64
	Redelivered  bool
	Exchange     string
	RoutingKey   string
	MessageCount uint32
}

type BasicProperties struct {
	ContentType     string         // shortstr
	ContentEncoding string         // shortstr
	Headers         map[string]any // table
	DeliveryMode    uint8          // octet
	Priority        uint8          // octet
	CorrelationID   string         // shortstr
	ReplyTo         string         // shortstr
	Expiration      string         // shortstr
	MessageID       string         // shortstr
	Timestamp       time.Time      // timestamp (64 bits)
	Type            string         // shortsrt
	UserID          string         // shortstr
	AppID           string         // shortstr
	Reserved        string         // shortstr
}

func (props *BasicProperties) encodeBasicProperties() ([]byte, uint16, error) {
	var buf bytes.Buffer
	var flags uint16

	if props.ContentType != "" {
		flags |= (1 << 15)
		if err := encodeShortStr(&buf, props.ContentType); err != nil {
			return nil, 0, err
		}
	}
	if props.ContentEncoding != "" {
		flags |= (1 << 14)
		if err := encodeShortStr(&buf, props.ContentEncoding); err != nil {
			return nil, 0, err
		}
	}
	if props.Headers != nil {
		flags |= (1 << 13)
		encodedTable, err := encodeTable(props.Headers)
		if err != nil {
			return nil, 0, err
		}
		if _, err := buf.Write(encodedTable); err != nil {
			return nil, 0, err
		}
	}
	if props.DeliveryMode != 0 {
		flags |= (1 << 12)
		if err := encodeOctet(&buf, props.DeliveryMode); err != nil {
			return nil, 0, err
		}
	}
	if props.Priority != 0 {
		flags |= (1 << 11)
		if err := encodeOctet(&buf, props.Priority); err != nil {
			return nil, 0, err
		}
	}
	if props.CorrelationID != "" {
		flags |= (1 << 10)
		if err := encodeShortStr(&buf, props.CorrelationID); err != nil {
			return nil, 0, err
		}
	}
	if props.ReplyTo != "" {
		flags |= (1 << 9)
		if err := encodeShortStr(&buf, props.ReplyTo); err != nil {
			return nil, 0, err
		}
	}
	if props.Expiration != "" {
		flags |= (1 << 8)
		if err := encodeShortStr(&buf, props.Expiration); err != nil {
			return nil, 0, err
		}
	}
	if props.MessageID != "" {
		flags |= (1 << 7)
		if err := encodeShortStr(&buf, props.MessageID); err != nil {
			return nil, 0, err
		}
	}
	if !props.Timestamp.IsZero() {
		flags |= (1 << 6)
		if err := encodeTimestamp(&buf, props.Timestamp); err != nil {
			return nil, 0, err
		}
	}
	if props.Type != "" {
		flags |= (1 << 5)
		if err := encodeShortStr(&buf, props.Type); err != nil {
			return nil, 0, err
		}
	}
	if props.UserID != "" {
		flags |= (1 << 4)
		if err := encodeShortStr(&buf, props.UserID); err != nil {
			return nil, 0, err
		}
	}
	if props.AppID != "" {
		flags |= (1 << 3)
		if err := encodeShortStr(&buf, props.AppID); err != nil {
			return nil, 0, err
		}
	}
	if props.Reserved != "" {
		flags |= (1 << 2)
		if err := encodeShortStr(&buf, props.Reserved); err != nil {
			return nil, 0, err
		}
	}
	return buf.Bytes(), flags, nil
}

func decodeBasicGetFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"noAck", "flag2", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}

func decodeBasicPublishFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"mandatory", "immediate", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}

func encodeShortStr(buf *bytes.Buffer, value string) error {
	if err := buf.WriteByte(byte(len(value))); err != nil {
		return err
	}
	if _, err := buf.WriteString(value); err != nil {
		return err
	}
	return nil
}

func encodeTable(table map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	for key, value := range table { // Field name
		if err := encodeShortStr(&buf, key); err != nil {
			return nil, err
		}

		// Field value type and value
		switch v := value.(type) {
		case string:
			buf.WriteByte('S') // Field value type 'S' (string)
			if err := encodeLongStr(&buf, v); err != nil {
				return nil, err
			}

		case int:
			buf.WriteByte('I') // Field value type 'I' (int)
			if err := binary.Write(&buf, binary.BigEndian, int32(v)); err != nil {
				return nil, err
			}

		case map[string]interface{}: // Recursively encode the nested map
			buf.WriteByte('F') // Field value type 'F' (field table)
			encodedTable, err := encodeTable(v)
			if err != nil {
				return nil, err
			}
			if err := encodeLongStr(&buf, string(encodedTable)); err != nil {
				return nil, err
			}

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
	return buf.Bytes(), nil
}

func encodeOctet(buf *bytes.Buffer, value uint8) error {
	return buf.WriteByte(value)
}

func encodeTimestamp(buf *bytes.Buffer, value time.Time) error {
	timestamp := value.Unix()
	return binary.Write(buf, binary.BigEndian, timestamp)
}

func encodeLongStr(buf *bytes.Buffer, value string) error {
	if err := binary.Write(buf, binary.BigEndian, int32(len(value))); err != nil {
		return err
	}
	if _, err := buf.WriteString(value); err != nil {
		return err
	}
	return nil
}
