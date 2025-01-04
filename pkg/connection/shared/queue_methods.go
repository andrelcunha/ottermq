package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp/message"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

func parseQueueMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(constants.QUEUE_DECLARE):
		fmt.Printf("[DEBUG] Received QUEUE_DECLARE frame \n")
		return parseQueueDeclareFrame(payload)
	case uint16(constants.QUEUE_DELETE):
		fmt.Printf("[DEBUG] Received QUEUE_DELETE frame \n")
		return parseQueueDeleteFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

// Fields:
// 0-1: reserved short int
// 2: exchange name - length (short)
// 3: type - (string)
// 4: passive - (bit)
// 5: durable - (bit)
// 6: reserved 2
// 7: reserved 3
// 8: no-wait - (bit)
// 9: arguments - (table)
func parseQueueDeclareFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)
	reserverd1, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserverd1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	queueName, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := DecodeQueueDeclareFlags(octet)
	ifUnused := flags["ifUnused"]
	durable := flags["durable"]
	exclusive := flags["exclusive"]
	noWait := flags["noWait"]

	var arguments map[string]interface{}
	if buf.Len() > 4 {

		argumentsStr, err := DecodeLongStr(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to decode arguments: %v", err)
		}
		arguments, err = DecodeTable([]byte(argumentsStr))
		if err != nil {
			return nil, fmt.Errorf("failed to read arguments: %v", err)
		}
	}
	msg := &message.QueueDeclareMessage{
		QueueName: queueName,
		Durable:   durable,
		IfUnused:  ifUnused,
		Exclusive: exclusive,
		NoWait:    noWait,
		Arguments: arguments,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Queue fomated: %+v \n", msg)
	return request, nil
}

// Fields:
// 0-1: reserved short int
// 2: exchange name - length (short)
// 3: if-unused - (bit)
// 4: no-wait - (bit)
func parseQueueDeleteFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] Received QUEUE_DELETE frame %x \n", payload)

	buf := bytes.NewReader(payload)
	reserverd1, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserverd1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	exchangeName, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := DecodeQueueDeclareFlags(octet)
	ifUnused := flags["ifUnused"]
	noWait := flags["noWait"]

	msg := &message.QueueDeleteMessage{
		QueueName: exchangeName,
		IfUnused:  ifUnused,
		NoWait:    noWait,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Queue fomated: %+v \n", msg)
	return request, nil
}
