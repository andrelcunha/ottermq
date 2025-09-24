package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

func parseFrame(configurations *map[string]any, conn net.Conn, currentChannel uint16, frame []byte) (any, error) {
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
		log.Printf("[DEBUG] Received METHOD frame on channel %d\n", channel)
		request, err := parseMethodFrame(configurations, channel, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to parse method frame: %v", err)
		}
		return request, nil

	case byte(TYPE_HEADER):
		fmt.Printf("Received HEADER frame on channel %d\n", channel)
		return parseHeaderFrame(channel, payloadSize, payload)

	case byte(TYPE_BODY):
		log.Printf("[DEBUG] Received BODY frame on channel %d\n", channel)
		return parseBodyFrame(channel, payloadSize, payload)

	case byte(TYPE_HEARTBEAT):
		log.Printf("[DEBUG] Received HEARTBEAT frame on channel %d\n", channel)
		return nil, nil

	default:
		log.Printf("[DEBUG] Received: %x\n", frame)
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func parseMethodFrame(configurations *map[string]interface{}, channel uint16, payload []byte) (*ChannelState, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short")
	}

	classID := binary.BigEndian.Uint16(payload[0:2])
	methodID := binary.BigEndian.Uint16(payload[2:4])
	methodPayload := payload[4:]

	switch classID {
	case uint16(CONNECTION):
		fmt.Printf("[DEBUG] Received CONNECTION frame on channel %d\n", channel)
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
		fmt.Printf("[DEBUG] Received CHANNEL frame on channel %d\n", channel)
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
		fmt.Printf("[DEBUG] Received EXCHANGE frame on channel %d\n", channel)
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
		fmt.Printf("[DEBUG] Received QUEUE frame on channel %d\n", channel)
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
		fmt.Printf("[DEBUG] Received BASIC frame on channel %d\n", channel)
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
		fmt.Printf("[DEBUG] Unknown class ID: %d\n", classID)
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func parseHeaderFrame(channel uint16, payloadSize uint32, payload []byte) (*ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] payload size: %d\n", payloadSize)

	classID := binary.BigEndian.Uint16(payload[0:2])

	switch classID {

	case uint16(BASIC):
		fmt.Printf("[DEBUG] Received BASIC HEADER frame on channel %d\n", channel)
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
		fmt.Printf("[DEBUG] Unknown class ID: %d\n", classID)
		return nil, fmt.Errorf("unknown class ID: %d", classID)
	}
}

func parseBodyFrame(channel uint16, payloadSize uint32, payload []byte) (*ChannelState, error) {

	if len(payload) < int(payloadSize) {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("[DEBUG] payload size: %d\n", payloadSize)

	// headerPayload := payload[2:]

	fmt.Printf("[DEBUG] Received BASIC HEADER frame on channel %d\n", channel)

	state := &ChannelState{
		Body: payload,
	}
	return state, nil

}

// REGION Queue_Methods

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
	msg := &QueueDeclareMessage{
		QueueName: queueName,
		Durable:   durable,
		IfUnused:  ifUnused,
		Exclusive: exclusive,
		NoWait:    noWait,
		Arguments: arguments,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Queue fomated: %+v \n", msg)
	return request, nil
}

func parseQueueMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(QUEUE_DECLARE):
		fmt.Printf("[DEBUG] Received QUEUE_DECLARE frame \n")
		return parseQueueDeclareFrame(payload)
	case uint16(QUEUE_DELETE):
		fmt.Printf("[DEBUG] Received QUEUE_DELETE frame \n")
		return parseQueueDeleteFrame(payload)
	case uint16(QUEUE_BIND):
		fmt.Printf("[DEBUG] Received QUEUE_BIND frame \n")
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

	msg := &QueueDeleteMessage{
		QueueName: exchangeName,
		IfUnused:  ifUnused,
		NoWait:    noWait,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Queue fomated: %+v \n", msg)
	return request, nil
}

func parseQueueBindFrame(payload []byte) (*RequestMethodMessage, error) {
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
	fmt.Printf("[DEBUG] Queue fomated: %+v \n", msg)
	return request, nil
}

// REGION Exchange_Methods
func parseExchangeMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(EXCHANGE_DECLARE):
		fmt.Printf("[DEBUG] Received EXCHANGE_DECLARE frame \n")
		return parseExchangeDeclareFrame(payload)
	case uint16(EXCHANGE_DELETE):
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
func parseExchangeDeclareFrame(payload []byte) (*RequestMethodMessage, error) {
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
	msg := &ExchangeDeclareMessage{
		ExchangeName: exchangeName,
		ExchangeType: exchangeType,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Internal:     internal,
		NoWait:       noWait,
		Arguments:    arguments,
	}
	request := &RequestMethodMessage{
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
func parseExchangeDeleteFrame(payload []byte) (*RequestMethodMessage, error) {
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

	msg := &ExchangeDeleteMessage{
		ExchangeName: exchangeName,
		IfUnused:     ifUnused,
		NoWait:       noWait,
	}
	request := &RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] Exchange fomated: %+v \n", msg)
	return request, nil
}

// REGION Connection_Methods

func parseConnectionMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(CONNECTION_START):
		fmt.Printf("Received CONNECTION_START frame \n")
		return parseConnectionStartFrame(payload)

	case uint16(CONNECTION_START_OK):
		fmt.Printf("[DEBUG] Received connection.start-ok: %x\n", payload)
		return parseConnectionStartOkFrame(payload)

	case uint16(CONNECTION_TUNE):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		tuneResponse, err := parseConnectionTuneFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse connection.tune frame: %v", err)
		}
		tuneOkRequest := createConnectionTuneOkFrame(fineTune(tuneResponse))
		return tuneOkRequest, nil

	case uint16(CONNECTION_TUNE_OK):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		return parseConnectionTuneOkFrame(payload)

	case uint16(CONNECTION_OPEN):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenFrame(payload)

	case uint16(CONNECTION_OPEN_OK):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenOkFrame(payload)

	case uint16(CONNECTION_CLOSE):
		fmt.Printf("Received CONNECTION_CLOSE frame \n")
		request, err := parseConnectionCloseFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse connection.close frame: %v", err)
		}
		request.MethodID = uint16(CONNECTION_CLOSE_OK)
		fmt.Printf("[DEBUG] connection close request: %+v\n", request)
		content, ok := request.Content.(*ConnectionCloseMessage)
		if !ok {
			return nil, fmt.Errorf("Invalid message content type")
		}
		log.Printf("Received connection.close: (%d) '%s'\n", content.ReplyCode, content.ReplyText)
		return request, nil

	case uint16(CONNECTION_CLOSE_OK):
		fmt.Printf("Received CONNECTION_CLOSE_OK frame \n")
		return parseConnectionCloseOkFrame(payload)

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
	fmt.Printf("[DEBUG] connection close request: %+v\n", request)
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
	return &ConnectionOpenFrame{
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

func parseConnectionTuneFrame(payload []byte) (*ConnectionTuneFrame, error) {
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
	return &ConnectionTuneFrame{
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
	return &ConnectionTuneFrame{
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
	// fmt.Printf("Version Major: %d\n", versionMajor)

	versionMinor, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to decode version-minor: %v", err)
	}
	// fmt.Printf("Version Minor: %d\n", versionMinor)

	/***Server Properties*/
	serverPropertiesStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	serverProperties, err := utils.DecodeTable([]byte(serverPropertiesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	// fmt.Printf("Server Properties: %+v\n", serverProperties)

	/***Mechanisms*/
	mechanismsStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mechanism: %v", err)
	}
	// fmt.Printf("Mechanisms Srt: %s\n", mechanismsStr)

	/***Locales*/
	localesStr, err := utils.DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode locale: %v", err)
	}
	// fmt.Printf("Locales Srt: %s\n", localesStr)

	return &ConnectionStartFrame{
		VersionMajor:     versionMajor,
		VersionMinor:     versionMinor,
		ServerProperties: serverProperties,
		Mechanisms:       mechanismsStr,
		Locales:          localesStr,
	}, nil
}

func parseConnectionStartOkFrame(payload []byte) (*ConnectionStartOkFrame, error) {
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

	return &ConnectionStartOkFrame{
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
		fmt.Printf("Received CHANNEL_OPEN frame \n")
		return parseChannelOpenFrame(payload)

	case uint16(CHANNEL_OPEN_OK):
		fmt.Printf("Received CHANNEL_OPEN_OK frame \n")
		return parseChannelOpenOkFrame(payload)

	case uint16(CHANNEL_CLOSE):
		fmt.Printf("Received CHANNEL_CLOSE frame \n")
		return parseChannelCloseFrame(payload)

	case uint16(CHANNEL_CLOSE_OK):
		fmt.Printf("Received CHANNEL_CLOSE_OK frame \n")
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
	fmt.Printf("[DEBUG] Received CHANNEL_CLOSE frame: %x \n", payload)
	if len(payload) < 6 {
		return nil, fmt.Errorf("frame too short")
	}

	buf := bytes.NewReader(payload)
	replyCode, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reply code: %v", err)
	}

	replyText, err := utils.DecodeShortStr(buf)
	classID, err := utils.DecodeShortInt(buf)
	methodID, err := utils.DecodeShortInt(buf)
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
	fmt.Printf("[DEBUG] Class ID: %d\n", classID)

	weight, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode weight: %v", err)
	}
	if weight != 0 {
		return nil, fmt.Errorf("weight must be 0")
	}
	fmt.Printf("[DEBUG] Weight: %d\n", weight)

	bodySize, err := utils.DecodeLongLongInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode body size: %v", err)
	}
	fmt.Printf("[DEBUG] Body Size: %d\n", bodySize)

	shortFlags, err := utils.DecodeShortInt(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode flags: %v", err)
	}
	flags := decodeBasicHeaderFlags(shortFlags)
	properties, err := createContentPropertiesTable(flags, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode properties: %v", err)
	}
	fmt.Printf("[DEBUG] properties: %v\n", properties)
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
	// 	fmt.Printf("[DEBUG] Received BASIC_ACK frame \n")
	// 	return parseBasicAckFrame(payload)

	// case uint16(BASIC_REJECT):
	// 	fmt.Printf("[DEBUG] Received BASIC_REJECT frame \n")
	// 	return parseBasicRejectFrame(payload)

	case uint16(BASIC_PUBLISH):
		fmt.Printf("[DEBUG] Received BASIC_PUBLISH frame \n")
		return parseBasicPublishFrame(payload)

	// case uint16(BASIC_RETURN):
	// 	fmt.Printf("[DEBUG] Received BASIC_RETURN frame \n")
	// 	return parseBasicReturnFrame(payload)

	// case uint16(BASIC_DELIVER):
	// 	fmt.Printf("[DEBUG] Received BASIC_DELIVER frame \n")
	// 	return parseBasicDeliverFrame(payload)

	case uint16(BASIC_GET):
		fmt.Printf("[DEBUG] Received BASIC_GET frame \n")
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
	flags := DecodeBasicPublishFlags(octet)
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
	fmt.Printf("[DEBUG] BasicPublish fomated: %+v \n", msg)
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
	flags := DecodeBasicGetFlags(octet)
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
