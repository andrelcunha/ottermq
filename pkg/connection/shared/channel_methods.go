package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp/message"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	. "github.com/andrelcunha/ottermq/pkg/connection/utils"
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
		fmt.Printf("Received CHANNEL_CLOSE frame \n")
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
	fmt.Printf("[DEBUG] Received CHANNEL_CLOSE frame: %x \n", payload)
	if len(payload) < 6 {
		return nil, fmt.Errorf("frame too short")
	}

	buf := bytes.NewReader(payload)
	replyCode, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reply code: %v", err)
	}

	replyText, err := DecodeShortStr(buf)
	classID, err := DecodeShortInt(buf)
	methodID, err := DecodeShortInt(buf)
	msg := &message.ChannelCloseMessage{
		ReplyCode: replyCode,
		ReplyText: replyText,
		ClassID:   classID,
		MethodID:  methodID,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
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
