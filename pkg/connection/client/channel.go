package client

import (
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

type Channel struct {
	Id uint16
	c  *Client
}

func (ch *Channel) Close() error {
	client := ch.c
	// Remove the channel from the map
	// create a channel close frame
	frame := amqp.ResponseMethodMessage{
		Channel:  ch.Id,
		ClassID:  uint16(constants.CHANNEL),
		MethodID: uint16(constants.CHANNEL_CLOSE),
		Content: amqp.ContentList{
			KeyValuePairs: []amqp.KeyValue{
				{
					Key:   amqp.INT_SHORT,
					Value: uint16(200), // reply code
				},
				{
					Key:   amqp.STRING_SHORT,
					Value: "ok", // reply text
				},
				{
					Key:   amqp.INT_SHORT,
					Value: uint16(0), // class id
				},
				{
					Key:   amqp.INT_SHORT,
					Value: uint16(0), // method id
				},
			},
		},
	}.FormatMethodFrame()
	client.PostFrame(frame)

	return nil
}
