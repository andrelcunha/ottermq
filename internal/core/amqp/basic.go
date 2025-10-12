package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type BasicConsumeContent struct {
	Queue       string
	ConsumerTag string
	NoLocal     bool
	NoAck       bool
	Exclusive   bool
	NoWait      bool
	Arguments   map[string]any
}

type BasicPublishContent struct {
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
	ContentType     ContentType    // shortstr
	ContentEncoding string         // shortstr
	Headers         map[string]any // table
	DeliveryMode    DeliveryMode   // from octet: (1=non-persistent, 2=persistent)
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

func createBasicGetEmptyFrame(request *RequestMethodMessage) []byte {
	// Send Basic.GetEmpty
	reserved1 := KeyValue{
		Key:   STRING_SHORT,
		Value: "",
	}
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(BASIC_GET_EMPTY),
		Content:  ContentList{KeyValuePairs: []KeyValue{reserved1}},
	}.FormatMethodFrame()
	return frame
}

func createBasicGetOkFrame(request *RequestMethodMessage, exchange, routingkey string, msgCount uint32) []byte {
	msgGetOk := &BasicGetOk{
		DeliveryTag:  1,
		Redelivered:  false,
		Exchange:     exchange,
		RoutingKey:   routingkey,
		MessageCount: msgCount,
	}

	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(BASIC_GET_OK),
		Content:  *EncodeGetOkToContentList(msgGetOk),
	}.FormatMethodFrame()
	return frame
}

func (props *BasicProperties) encodeBasicProperties() ([]byte, uint16, error) {
	var buf bytes.Buffer
	var flags uint16

	if props.ContentType != "" {
		flags |= (1 << 15)
		if err := encodeShortStr(&buf, string(props.ContentType)); err != nil {
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
		if err := encodeOctet(&buf, uint8(props.DeliveryMode)); err != nil {
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

func decodeBasicHeaderFlags(short uint16) []string {
	flagNames := []string{
		"contentType",
		"contentEncoding",
		"headers",
		"deliveryMode",
		"priority",
		"correlationID",
		"replyTo",
		"expiration",
		"messageID",
		"timestamp",
		"type",
		"userID",
		"appID",
		"reserved",
	}
	var flags []string
	for i := 0; i < len(flagNames); i++ {
		if (short & (1 << uint(15-i))) != 0 {
			flags = append(flags, flagNames[i])
		}
	}
	return flags
}

// ClassID: short
// Weight: short
// Body Size: long long
// Properties flags: short
// Properties: long (table)

func parseBasicHeader(headerPayload []byte) (*HeaderFrame, error) {

	buf := bytes.NewReader(headerPayload)
	classID, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode class ID: %v", err)
	}
	log.Printf("[DEBUG] Class ID: %d\n", classID)

	weight, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode weight: %v", err)
	}
	if weight != 0 {
		return nil, fmt.Errorf("weight must be 0")
	}
	log.Printf("[DEBUG] Weight: %d\n", weight)

	bodySize, err := DecodeLongLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode body size: %v", err)
	}
	log.Printf("[DEBUG] Body Size: %d\n", bodySize)

	shortFlags, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flags: %v", err)
	}
	flags := decodeBasicHeaderFlags(shortFlags)
	properties, err := createContentPropertiesTable(flags, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode properties: %v", err)
	}
	log.Printf("[DEBUG] properties: %v\n", properties)
	header := &HeaderFrame{
		ClassID:    classID,
		BodySize:   bodySize,
		Properties: properties,
	}
	return header, nil
}

// REGION Basic_Methods

func parseBasicMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	// case uint16(BASIC_ACK):
	// 	log.Printf("[DEBUG] Received BASIC_ACK frame \n")
	// 	return parseBasicAckFrame(payload)

	// case uint16(BASIC_REJECT):
	// 	log.Printf("[DEBUG] Received BASIC_REJECT frame \n")
	// 	return parseBasicRejectFrame(payload)

	case uint16(BASIC_PUBLISH):
		log.Printf("[DEBUG] Received BASIC_PUBLISH frame \n")
		return parseBasicPublishFrame(payload)

	// case uint16(BASIC_RETURN):
	// 	log.Printf("[DEBUG] Received BASIC_RETURN frame \n")
	// 	return parseBasicReturnFrame(payload)

	// case uint16(BASIC_DELIVER):
	// 	log.Printf("[DEBUG] Received BASIC_DELIVER frame \n")
	// 	return parseBasicDeliverFrame(payload)

	case uint16(BASIC_GET):
		log.Printf("[DEBUG] Received BASIC_GET frame \n")
		return parseBasicGetFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseBasicPublishFrame(payload []byte) (*RequestMethodMessage, error) {
	// the payload must be at least 5 bytes long
	// 2 (reserved1) => 2 bytes
	// 1+ (exchange name) => short str = 1  (length) + 1 byte
	// 1+ (routing key) => short str = 1 (length) + 1 byte
	// 1 (flags) => octet = 1 byte
	if len(payload) < 5 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)

	// Note: AMQP spec says "shortstr" but all real implementations
	// (RabbitMQ, major clients) use short int (2 bytes) for reserved fields.
	// We follow industry practice for compatibility.
	_, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	exchange, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange: %v", err)
	}
	routingKey, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode routing key: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := DecodeFlags(octet, []string{"mandatory", "immediate"}, true)
	mandatory := flags["mandatory"]
	immediate := flags["immediate"]
	msg := &BasicPublishContent{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Mandatory:  mandatory,
		Immediate:  immediate,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Printf("[DEBUG] BasicPublish fomated: %+v \n", msg)
	return request, nil
}

func parseBasicGetFrame(payload []byte) (*RequestMethodMessage, error) {
	buf := bytes.NewReader(payload)
	reserved1, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserved1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	queue, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := DecodeFlags(octet, []string{"noAck"}, true)
	noAck := flags["noAck"]
	msg := &BasicGetMessage{
		Reserved1: reserved1,
		Queue:     queue,
		NoAck:     noAck,
	}
	return &RequestMethodMessage{
		Content: msg,
	}, nil
}
