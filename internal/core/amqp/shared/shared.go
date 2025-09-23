package shared

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

type Heartbeat struct {
}

type AMQP_Key struct {
	Key  string
	Type string
}

type AMQP_Type struct {
}

func SendProtocolHeader(conn net.Conn) error {
	header := []byte(amqp.AMQP_PROTOCOL_HEADER)
	_, err := conn.Write(header)
	return err
}

func ReadProtocolHeader(conn net.Conn) ([]byte, error) {
	header := make([]byte, 8)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func ReadHeader(conn net.Conn) ([]byte, error) {
	header := make([]byte, 7)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func ReadFrame(conn net.Conn) ([]byte, error) {
	// all frames starts with a 7-octet header
	frameHeader := make([]byte, 7)
	_, err := io.ReadFull(conn, frameHeader)
	if err != nil {
		return nil, err
	}

	// fist octet is the type of frame
	// frameType := binary.BigEndian.Uint16(frameHeader[0:1])

	// 2nd and 3rd octets (short) are the channel number
	// channelNum := binary.BigEndian.Uint16(frameHeader[1:3])

	// 4th to 7th octets (long) are the size of the payload
	payloadSize := binary.BigEndian.Uint32(frameHeader[3:])

	// read the framePayload
	framePayload := make([]byte, payloadSize)
	_, err = io.ReadFull(conn, framePayload)
	if err != nil {
		return nil, err
	}

	// frame-end is a 1-octet after the payload
	frameEnd := make([]byte, 1)
	_, err = io.ReadFull(conn, frameEnd)
	if err != nil {
		return nil, err
	}

	// check if the frame-end is correct (0xCE)
	if frameEnd[0] != 0xCE {
		// return nil, ErrInvalidFrameEnd
		return nil, fmt.Errorf("invalid frame end octet")
	}

	return append(frameHeader, framePayload...), nil
}

func FormatHeader(frameType uint8, channel uint16, payloadSize uint32) []byte {
	header := make([]byte, 7)
	header[0] = frameType
	binary.BigEndian.PutUint16(header[1:3], channel)
	binary.BigEndian.PutUint32(header[3:7], uint32(payloadSize))
	return header
}

func ParseFrame(configurations *map[string]any, conn net.Conn, currentChannel uint16, frame []byte) (any, error) {
	if len(frame) < 7 {
		return nil, fmt.Errorf("frame too short")
	}

	frameType := frame[0]
	channel := binary.BigEndian.Uint16(frame[1:3])
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if len(frame) < int(7+payloadSize) {
		return nil, fmt.Errorf("frame too short")
	}
	payload := frame[7:]

	switch frameType {
	case byte(amqp.TYPE_METHOD):
		log.Printf("[DEBUG] Received METHOD frame on channel %d\n", channel)
		request, err := ParseMethodFrame(configurations, channel, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse method frame: %v", err)
		}
		return request, nil

	case byte(amqp.TYPE_HEADER):
		fmt.Printf("Received HEADER frame on channel %d\n", channel)
		return ParseHeaderFrame(channel, payloadSize, payload)

	case byte(amqp.TYPE_BODY):
		log.Printf("[DEBUG] Received BODY frame on channel %d\n", channel)
		return ParseBodyFrame(channel, payloadSize, payload)

	case byte(amqp.TYPE_HEARTBEAT):
		log.Printf("[DEBUG] Received HEARTBEAT frame on channel %d\n", channel)
		return nil, nil

	default:
		log.Printf("[DEBUG] Received: %x\n", frame)
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func ParseMethodFrame(configurations *map[string]interface{}, channel uint16, payload []byte) (*amqp.ChannelState, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short")
	}

	classID := binary.BigEndian.Uint16(payload[0:2])
	methodID := binary.BigEndian.Uint16(payload[2:4])
	methodPayload := payload[4:]

	switch classID {
	case uint16(amqp.CONNECTION):
		fmt.Printf("[DEBUG] Received CONNECTION frame on channel %d\n", channel)
		startOkFrame, err := parseConnectionMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}

		request := &amqp.RequestMethodMessage{
			Channel:  channel,
			ClassID:  classID,
			MethodID: methodID,
			Content:  startOkFrame,
		}
		state := &amqp.ChannelState{
			MethodFrame: request,
		}
		return state, nil

	case uint16(amqp.CHANNEL):
		fmt.Printf("[DEBUG] Received CHANNEL frame on channel %d\n", channel)
		request, err := parseChannelMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*amqp.RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &amqp.ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	case uint16(amqp.EXCHANGE):
		fmt.Printf("[DEBUG] Received EXCHANGE frame on channel %d\n", channel)
		request, err := parseExchangeMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*amqp.RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &amqp.ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	case uint16(amqp.QUEUE):
		fmt.Printf("[DEBUG] Received QUEUE frame on channel %d\n", channel)
		request, err := parseQueueMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*amqp.RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &amqp.ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	case uint16(amqp.BASIC):
		fmt.Printf("[DEBUG] Received BASIC frame on channel %d\n", channel)
		request, err := parseBasicMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*amqp.RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &amqp.ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	default:
		fmt.Printf("[DEBUG] Unknown class ID: %d\n", classID)
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func SendFrame(conn net.Conn, frame []byte) error {
	fmt.Printf("[DEBUG] Sending frame: %x\n", frame)
	_, err := conn.Write(frame)
	return err
}

func CreateHeartbeatFrame() []byte {
	frame := make([]byte, 8)
	frame[0] = byte(amqp.TYPE_HEARTBEAT)
	binary.BigEndian.PutUint16(frame[1:3], 0)
	binary.BigEndian.PutUint32(frame[3:7], 0)
	frame[7] = 0xCE
	return frame
}

func ParseHeaderFrame(channel uint16, payloadSize uint32, payload []byte) (*amqp.ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] payload size: %d\n", payloadSize)

	classID := binary.BigEndian.Uint16(payload[0:2])

	// headerPayload := payload[2:]

	switch classID {

	case uint16(amqp.BASIC):
		fmt.Printf("[DEBUG] Received BASIC HEADER frame on channel %d\n", channel)
		request, err := parseBasicHeader(payload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			request.Channel = channel
			state := &amqp.ChannelState{
				HeaderFrame: request,
			}
			return state, nil
		}
		return nil, nil

	default:
		fmt.Printf("[DEBUG] Unknown class ID: %d\n", classID)
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func ParseBodyFrame(channel uint16, payloadSize uint32, payload []byte) (*amqp.ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] payload size: %d\n", payloadSize)

	// headerPayload := payload[2:]

	fmt.Printf("[DEBUG] Received BASIC HEADER frame on channel %d\n", channel)

	state := &amqp.ChannelState{
		Body: payload,
	}
	return state, nil

}
