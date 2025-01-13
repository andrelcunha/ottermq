package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp/message"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	. "github.com/andrelcunha/ottermq/pkg/connection/utils"
)

func parseBasicMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	// case uint16(constants.BASIC_ACK):
	// 	fmt.Printf("[DEBUG] Received BASIC_ACK frame \n")
	// 	return parseBasicAckFrame(payload)

	// case uint16(constants.BASIC_REJECT):
	// 	fmt.Printf("[DEBUG] Received BASIC_REJECT frame \n")
	// 	return parseBasicRejectFrame(payload)

	case uint16(constants.BASIC_PUBLISH):
		fmt.Printf("[DEBUG] Received BASIC_PUBLISH frame \n")
		return parseBasicPublishFrame(payload)

	// case uint16(constants.BASIC_RETURN):
	// 	fmt.Printf("[DEBUG] Received BASIC_RETURN frame \n")
	// 	return parseBasicReturnFrame(payload)

	// case uint16(constants.BASIC_DELIVER):
	// 	fmt.Printf("[DEBUG] Received BASIC_DELIVER frame \n")
	// 	return parseBasicDeliverFrame(payload)

	case uint16(constants.BASIC_GET):
		fmt.Printf("[DEBUG] Received BASIC_GET frame \n")
		return parseBasicGetFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseBasicGetFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
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
	flags := DecodeBasicGetFlags(octet)
	noAck := flags["noAck"]
	msg := &message.BasicGetMessage{
		Reserved1: reserved1,
		Queue:     queue,
		NoAck:     noAck,
	}
	return &amqp.RequestMethodMessage{
		Content: msg,
	}, nil
}

func DecodeBasicGetFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"noAck", "flag2", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}

func parseBasicPublishFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 10 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)
	reserved1, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserved1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
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
	flags := DecodeBasicPublishFlags(octet)
	mandatory := flags["mandatory"]
	immediate := flags["immediate"]
	msg := &message.BasicPublishMessage{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Mandatory:  mandatory,
		Immediate:  immediate,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] BasicPublish fomated: %+v \n", msg)
	return request, nil
}

func DecodeBasicPublishFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"mandatory", "immediate", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}
