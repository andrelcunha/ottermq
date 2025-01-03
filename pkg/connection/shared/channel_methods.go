package shared

import (
	"encoding/binary"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp/message"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

func parseChannelMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(constants.CHANNEL_OPEN):
		fmt.Printf("Received CHANNEL_OPEN frame \n")
		return parseChannelOpenFrame(payload)

	case uint16(constants.CHANNEL_OPEN_OK):
		fmt.Printf("Received CHANNEL_OPEN_OK frame \n")
		return parseChannelOpenOkFrame(payload)

	case uint16(constants.CHANNEL_CLOSE):
		fmt.Printf("Received CHANNEL_OPEN frame \n")
		return parseChannelCloseFrame(payload)

	case uint16(constants.CHANNEL_CLOSE_OK):
		fmt.Printf("Received CHANNEL_CLOSE_OK frame \n")
		return parseChannelCloseOkFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseChannelOpenFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short")
	}
	request := &amqp.RequestMethodMessage{
		Content: nil,
	}

	return request, nil
}

func parseChannelOpenOkFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short")
	}
	request := &amqp.RequestMethodMessage{
		Content: nil,
	}

	return request, nil
}

func parseChannelCloseFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) < 12 {
		return nil, fmt.Errorf("frame too short")
	}

	index := uint16(0)
	replyCode := binary.BigEndian.Uint16(payload[index : index+2])
	index += 2
	replyTextLen := payload[index]
	index++
	replyText := string(payload[index : index+uint16(replyTextLen)])
	index += uint16(replyTextLen)
	classID := binary.BigEndian.Uint16(payload[index : index+2])
	index += 2
	methodID := binary.BigEndian.Uint16(payload[index : index+2])
	msg := &message.ChannelCloseMessage{
		ReplyCode: replyCode,
		ReplyText: replyText,
		ClassID:   classID,
		MethodID:  methodID,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] channel close request: %+v\n", request)
	return request, nil
}

func parseChannelCloseOkFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	if len(payload) > 1 {
		return nil, fmt.Errorf("unexxpected payload length")
	}
	request := &amqp.RequestMethodMessage{
		Content: nil,
	}

	return request, nil
}
