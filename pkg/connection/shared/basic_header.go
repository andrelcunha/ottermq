package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
)

// ClassID: short
// Weight: short
// Body Size: long long
// Properties flags: short
// Properties: long (table)

func parseBasicHeader(headerPayload []byte) (*amqp.HeaderFrame, error) {

	buf := bytes.NewReader(headerPayload)
	classID, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode class ID: %v", err)
	}
	fmt.Printf("[DEBUG] Class ID: %d\n", classID)

	weight, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode weight: %v", err)
	}
	if weight != 0 {
		return nil, fmt.Errorf("weight must be 0")
	}
	fmt.Printf("[DEBUG] Weight: %d\n", weight)

	bodySize, err := DecodeLongLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode body size: %v", err)
	}
	fmt.Printf("[DEBUG] Body Size: %d\n", bodySize)

	shortFlags, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flags: %v", err)
	}
	flags := DecodeBasicHeaderFlags(shortFlags)
	properties, err := createContentPropertiesTable(flags, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode properties: %v", err)
	}
	fmt.Printf("[DEBUG] properties: %v\n", properties)
	header := &amqp.HeaderFrame{
		ClassID:    classID,
		BodySize:   bodySize,
		Properties: properties,
	}
	return header, nil

}

func DecodeBasicHeaderFlags(short uint16) []string {
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

func createContentPropertiesTable(flags []string, buf *bytes.Reader) (map[string]interface{}, error) {
	properties := make(map[string]interface{})
	for _, flag := range flags {
		switch flag {
		case "contentType": // shortstr
			contentType, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode content type: %v", err)
			}
			properties["contentType"] = contentType

		case "contentEncoding": // shortstr
			contentEncoding, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode content encoding: %v", err)
			}
			properties["contentEncoding"] = contentEncoding

		case "headers": // longstr (table)
			headersStr, err := DecodeLongStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode headers: %v", err)
			}
			headers, err := DecodeTable([]byte(headersStr))
			if err != nil {
				return nil, fmt.Errorf("failed to decode headers: %v", err)
			}
			properties["headers"] = headers

		case "deliveryMode": // octet
			deliveryMode, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to decode delivery mode: %v", err)
			}
			if deliveryMode != 1 && deliveryMode != 2 {
				return nil, fmt.Errorf("delivery mode must be 1 or 2")
			}
			var deliveryModeStr string // 1: non-persistent, 2: persistent
			if deliveryMode == 1 {
				deliveryModeStr = "non-persistent"
			} else {
				deliveryModeStr = "persistent"
			}
			properties["deliveryMode"] = deliveryModeStr

		case "priority": // octet (0-9)
			priority, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("failed to decode priority: %v", err)
			}
			properties["priority"] = priority
			if priority < 0 || priority > 9 {
				return nil, fmt.Errorf("priority must be between 0 and 9")
			}
			properties["priority"] = priority

		case "correlationID": // shortstr
			correlationID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode correlation ID: %v", err)
			}
			properties["correlationID"] = correlationID

		case "replyTo": // shortstr
			replyTo, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode reply to: %v", err)
			}
			properties["replyTo"] = replyTo

		case "expiration": // shortstr
			expiration, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode expiration: %v", err)
			}
			properties["expiration"] = expiration

		case "messageID": // shortstr
			messageID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode message ID: %v", err)
			}
			properties["messageID"] = messageID

		case "timestamp": // 64 bit timestamp
			timestamp, err := DecodeTimestamp(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode timestamp: %v", err)
			}
			properties["timestamp"] = timestamp

		case "type": // shortstr
			type_, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode type: %v", err)
			}
			properties["type"] = type_

		case "userID": // shortstr
			userID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode user ID: %v", err)
			}
			properties["userID"] = userID

		case "appID": // shortstr
			appID, err := DecodeShortStr(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode app ID: %v", err)
			}
			properties["appID"] = appID

		case "reserved": // shortstr
			reserved, err := DecodeShortInt(buf)
			if err != nil {
				return nil, fmt.Errorf("failed to decode reserved: %v", err)
			}
			properties["reserved"] = reserved

		default:
			return nil, fmt.Errorf("unknown flag: %s", flag)
		}
	}
	return properties, nil
}
