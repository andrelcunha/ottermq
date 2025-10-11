package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func createContentPropertiesTable(flags []string, buf *bytes.Reader) (*BasicProperties, error) {
	props := &BasicProperties{}
	for _, flag := range flags {
		switch flag {
		case "contentType": // shortstr
			contentType, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode content type: %v", err)
			}
			props.ContentType = contentType

		case "contentEncoding": // shortstr
			contentEncoding, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode content encoding: %v", err)
			}
			props.ContentEncoding = contentEncoding

		case "headers": // longstr (table)
			headersStr, err := DecodeLongStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode headers: %v", err)
			}
			headers, err := DecodeTable([]byte(headersStr))
			if err != nil {
				return nil, fmt.Errorf("failed to decode headers: %v", err)
			}
			props.Headers = headers

		case "deliveryMode": // octet
			deliveryMode, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to decode delivery mode: %v", err)
			}
			if deliveryMode != 1 && deliveryMode != 2 { // 1: non-persistent, 2: persistent
				return nil, fmt.Errorf("delivery mode must be 1 or 2")
			}
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
			correlationID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode correlation ID: %v", err)
			}
			props.CorrelationID = correlationID

		case "replyTo": // shortstr
			replyTo, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode reply to: %v", err)
			}
			props.ReplyTo = replyTo

		case "expiration": // shortstr
			expiration, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode expiration: %v", err)
			}
			props.Expiration = expiration

		case "messageID": // shortstr
			messageID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode message ID: %v", err)
			}
			props.MessageID = messageID

		case "timestamp": // 64 bit timestamp
			timestamp, err := DecodeTimestamp(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode timestamp: %v", err)
			}
			props.Timestamp = timestamp

		case "type": // shortstr
			type_, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode type: %v", err)
			}
			props.Type = type_

		case "userID": // shortstr
			userID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode user ID: %v", err)
			}
			props.UserID = userID

		case "appID": // shortstr
			appID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode app ID: %v", err)
			}
			props.AppID = appID

		case "reserved": // shortstr
			reserved, err := DecodeShortStr(buf)
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

	_ = binary.Write(&payloadBuf, binary.BigEndian, uint16(class))  // Error ignored as bytes.Buffer.Write never fails
	_ = binary.Write(&payloadBuf, binary.BigEndian, uint16(method)) // Error ignored as bytes.Buffer.Write never fails

	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(TYPE_METHOD) // METHOD frame type
	headerBuf := formatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, FRAME_END) // frame-end

	return frame
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
