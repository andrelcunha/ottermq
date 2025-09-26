package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

type Framer interface {
	ReadFrame(conn net.Conn) ([]byte, error)
	SendFrame(conn net.Conn, frame []byte) error
	Handshake(configurations *map[string]any, conn net.Conn) (*ConnectionInfo, error)
	ParseFrame(frame []byte) (any, error)
	SendHearbeat(conn net.Conn) error
	CloseChannelFrame(channel uint16) []byte
	CloseConnectionFrame(channel uint16) []byte
	CreateExchangeDeclareFrame(channel uint16, request *RequestMethodMessage) []byte
}

type DefaultFramer struct{}

func (d *DefaultFramer) ReadFrame(conn net.Conn) ([]byte, error) {
	return readFrame(conn)
}

func (d *DefaultFramer) SendFrame(conn net.Conn, frame []byte) error {
	return sendFrame(conn, frame)
}

func (d *DefaultFramer) Handshake(configurations *map[string]any, conn net.Conn) (*ConnectionInfo, error) {
	return handshake(configurations, conn)
}

func (d *DefaultFramer) ParseFrame(frame []byte) (any, error) {
	return parseFrame(frame)
}

func (d *DefaultFramer) SendHearbeat(conn net.Conn) error {
	heartbeatFrame := createHeartbeatFrame()
	return sendFrame(conn, heartbeatFrame)
}

func (d *DefaultFramer) CloseChannelFrame(channel uint16) []byte {
	return closeChannelFrame(channel)
}

func (d *DefaultFramer) CloseConnectionFrame(channel uint16) []byte {
	return createConnectionCloseFrame(channel)
}

func (d *DefaultFramer) CreateExchangeDeclareFrame(channel uint16, request *RequestMethodMessage) []byte {
	return createExchangeDeclareFrame(channel, request)
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

func createContentPropertiesTable(flags []string, buf *bytes.Reader) (*BasicProperties, error) {
	props := &BasicProperties{}
	for _, flag := range flags {
		switch flag {
		case "contentType": // shortstr
			contentType, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode content type: %v", err)
			}
			props.ContentType = contentType

		case "contentEncoding": // shortstr
			contentEncoding, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode content encoding: %v", err)
			}
			props.ContentEncoding = contentEncoding

		case "headers": // longstr (table)
			headersStr, err := utils.DecodeLongStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode headers: %v", err)
			}
			headers, err := utils.DecodeTable([]byte(headersStr))
			if err != nil {
				return nil, fmt.Errorf("failed to decode headers: %v", err)
			}
			props.Headers = headers

		case "deliveryMode": // octet
			deliveryMode, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to decode delivery mode: %v", err)
			}
			if deliveryMode != 1 && deliveryMode != 2 {
				return nil, fmt.Errorf("delivery mode must be 1 or 2")
			}
			// var deliveryModeStr string // 1: non-persistent, 2: persistent
			// if deliveryMode == 1 {
			// 	deliveryModeStr = "non-persistent"
			// } else {
			// 	deliveryModeStr = "persistent"
			// }
			props.DeliveryMode = deliveryMode

		case "priority": // octet (0-9)
			priority, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to decode priority: %v", err)
			}
			if int(priority) < 0 || int(priority) > 9 {
				return nil, fmt.Errorf("priority must be between 0 and 9")
			}
			props.Priority = priority

		case "correlationID": // shortstr
			correlationID, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode correlation ID: %v", err)
			}
			props.CorrelationID = correlationID

		case "replyTo": // shortstr
			replyTo, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode reply to: %v", err)
			}
			props.ReplyTo = replyTo

		case "expiration": // shortstr
			expiration, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode expiration: %v", err)
			}
			props.Expiration = expiration

		case "messageID": // shortstr
			messageID, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode message ID: %v", err)
			}
			props.MessageID = messageID

		case "timestamp": // 64 bit timestamp
			timestamp, err := utils.DecodeTimestamp(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode timestamp: %v", err)
			}
			props.Timestamp = timestamp

		case "type": // shortstr
			type_, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode type: %v", err)
			}
			props.Type = type_

		case "userID": // shortstr
			userID, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode user ID: %v", err)
			}
			props.UserID = userID

		case "appID": // shortstr
			appID, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode app ID: %v", err)
			}
			props.AppID = appID

		case "reserved": // shortstr
			reserved, err := utils.DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode reserved: %v", err)
			}
			props.Reserved = reserved

		default:
			return nil, fmt.Errorf("unknown flag: %s", flag)
		}
	}
	return props, nil
}

func formatMethodFrame(channelNum uint16, class TypeClass, method TypeMethod, methodPayload []byte) []byte {
	var payloadBuf bytes.Buffer

	binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	// payloadBuf.WriteByte(binary.BigEndian.AppendUint16()[])

	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(TYPE_METHOD) // METHOD frame type
	headerBuf := formatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame
}

func createConnectionTuneFrame(tune *ConnectionTune) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_TUNE

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func createConnectionTuneOkFrame(tune *ConnectionTune) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_TUNE_OK

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func createConnectionOpenOkFrame() []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_OPEN_OK

	// Reserved-1 (bit) - set to 0
	payloadBuf.WriteByte(0)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func fineTune(tune *ConnectionTune) *ConnectionTune {
	// TODO: get values from config
	tune.ChannelMax = getSmalestShortInt(2047, tune.ChannelMax)
	tune.FrameMax = getSmalestLongInt(131072, tune.FrameMax)
	tune.Heartbeat = getSmalestShortInt(10, tune.Heartbeat)

	return tune
}

func getSmalestShortInt(a, b uint16) uint16 {
	if a == 0 {
		a = math.MaxUint16
	}
	if b == 0 {
		b = math.MaxInt16
	}
	if a < b {
		return a
	}
	return b
}

func getSmalestLongInt(a, b uint32) uint32 {
	if a == 0 {
		a = math.MaxUint32
	}
	if b == 0 {
		b = math.MaxInt32
	}
	if a < b {
		return a
	}
	return b
}
