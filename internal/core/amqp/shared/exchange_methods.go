package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

func parseExchangeMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(amqp.EXCHANGE_DECLARE):
		fmt.Printf("[DEBUG] Received EXCHANGE_DECLARE frame \n")
		return parseExchangeDeclareFrame(payload)
	case uint16(amqp.EXCHANGE_DELETE):
		fmt.Printf("[DEBUG] Received EXCHANGE_DELETE frame \n")
		return parseExchangeDeleteFrame(payload)

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
func parseExchangeDeclareFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
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
	exchangeName, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange name: %v", err)
	}
	exchangeType, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange type: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := utils.DecodeExchangeDeclareFlags(octet)
	autoDelete := flags["autoDelete"]
	durable := flags["durable"]
	internal := flags["internal"]
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
	msg := &amqp.ExchangeDeclareMessage{
		ExchangeName: exchangeName,
		ExchangeType: exchangeType,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Internal:     internal,
		NoWait:       noWait,
		Arguments:    arguments,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Exchange fomated: %+v \n", msg)
	return request, nil
}

// Fields:
// 0-1: reserved short int
// 2: exchange name - length (short)
// 3: if-unused - (bit)
// 4: no-wait - (bit)
func parseExchangeDeleteFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] Received EXCHANGE_DELETE frame %x \n", payload)

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
	flags := utils.DecodeExchangeDeclareFlags(octet)
	ifUnused := flags["ifUnused"]
	noWait := flags["noWait"]

	msg := &amqp.ExchangeDeleteMessage{
		ExchangeName: exchangeName,
		IfUnused:     ifUnused,
		NoWait:       noWait,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Exchange fomated: %+v \n", msg)
	return request, nil
}
