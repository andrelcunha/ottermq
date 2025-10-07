package amqp

import (
	"bytes"
	"fmt"

	"github.com/rs/zerolog/log"
)

func createQueueDeclareFrame(request *RequestMethodMessage, queueName string, messageCount, consumerCount uint32) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(QUEUE_DECLARE_OK),
		Content: ContentList{
			KeyValuePairs: []KeyValue{
				{
					Key:   STRING_SHORT,
					Value: queueName,
				},
				{
					Key:   INT_LONG,
					Value: messageCount,
				},
				{
					Key:   INT_LONG,
					Value: consumerCount,
				},
			},
		},
	}.FormatMethodFrame()
	return frame
}

func createQueueBindOkFrame(request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(QUEUE_BIND_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}

func createQueueDeleteOkFrame(request *RequestMethodMessage, messageCount uint32) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(QUEUE_DELETE_OK),
		Content: ContentList{
			KeyValuePairs: []KeyValue{
				{
					Key:   INT_LONG,
					Value: messageCount,
				},
			},
		},
	}.FormatMethodFrame()
	return frame
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
func parseQueueDeclareFrame(payload []byte) (*RequestMethodMessage, error) {
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
	flags := DecodeFlags(octet, []string{"passive", "durable", "exclusive", "autoDelete", "noWait"}, true)
	durable := flags["durable"]
	exclusive := flags["exclusive"]
	autoDelete := flags["autoDelete"]
	noWait := flags["noWait"]

	var arguments map[string]any
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
	msg := &QueueDeclareMessage{
		QueueName:  queueName,
		Durable:    durable,
		Exclusive:  exclusive,
		AutoDelete: autoDelete,
		NoWait:     noWait,
		Arguments:  arguments,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Debug().Interface("queue", msg).Msg("Queue formatted")
	return request, nil
}

func parseQueueMethod(methodID uint16, payload []byte) (any, error) {
	switch methodID {
	case uint16(QUEUE_DECLARE):
		log.Debug().Msg("Received QUEUE_DECLARE frame")
		return parseQueueDeclareFrame(payload)
	case uint16(QUEUE_DELETE):
		log.Debug().Msg("Received QUEUE_DELETE frame")
		return parseQueueDeleteFrame(payload)
	case uint16(QUEUE_BIND):
		log.Printf("[DEBUG] Received QUEUE_BIND frame \n")
		return parseQueueBindFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

// Fields:
// 0-1: reserved short int
// 2: exchange name - length (short)
// 3: if-unused - (bit)
// 4: no-wait - (bit)
func parseQueueDeleteFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	log.Printf("[DEBUG] Received QUEUE_DELETE frame %x \n", payload)

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
	flags := DecodeFlags(octet, []string{"ifUnused", "noWait"}, true)
	ifUnused := flags["ifUnused"]
	noWait := flags["noWait"]

	msg := &QueueDeleteMessage{
		QueueName: exchangeName,
		IfUnused:  ifUnused,
		NoWait:    noWait,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Debug().Interface("queue", msg).Msg("Queue formatted")
	return request, nil
}

func parseQueueBindFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	log.Printf("[DEBUG] Received QUEUE_BIND frame %x \n", payload)
	buf := bytes.NewReader(payload)
	reserverd1, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserverd1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	queue, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode queue name: %v", err)
	}
	exchange, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	log.Printf("[DEBUG] Exchange name: %s\n", exchange)
	routingKey, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode routing key: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := DecodeFlags(octet, []string{"noWait"}, true)
	noWait := flags["noWait"]
	arguments := make(map[string]any)
	if buf.Len() > 4 {
		argumentsStr, _ := DecodeLongStr(buf)
		arguments, err = DecodeTable([]byte(argumentsStr))
		if err != nil {
			return nil, fmt.Errorf("failed to read arguments: %v", err)
		}
	}
	msg := &QueueBindMessage{
		Queue:      queue,
		Exchange:   exchange,
		RoutingKey: routingKey,
		NoWait:     noWait,
		Arguments:  arguments,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Debug().Interface("queue", msg).Msg("Queue formatted")
	return request, nil
}
