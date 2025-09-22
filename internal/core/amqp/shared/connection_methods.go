package shared

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

type ConnectionStartFrame struct {
	VersionMajor     byte
	VersionMinor     byte
	ServerProperties map[string]interface{}
	Mechanisms       string
	Locales          string
}

type ConnectionStartOkFrame struct {
	ClientProperties map[string]interface{}
	Mechanism        string
	Response         string
	Locale           string
}

type ConnectionTuneFrame struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

type ConnectionOpenFrame struct {
	VirtualHost string
}

func parseConnectionMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(amqp.CONNECTION_START):
		fmt.Printf("Received CONNECTION_START frame \n")
		return parseConnectionStartFrame(payload)

	case uint16(amqp.CONNECTION_START_OK):
		fmt.Printf("[DEBUG] Received connection.start-ok: %x\n", payload)
		return parseConnectionStartOkFrame(payload)

	case uint16(amqp.CONNECTION_TUNE):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		tuneResponse, err := parseConnectionTuneFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse connection.tune frame: %v", err)
		}
		tuneOkRequest := createConnectionTuneOkFrame(fineTune(tuneResponse))
		return tuneOkRequest, nil

	case uint16(amqp.CONNECTION_TUNE_OK):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		return parseConnectionTuneOkFrame(payload)

	case uint16(amqp.CONNECTION_OPEN):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenFrame(payload)

	case uint16(amqp.CONNECTION_OPEN_OK):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenOkFrame(payload)

	case uint16(amqp.CONNECTION_CLOSE):
		fmt.Printf("Received CONNECTION_CLOSE frame \n")
		request, err := parseConnectionCloseFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse connection.close frame: %v", err)
		}
		request.MethodID = uint16(amqp.CONNECTION_CLOSE_OK)
		fmt.Printf("[DEBUG] connection close request: %+v\n", request)
		content, ok := request.Content.(*amqp.ConnectionCloseMessage)
		if !ok {
			return nil, fmt.Errorf("Invalid message content type")
		}
		log.Printf("Received connection.close: (%d) '%s'\n", content.ReplyCode, content.ReplyText)
		return request, nil

	case uint16(amqp.CONNECTION_CLOSE_OK):
		fmt.Printf("Received CONNECTION_CLOSE_OK frame \n")
		return parseConnectionCloseOkFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseConnectionCloseFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
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
	msg := &amqp.ConnectionCloseMessage{
		ReplyCode: replyCode,
		ReplyText: replyText,
		ClassID:   classID,
		MethodID:  methodID,
	}
	request := &amqp.RequestMethodMessage{
		Content: msg,
	}
	fmt.Printf("[DEBUG] connection close request: %+v\n", request)
	return request, nil
}

func parseConnectionCloseOkFrame(payload []byte) (*amqp.RequestMethodMessage, error) {
	return &amqp.RequestMethodMessage{Content: payload}, nil
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

func CreateConnectionStartFrame() []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := amqp.CONNECTION
	methodID := amqp.CONNECTION_START

	payloadBuf.WriteByte(0) // version-major
	payloadBuf.WriteByte(9) // version-minor

	serverProperties := map[string]interface{}{
		"product": "OtterMQ",
	}
	encodedProperties := utils.EncodeTable(serverProperties)
	payloadBuf.Write(utils.EncodeLongStr(encodedProperties))

	payloadBuf.Write(utils.EncodeLongStr([]byte("PLAIN")))

	payloadBuf.Write(utils.EncodeLongStr([]byte("en_US")))

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())

	return frame
}

func CreateConnectionStartOkPayload(configurations *map[string]interface{}, startFrameResponse *ConnectionStartFrame) (ConnectionStartOkFrame, error) {

	clientProperties := (*configurations)["clientProperties"].(map[string]interface{})
	fmt.Printf("[DEBUG] clientProperties: %+v\n", clientProperties)

	mecanismsStr := startFrameResponse.Mechanisms
	mechanismsList := strings.Split(mecanismsStr, " ")
	mechanismsMap := make(map[string]bool)
	for _, m := range mechanismsList {
		mechanismsMap[m] = true
	}
	fmt.Printf("[DEBUG] Server mechanismsMap: %+v\n", mechanismsMap)

	// find 'PLAIN' in mechanismsList
	mechanism := (*configurations)["mechanism"].(string)
	if !mechanismsMap[mechanism] {
		return ConnectionStartOkFrame{}, fmt.Errorf("unsupported mechanism: %s", mechanism)
	}
	fmt.Printf("[DEBUG] mechanism: %s\n", mechanism)

	// fmt.Printf("mechanismsList: %+v\n", mechanismsList)
	username := (*configurations)["username"].(string)
	password := (*configurations)["password"].(string)

	securityStr := fmt.Sprintf(" %s %s", username, password)
	// find 'en_US' in localesList
	localesStr := startFrameResponse.Locales
	localesList := strings.Split(localesStr, ",")
	localesMap := make(map[string]bool)
	for _, l := range localesList {
		localesMap[l] = true
	}
	locale := (*configurations)["locale"].(string)
	if !localesMap[locale] {
		// locale = "en_US"
		return ConnectionStartOkFrame{}, fmt.Errorf("unsupported locale: %s", locale)
	}
	startOk := ConnectionStartOkFrame{
		ClientProperties: clientProperties,
		Mechanism:        mechanism,
		Response:         securityStr,
		Locale:           locale,
	}
	return startOk, nil
}

func CreateConnectionStartOkFrame(startOk *ConnectionStartOkFrame) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := amqp.CONNECTION
	methodID := amqp.CONNECTION_START_OK

	clientProperties := startOk.ClientProperties
	encodedProperties := utils.EncodeTable(clientProperties)
	payloadBuf.Write(utils.EncodeLongStr(encodedProperties))

	payloadBuf.Write(utils.EncodeShortStr(startOk.Mechanism))
	payloadBuf.Write(utils.EncodeSecurityPlain(startOk.Response))

	payloadBuf.Write(utils.EncodeShortStr(startOk.Locale))

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())

	return frame
}

func FormatMethodFrame(channelNum uint16, class amqp.TypeClass, method amqp.TypeMethod, methodPayload []byte) []byte {
	var payloadBuf bytes.Buffer

	binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	// payloadBuf.WriteByte(binary.BigEndian.AppendUint16()[])

	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(amqp.TYPE_METHOD) // METHOD frame type
	headerBuf := FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame
}

func CreateConnectionTuneFrame(tune *ConnectionTuneFrame) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := amqp.CONNECTION
	methodID := amqp.CONNECTION_TUNE

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func createConnectionTuneOkFrame(tune *ConnectionTuneFrame) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := amqp.CONNECTION
	methodID := amqp.CONNECTION_TUNE_OK

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func CreateConnectionOpenFrame(vhost string) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := amqp.CONNECTION
	methodID := amqp.CONNECTION_OPEN

	// Vhost - short string
	payloadBuf.Write(utils.EncodeShortStr(vhost))

	// Reserved-1 (bit) - set to 0
	payloadBuf.WriteByte(0)

	// Reserved-2 (shortstr) - typically empty
	payloadBuf.Write(utils.EncodeShortStr(""))

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func CreateConnectionOpenOkFrame() []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := amqp.CONNECTION
	methodID := amqp.CONNECTION_OPEN_OK

	// Reserved-1 (bit) - set to 0
	payloadBuf.WriteByte(0)

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func fineTune(tune *ConnectionTuneFrame) *ConnectionTuneFrame {
	// TODO: get values from config
	tune.ChannelMax = getSmalestShortInt(2047, tune.ChannelMax)
	tune.FrameMax = getSmalestLongInt(131072, tune.FrameMax)
	tune.Heartbeat = getSmalestShortInt(10, tune.Heartbeat)

	return tune
}

func getSmalestShortInt(a, b uint16) uint16 {
	if a == 0 {
		a = math.MaxUint16
	}
	if b == 0 {
		b = math.MaxInt16
	}
	if a < b {
		return a
	}
	return b
}

func getSmalestLongInt(a, b uint32) uint32 {
	if a == 0 {
		a = math.MaxUint32
	}
	if b == 0 {
		b = math.MaxInt32
	}
	if a < b {
		return a
	}
	return b
}
