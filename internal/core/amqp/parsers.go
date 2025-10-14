package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/rs/zerolog/log"
)

func parseFrame(frame []byte) (any, error) {
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
	case byte(TYPE_METHOD):
		log.Trace().Uint16("channel", channel).Msg("Received METHOD frame")
		request, err := parseMethodFrame(channel, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse method frame: %v", err)
		}
		return request, nil

	case byte(TYPE_HEADER):
		log.Trace().Uint16("channel", channel).Msg("Received HEADER frame")
		return parseHeaderFrame(channel, payloadSize, payload)

	case byte(TYPE_BODY):
		log.Trace().Uint16("channel", channel).Msg("Received BODY frame")
		return parseBodyFrame(channel, payloadSize, payload)

	case byte(TYPE_HEARTBEAT):
		log.Trace().Msg("Received HEARTBEAT frame")
		return &Heartbeat{}, nil

	default:
		log.Trace().Bytes("frame", frame).Msg("Received unknown frame")
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func parseMethodFrame(channel uint16, payload []byte) (*ChannelState, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short")
	}

	classID := binary.BigEndian.Uint16(payload[0:2])
	methodID := binary.BigEndian.Uint16(payload[2:4])
	methodPayload := payload[4:]

	switch classID {
	case uint16(CONNECTION):
		log.Trace().Msgf("Received CONNECTION frame on channel %d\n", channel)
		startOkFrame, err := parseConnectionMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}

		request := &RequestMethodMessage{
			Channel:  channel,
			ClassID:  classID,
			MethodID: methodID,
			Content:  startOkFrame,
		}
		state := &ChannelState{
			MethodFrame: request,
		}
		return state, nil

	case uint16(CHANNEL):
		log.Trace().Msgf("Received CHANNEL frame on channel %d\n", channel)
		request, err := parseChannelMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	case uint16(EXCHANGE):
		log.Trace().Msgf("Received EXCHANGE frame on channel %d\n", channel)
		request, err := parseExchangeMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	case uint16(QUEUE):
		log.Trace().Msgf("Received QUEUE frame on channel %d\n", channel)
		request, err := parseQueueMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	case uint16(BASIC):
		log.Debug().Uint16("channel", channel).Msg("Received BASIC frame")
		request, err := parseBasicMethod(methodID, methodPayload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			msg, ok := request.(*RequestMethodMessage)
			if ok {
				msg.Channel = channel
				msg.ClassID = classID
				msg.MethodID = methodID
				state := &ChannelState{
					MethodFrame: msg,
				}
				return state, nil
			}
		}
		return nil, nil

	default:
		log.Warn().Uint16("class_id", classID).Msg("Unknown class ID")
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func parseHeaderFrame(channel uint16, payloadSize uint32, payload []byte) (*ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	log.Trace().Uint32("payload_size", payloadSize).Msg("Received payload size")

	classID := binary.BigEndian.Uint16(payload[0:2])

	switch classID {

	case uint16(BASIC):
		log.Debug().Uint16("channel", channel).Msg("Received BASIC HEADER frame")
		request, err := parseBasicHeader(payload)
		if err != nil {
			return nil, err
		}
		if request != nil {
			request.Channel = channel
			state := &ChannelState{
				HeaderFrame: request,
			}
			return state, nil
		}
		return nil, nil

	default:
		log.Warn().Uint16("class_id", classID).Msg("Unknown class ID")
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func parseBodyFrame(channel uint16, payloadSize uint32, payload []byte) (*ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	log.Trace().Uint32("payload_size", payloadSize).Msg("Received payload size")

	state := &ChannelState{
		Body: payload,
	}
	return state, nil

}

// REGION Exchange_Methods
func parseExchangeMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(EXCHANGE_DECLARE):
		log.Debug().Msg("Received EXCHANGE_DECLARE frame \n")
		return parseExchangeDeclareFrame(payload)
	case uint16(EXCHANGE_DELETE):
		log.Debug().Msg("Received EXCHANGE_DELETE frame \n")
		return parseExchangeDeleteFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

// REGION Connection_Methods

func parseConnectionMethod(methodID uint16, payload []byte) (any, error) {
	switch methodID {
	case uint16(CONNECTION_START):
		log.Debug().Msg("Received CONNECTION_START frame")
		return parseConnectionStartFrame(payload)

	case uint16(CONNECTION_START_OK):
		log.Trace().Bytes("payload", payload).Msg("Received connection.start-ok")
		return parseConnectionStartOkFrame(payload)

	case uint16(CONNECTION_TUNE):
		log.Debug().Msg("Received CONNECTION_TUNE frame")
		tuneResponse, err := parseConnectionTuneFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection.tune frame: %v", err)
		}
		tuneOkRequest := createConnectionTuneOkFrame(fineTune(tuneResponse))
		return tuneOkRequest, nil

	case uint16(CONNECTION_TUNE_OK):
		log.Debug().Msg("Received CONNECTION_TUNE_OK frame")
		return parseConnectionTuneOkFrame(payload)

	case uint16(CONNECTION_OPEN):
		log.Debug().Msg("Received CONNECTION_OPEN frame")
		return parseConnectionOpenFrame(payload)

	case uint16(CONNECTION_OPEN_OK):
		log.Debug().Msg("Received CONNECTION_OPEN_OK frame")
		return parseConnectionOpenOkFrame(payload)

	case uint16(CONNECTION_CLOSE):
		request, err := parseConnectionCloseFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection.close frame: %v", err)
		}
		request.MethodID = methodID
		content, ok := request.Content.(*ConnectionCloseMessage)
		if !ok {
			return nil, fmt.Errorf("invalid message content type")
		}
		log.Printf("[TRACE] Received CONNECTION_CLOSE: (%d) '%s'", content.ReplyCode, content.ReplyText)
		return request, nil

	case uint16(CONNECTION_CLOSE_OK):
		request, err := parseConnectionCloseOkFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CONNECTION_CLOSE_OK frame: %v", err)
		}
		request.MethodID = methodID
		log.Debug().Msg("Received CONNECTION_CLOSE_OK")
		return request, nil

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseConnectionCloseFrame(payload []byte) (*RequestMethodMessage, error) {
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
	msg := &ConnectionCloseMessage{
		ReplyCode: replyCode,
		ReplyText: replyText,
		ClassID:   classID,
		MethodID:  methodID,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Trace().Msgf("connection close request: %+v\n", request)
	return request, nil
}

func parseConnectionCloseOkFrame(payload []byte) (*RequestMethodMessage, error) {
	return &RequestMethodMessage{Content: payload}, nil
}

func parseConnectionOpenFrame(payload []byte) (interface{}, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short")
	}
	buf := bytes.NewReader(payload)
	virtualHost, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode virtual host: %v", err)
	}
	return &ConnectionOpen{
		VirtualHost: virtualHost,
	}, nil
}

func parseConnectionOpenOkFrame(payload []byte) (interface{}, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short")
	}
	// buf := bytes.NewReader(payload)
	return nil, nil
}

func parseConnectionTuneFrame(payload []byte) (*ConnectionTune, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	buf := bytes.NewReader(payload)
	channelMax, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode channel max: %v", err)
	}

	frameMax, err := DecodeLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode frame max: %v", err)
	}
	heartbeat, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode heartbeat: %v", err)
	}
	return &ConnectionTune{
		ChannelMax: channelMax,
		FrameMax:   frameMax,
		Heartbeat:  heartbeat,
	}, nil
}

func parseConnectionTuneOkFrame(payload []byte) (interface{}, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	buf := bytes.NewReader(payload)
	channelMax, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode channel max: %v", err)
	}

	frameMax, err := DecodeLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode frame max: %v", err)
	}
	heartbeat, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode heartbeat: %v", err)
	}
	return &ConnectionTune{
		ChannelMax: channelMax,
		FrameMax:   frameMax,
		Heartbeat:  heartbeat,
	}, nil
}

func parseConnectionStartFrame(payload []byte) (*ConnectionStartFrame, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)

	// Decode version-major and version-minor
	versionMajor, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode version-major: %v", err)
	}
	log.Trace().Msgf("Version Major: %d\n", versionMajor)

	versionMinor, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode version-minor: %v", err)
	}
	log.Trace().Msgf("Version Minor: %d\n", versionMinor)

	/***Server Properties*/
	serverPropertiesStr, err := DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	serverProperties, err := DecodeTable([]byte(serverPropertiesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	log.Trace().Msgf("Server Properties: %+v\n", serverProperties)

	/***Mechanisms*/
	mechanismsStr, err := DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mechanism: %v", err)
	}
	log.Trace().Msgf("Mechanisms: %s\n", mechanismsStr)

	/***Locales*/
	localesStr, err := DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode locale: %v", err)
	}
	log.Trace().Msgf("Locales: %s\n", localesStr)

	return &ConnectionStartFrame{
		VersionMajor:     versionMajor,
		VersionMinor:     versionMinor,
		ServerProperties: serverProperties,
		Mechanisms:       mechanismsStr,
		Locales:          localesStr,
	}, nil
}

func parseConnectionStartOkFrame(payload []byte) (*ConnectionStartOk, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)

	clientPropertiesStr, err := DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}

	clientProperties, err := DecodeTable([]byte(clientPropertiesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}

	mechanism, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mechanism: %v", err)
	}

	if mechanism != "PLAIN" {
		return nil, fmt.Errorf("unsupported mechanism: %s", mechanism)
	}

	security, err := DecodeSecurityPlain(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode security: %v", err)
	}

	locale, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode locale: %v", err)
	}

	return &ConnectionStartOk{
		ClientProperties: clientProperties,
		Mechanism:        mechanism,
		Response:         security,
		Locale:           locale,
	}, nil
}

// REGION Channel_Methods
func parseChannelMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(CHANNEL_OPEN):
		log.Debug().Msg("Received CHANNEL_OPEN frame")
		return parseChannelOpenFrame(payload)

	case uint16(CHANNEL_OPEN_OK):
		log.Debug().Msg("Received CHANNEL_OPEN_OK frame")
		return parseChannelOpenOkFrame(payload)

	case uint16(CHANNEL_CLOSE):
		log.Debug().Msg("Received CHANNEL_CLOSE frame")
		return parseChannelCloseFrame(payload)

	case uint16(CHANNEL_CLOSE_OK):
		log.Debug().Msg("Received CHANNEL_CLOSE_OK frame")
		return parseChannelCloseOkFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseChannelOpenFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short")
	}
	request := &RequestMethodMessage{
		Content: nil,
	}

	return request, nil
}

func parseChannelOpenOkFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short")
	}
	request := &RequestMethodMessage{
		Content: nil,
	}

	return request, nil
}

func parseChannelCloseFrame(payload []byte) (*RequestMethodMessage, error) {
	log.Trace().Msgf("Received CHANNEL_CLOSE frame: %x \n", payload)
	if len(payload) < 6 {
		return nil, fmt.Errorf("frame too short")
	}

	buf := bytes.NewReader(payload)
	replyCode, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding reply code: %v", err)
	}
	replyText, err := DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding reply text: %v", err)
	}
	classID, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding class id: %v", err)
	}
	methodID, err := DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding method id: %v", err)
	}

	msg := &ChannelCloseMessage{
		ReplyCode: replyCode,
		ReplyText: replyText,
		ClassID:   classID,
		MethodID:  methodID,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	return request, nil
}

func parseChannelCloseOkFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) > 1 {
		return nil, fmt.Errorf("unexxpected payload length")
	}
	request := &RequestMethodMessage{
		Content: nil,
	}

	return request, nil
}

// REGION Basic_headers
