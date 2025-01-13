package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp/message"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/utils"
)

const (
	FRAME_END = 0xCE
)

type ChannelState struct {
	MethodFrame *RequestMethodMessage
	HeaderFrame *HeaderFrame
	Body        []byte
	BodySize    uint64
}

type HeaderFrame struct {
	Channel    uint16
	ClassID    uint16
	BodySize   uint64
	Properties *message.BasicProperties
}

type RequestMethodMessage struct {
	Channel  uint16
	ClassID  uint16
	MethodID uint16
	Content  interface{}
}

type ResponseMethodMessage struct {
	Channel  uint16
	ClassID  uint16
	MethodID uint16
	Content  ContentList
}

type ResponseContent struct {
	Channel uint16
	ClassID uint16
	Weight  uint16
	Message Message
}

type Message struct {
	ID         string                  `json:"id"`
	Body       []byte                  `json:"body"`
	Properties message.BasicProperties `json:"properties"`
	Exchange   string                  `json:"exchange"`
	RoutingKey string                  `json:"routing_key"`
}

type ContentList struct {
	KeyValuePairs []KeyValue
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func (msg ResponseContent) FormatHeaderFrame() []byte {
	frameType := uint8(constants.TYPE_HEADER)
	var payloadBuf bytes.Buffer
	channel := msg.Channel
	classID := msg.ClassID

	weight := msg.Weight
	bodySize := len(msg.Message.Body)
	flag_list, flags, err := msg.Message.Properties.EncodeBasicProperties()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	binary.Write(&payloadBuf, binary.BigEndian, uint16(classID))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(weight))
	binary.Write(&payloadBuf, binary.BigEndian, uint64(bodySize))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(flags))
	payloadBuf.Write(flag_list)

	payloadSize := uint32(payloadBuf.Len())

	headerBuf := FormatHeader(frameType, channel, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)
	frame = append(frame, FRAME_END)
	return frame
}

func (msg ResponseContent) FormatBodyFrame() []byte {
	frameType := uint8(constants.TYPE_BODY)
	var payloadBuf bytes.Buffer
	channel := msg.Channel
	content := msg.Message.Body
	payloadBuf.Write(content)

	payloadSize := uint32(payloadBuf.Len())
	headerBuf := FormatHeader(frameType, channel, payloadSize)
	frame := append(headerBuf, payloadBuf.Bytes()...)
	frame = append(frame, FRAME_END)
	return frame
}

func (msg ResponseMethodMessage) FormatMethodFrame() []byte {
	var payloadBuf bytes.Buffer
	class := msg.ClassID
	method := msg.MethodID
	channelNum := msg.Channel

	binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	methodPayload := formatMethodPayload(msg.Content)
	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(constants.TYPE_METHOD) // METHOD frame type
	headerBuf := FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, FRAME_END) // frame-end

	return frame
}

func formatMethodPayload(content ContentList) []byte {
	var payloadBuf bytes.Buffer
	for _, kv := range content.KeyValuePairs {
		if kv.Key == INT_OCTET {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint8))
		} else if kv.Key == INT_SHORT {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint16))
		} else if kv.Key == INT_LONG {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint32))
		} else if kv.Key == INT_LONG_LONG {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(uint64))
		} else if kv.Key == BIT {
			if kv.Value.(bool) {
				payloadBuf.WriteByte(1)
			} else {
				payloadBuf.WriteByte(0)
			}
		} else if kv.Key == STRING_SHORT {
			payloadBuf.Write(utils.EncodeShortStr(kv.Value.(string)))
		} else if kv.Key == STRING_LONG {
			payloadBuf.Write(utils.EncodeLongStr(kv.Value.([]byte)))
		} else if kv.Key == TIMESTAMP {
			binary.Write(&payloadBuf, binary.BigEndian, kv.Value.(int64))
		} else if kv.Key == TABLE {
			encodedTable := utils.EncodeTable(kv.Value.(map[string]interface{}))
			payloadBuf.Write(utils.EncodeLongStr(encodedTable))
		}
	}
	return payloadBuf.Bytes()
}

func FormatHeader(frameType uint8, channel uint16, payloadSize uint32) []byte {
	header := make([]byte, 7)
	header[0] = frameType
	binary.BigEndian.PutUint16(header[1:3], channel)
	binary.BigEndian.PutUint32(header[3:7], uint32(payloadSize))
	return header
}

func EncodeGetOkToContentList(msg *message.BasicGetOk) *ContentList {
	KeyValuePairs := []KeyValue{
		{ // delivery_tag
			Key:   INT_LONG_LONG,
			Value: msg.DeliveryTag,
		},
		{ // redelivered
			Key:   BIT,
			Value: msg.Redelivered,
		},
		{ // exchange
			Key:   STRING_SHORT,
			Value: msg.Exchange,
		},
		{ // routing_key
			Key:   STRING_SHORT,
			Value: msg.RoutingKey,
		},
		{ // message_count
			Key:   INT_LONG,
			Value: msg.MessageCount,
		},
	}
	contentList := &ContentList{KeyValuePairs: KeyValuePairs}
	return contentList
}
