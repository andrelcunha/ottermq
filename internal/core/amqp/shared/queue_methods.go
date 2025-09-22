package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

func parseQueueMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(amqp.QUEUE_DECLARE):
		fmt.Printf("[DEBUG] Received QUEUE_DECLARE frame \n")
		return parseQueueDeclareFrame(payload)
	case uint16(amqp.QUEUE_DELETE):
		fmt.Printf("[DEBUG] Received QUEUE_DELETE frame \n")
		return parseQueueDeleteFrame(payload)
	case uint16(amqp.QUEUE_BIND):
		fmt.Printf("[DEBUG] Received QUEUE_BIND frame \n")
		return parseQueueBindFrame(payload)

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
	reserverd1, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserverd1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	queueName, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := utils.DecodeQueueDeclareFlags(octet)
	ifUnused := flags["ifUnused"]
	durable := flags["durable"]
	exclusive := flags["exclusive"]
	noWait := flags["noWait"]

	var arguments map[string]interface{}
	if buf.Len() > 4 {

		argumentsStr, err := utils.DecodeLongStr(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to decode arguments: %v", err)
		}
		arguments, err = utils.DecodeTable([]byte(argumentsStr))
		if err != nil {
			return nil, fmt.Errorf("failed to read arguments: %v", err)
		}
	}
	msg := &amqp.QueueDeclareMessage{
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
	reserverd1, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserverd1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	exchangeName, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := utils.DecodeQueueDeclareFlags(octet)
	ifUnused := flags["ifUnused"]
	noWait := flags["noWait"]

	msg := &amqp.QueueDeleteMessage{
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

func parseQueueBindFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] Received QUEUE_BIND frame %x \n", payload)
	buf := bytes.NewReader(payload)
	reserverd1, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserverd1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	queue, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode queue name: %v", err)
	}
	exchange, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	fmt.Printf("[DEBUG] Exchange name: %s\n", exchange)
	routingKey, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode routing key: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := utils.DecodeQueueBindFlags(octet)
	noWait := flags["noWait"]
	arguments := make(map[string]interface{})
	if buf.Len() > 4 {
		argumentsStr, err := utils.DecodeLongStr(buf)
		arguments, err = utils.DecodeTable([]byte(argumentsStr))
		if err != nil {
			return nil, fmt.Errorf("failed to read arguments: %v", err)
		}
	}
	msg := &amqp.QueueBindMessage{
		Queue:      queue,
		Exchange:   exchange,
		RoutingKey: routingKey,
		NoWait:     noWait,
		Arguments:  arguments,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Queue fomated: %+v \n", msg)
	return request, nil
}
