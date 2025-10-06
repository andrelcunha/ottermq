package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
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
		log.Printf("[TRACE] Received HEARTBEAT frame on channel %d\n", channel)
		return &Heartbeat{}, nil

	default:
		log.Printf("[TRACE] Received: %x\n", frame)
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
		log.Printf("[DEBUG] Received CONNECTION frame on channel %d\n", channel)
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
		log.Printf("[DEBUG] Received CHANNEL frame on channel %d\n", channel)
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
		log.Printf("[DEBUG] Received EXCHANGE frame on channel %d\n", channel)
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
		log.Printf("[DEBUG] Received QUEUE frame on channel %d\n", channel)
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
		log.Printf("[DEBUG] Received BASIC frame on channel %d\n", channel)
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
		log.Printf("[DEBUG] Unknown class ID: %d\n", classID)
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func parseHeaderFrame(channel uint16, payloadSize uint32, payload []byte) (*ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	log.Printf("[DEBUG] payload size: %d\n", payloadSize)

	classID := binary.BigEndian.Uint16(payload[0:2])

	switch classID {

	case uint16(BASIC):
		log.Printf("[DEBUG] Received BASIC HEADER frame on channel %d\n", channel)
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
		log.Printf("[DEBUG] Unknown class ID: %d\n", classID)
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func parseBodyFrame(channel uint16, payloadSize uint32, payload []byte) (*ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	log.Printf("[DEBUG] payload size: %d\n", payloadSize)

	log.Printf("[DEBUG] Received BASIC HEADER frame on channel %d\n", channel)

	state := &ChannelState{
		Body: payload,
	}
	return state, nil

}

// REGION Exchange_Methods
func parseExchangeMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(EXCHANGE_DECLARE):
		log.Printf("[DEBUG] Received EXCHANGE_DECLARE frame \n")
		return parseExchangeDeclareFrame(payload)
	case uint16(EXCHANGE_DELETE):
		log.Printf("[DEBUG] Received EXCHANGE_DELETE frame \n")
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
func parseExchangeDeclareFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)
	reserved1, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserved1 != 0 {
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
	// flags := utils.DecodeExchangeDeclareFlags(octet)
	flags := utils.DecodeFlags(octet, []string{"passive", "durable", "autoDelete", "internal", "noWait"}, true)
	passive := flags["passive"]
	durable := flags["durable"]
	autoDelete := flags["autoDelete"]
	internal := flags["internal"]
	noWait := flags["noWait"]

	var arguments map[string]any
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
	msg := &ExchangeDeclareMessage{
		ExchangeName: exchangeName,
		ExchangeType: exchangeType,
		Passive:      passive,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Internal:     internal,
		NoWait:       noWait,
		Arguments:    arguments,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Printf("[DEBUG] Exchange fomated: %+v \n", msg)
	return request, nil
}

// Fields:
// 0-1: reserved short int
// 2: exchange name - length (short)
// 3: if-unused - (bit)
// 4: no-wait - (bit)
func parseExchangeDeleteFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	log.Printf("[DEBUG] Received EXCHANGE_DELETE frame %x \n", payload)

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
	// flags := utils.DecodeExchangeDeclareFlags(octet)
	flags := utils.DecodeFlags(octet, []string{"ifUnused", "noWait"}, true)
	ifUnused := flags["ifUnused"]
	noWait := flags["noWait"]

	msg := &ExchangeDeleteMessage{
		ExchangeName: exchangeName,
		IfUnused:     ifUnused,
		NoWait:       noWait,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Printf("[DEBUG] Exchange fomated: %+v \n", msg)
	return request, nil
}

// REGION Connection_Methods

func parseConnectionMethod(methodID uint16, payload []byte) (any, error) {
	switch methodID {
	case uint16(CONNECTION_START):
		log.Printf("[DEBUG] Received CONNECTION_START frame \n")
		return parseConnectionStartFrame(payload)

	case uint16(CONNECTION_START_OK):
		log.Printf("[DEBUG] Received connection.start-ok: %x\n", payload)
		return parseConnectionStartOkFrame(payload)

	case uint16(CONNECTION_TUNE):
		log.Printf("[DEBUG] Received CONNECTION_TUNE_OK frame \n")
		tuneResponse, err := parseConnectionTuneFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection.tune frame: %v", err)
		}
		tuneOkRequest := createConnectionTuneOkFrame(fineTune(tuneResponse))
		return tuneOkRequest, nil

	case uint16(CONNECTION_TUNE_OK):
		log.Printf("[DEBUG] Received CONNECTION_TUNE_OK frame \n")
		return parseConnectionTuneOkFrame(payload)

	case uint16(CONNECTION_OPEN):
		log.Printf("[DEBUG] Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenFrame(payload)

	case uint16(CONNECTION_OPEN_OK):
		log.Printf("[DEBUG] Received CONNECTION_OPEN frame \n")
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
		log.Printf("[TRACE] Received CONNECTION_CLOSE_OK")
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
	log.Printf("[DEBUG] connection close request: %+v\n", request)
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
	virtualHost, err := utils.DecodeShortStr(buf)
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
	channelMax, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode channel max: %v", err)
	}

	frameMax, err := utils.DecodeLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode frame max: %v", err)
	}
	heartbeat, err := utils.DecodeShortInt(buf)
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
	channelMax, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode channel max: %v", err)
	}

	frameMax, err := utils.DecodeLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode frame max: %v", err)
	}
	heartbeat, err := utils.DecodeShortInt(buf)
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
	// log.Printf("[DEBUG] Version Major: %d\n", versionMajor)

	versionMinor, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode version-minor: %v", err)
	}
	// log.Printf("[DEBUG] Version Minor: %d\n", versionMinor)

	/***Server Properties*/
	serverPropertiesStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	serverProperties, err := utils.DecodeTable([]byte(serverPropertiesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	// log.Printf("[DEBUG] Server Properties: %+v\n", serverProperties)

	/***Mechanisms*/
	mechanismsStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mechanism: %v", err)
	}
	// log.Printf("[DEBUG] Mechanisms Srt: %s\n", mechanismsStr)

	/***Locales*/
	localesStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode locale: %v", err)
	}
	// log.Printf("[DEBUG] Locales Srt: %s\n", localesStr)

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

	clientPropertiesStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}

	clientProperties, err := utils.DecodeTable([]byte(clientPropertiesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}

	mechanism, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mechanism: %v", err)
	}

	if mechanism != "PLAIN" {
		return nil, fmt.Errorf("unsupported mechanism: %s", mechanism)
	}

	security, err := utils.DecodeSecurityPlain(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode security: %v", err)
	}

	locale, err := utils.DecodeShortStr(buf)
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
		log.Printf("[DEBUG] Received CHANNEL_OPEN frame")
		return parseChannelOpenFrame(payload)

	case uint16(CHANNEL_OPEN_OK):
		log.Printf("[DEBUG] Received CHANNEL_OPEN_OK frame \n")
		return parseChannelOpenOkFrame(payload)

	case uint16(CHANNEL_CLOSE):
		log.Printf("[DEBUG] Received CHANNEL_CLOSE frame \n")
		return parseChannelCloseFrame(payload)

	case uint16(CHANNEL_CLOSE_OK):
		log.Printf("[DEBUG] Received CHANNEL_CLOSE_OK frame \n")
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
	log.Printf("[DEBUG] Received CHANNEL_CLOSE frame: %x \n", payload)
	if len(payload) < 6 {
		return nil, fmt.Errorf("frame too short")
	}

	buf := bytes.NewReader(payload)
	replyCode, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding reply code: %v", err)
	}
	replyText, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding reply text: %v", err)
	}
	classID, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed decoding class id: %v", err)
	}
	methodID, err := utils.DecodeShortInt(buf)
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

// ClassID: short
// Weight: short
// Body Size: long long
// Properties flags: short
// Properties: long (table)

func parseBasicHeader(headerPayload []byte) (*HeaderFrame, error) {

	buf := bytes.NewReader(headerPayload)
	classID, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode class ID: %v", err)
	}
	log.Printf("[DEBUG] Class ID: %d\n", classID)

	weight, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode weight: %v", err)
	}
	if weight != 0 {
		return nil, fmt.Errorf("weight must be 0")
	}
	log.Printf("[DEBUG] Weight: %d\n", weight)

	bodySize, err := utils.DecodeLongLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode body size: %v", err)
	}
	log.Printf("[DEBUG] Body Size: %d\n", bodySize)

	shortFlags, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flags: %v", err)
	}
	flags := decodeBasicHeaderFlags(shortFlags)
	properties, err := createContentPropertiesTable(flags, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode properties: %v", err)
	}
	log.Printf("[DEBUG] properties: %v\n", properties)
	header := &HeaderFrame{
		ClassID:    classID,
		BodySize:   bodySize,
		Properties: properties,
	}
	return header, nil
}

// REGION Basic_Methods

func parseBasicMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	// case uint16(BASIC_ACK):
	// 	log.Printf("[DEBUG] Received BASIC_ACK frame \n")
	// 	return parseBasicAckFrame(payload)

	// case uint16(BASIC_REJECT):
	// 	log.Printf("[DEBUG] Received BASIC_REJECT frame \n")
	// 	return parseBasicRejectFrame(payload)

	case uint16(BASIC_PUBLISH):
		log.Printf("[DEBUG] Received BASIC_PUBLISH frame \n")
		return parseBasicPublishFrame(payload)

	// case uint16(BASIC_RETURN):
	// 	log.Printf("[DEBUG] Received BASIC_RETURN frame \n")
	// 	return parseBasicReturnFrame(payload)

	// case uint16(BASIC_DELIVER):
	// 	log.Printf("[DEBUG] Received BASIC_DELIVER frame \n")
	// 	return parseBasicDeliverFrame(payload)

	case uint16(BASIC_GET):
		log.Printf("[DEBUG] Received BASIC_GET frame \n")
		return parseBasicGetFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseBasicPublishFrame(payload []byte) (*RequestMethodMessage, error) {
	if len(payload) < 10 {
		return nil, fmt.Errorf("payload too short")
	}

	buf := bytes.NewReader(payload)
	reserved1, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserved1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	exchange, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange: %v", err)
	}
	routingKey, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode routing key: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := decodeBasicPublishFlags(octet)
	mandatory := flags["mandatory"]
	immediate := flags["immediate"]
	msg := &BasicPublishMessage{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Mandatory:  mandatory,
		Immediate:  immediate,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	log.Printf("[DEBUG] BasicPublish fomated: %+v \n", msg)
	return request, nil
}

func parseBasicGetFrame(payload []byte) (*RequestMethodMessage, error) {
	buf := bytes.NewReader(payload)
	reserved1, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reserved1: %v", err)
	}
	if reserved1 != 0 {
		return nil, fmt.Errorf("reserved1 must be 0")
	}
	queue, err := utils.DecodeShortStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exchange: %v", err)
	}
	octet, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read octet: %v", err)
	}
	flags := decodeBasicGetFlags(octet)
	noAck := flags["noAck"]
	msg := &BasicGetMessage{
		Reserved1: reserved1,
		Queue:     queue,
		NoAck:     noAck,
	}
	return &RequestMethodMessage{
		Content: msg,
	}, nil
}
