package shared

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
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

type ConnectionTuneOkFrame struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

type ConnectionOpenFrame struct {
	VirtualHost  string
	Capabilities string
	Properties   map[string]interface{}
}

func parseConnectionMethod(config *ClientConfig, methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(constants.CONNECTION_START):
		fmt.Printf("Received CONNECTION_START frame \n")
		startResponse, err := parseConnectionStartFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse connection.start frame: %v", err)
		}

		startOkRequest := createConnectionStartOkRequest(config, startResponse)
		startOkFrame := createConnectionStartOkFrame(&startOkRequest)
		return startOkFrame, nil

	case uint16(constants.CONNECTION_START_OK):
		fmt.Printf("Received CONNECTION_START_OK frame \n")
		return parseConnectionStartOkFrame(payload)

	case uint16(constants.CONNECTION_TUNE):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		tuneResponse, err := parseConnectionTuneFrame(payload)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse connection.tune frame: %v", err)
		}
		tuneOkRequest := createConnectionTuneOkFrame(fineTune(tuneResponse))
		return tuneOkRequest, nil

	case uint16(constants.CONNECTION_TUNE_OK):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		return parseConnectionTuneOkFrame(payload)

	case uint16(constants.CONNECTION_OPEN):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenFrame(payload)

	case uint16(constants.CONNECTION_OPEN_OK):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenOkFrame(payload)

	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
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

func parseConnectionTuneFrame(payload []byte) (*ConnectionTuneOkFrame, error) {
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
	return &ConnectionTuneOkFrame{
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
	return &ConnectionTuneOkFrame{
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
	serverPropertiesStr, err := DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	serverProperties, err := DecodeTable([]byte(serverPropertiesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to decode client properties: %v", err)
	}
	// fmt.Printf("Server Properties: %+v\n", serverProperties)

	/***Mechanisms*/
	mechanismsStr, err := DecodeLongStr(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mechanism: %v", err)
	}
	// fmt.Printf("Mechanisms Srt: %s\n", mechanismsStr)

	/***Locales*/
	localesStr, err := DecodeLongStr(buf)
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

	return &ConnectionStartOkFrame{
		ClientProperties: clientProperties,
		Mechanism:        mechanism,
		Response:         security,
		Locale:           locale,
	}, nil
}

func createConnectionStartOkRequest(config *ClientConfig, startFrameResponse *ConnectionStartFrame) ConnectionStartOkFrame {
	capabilities := map[string]interface{}{
		"basic.nack":             true,
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"publisher_confirms":     true,
	}
	clientProperties := map[string]interface{}{
		"capabilities": capabilities,
		"product":      "Ottermq Client",
		"version":      "0.6.0-alpha",
		"platform":     "golang",
	}
	fmt.Printf("clientProperties: %+v\n", clientProperties)
	mecanismsStr := startFrameResponse.Mechanisms
	localesStr := startFrameResponse.Locales
	mechanismsList := strings.Split(mecanismsStr, ",")
	fmt.Printf("mechanismsList: %+v\n", mechanismsList)
	securityStr := fmt.Sprintf(" %s %s", config.Username, config.Password)
	localesList := strings.Split(localesStr, ",")
	// find 'PLAIN' in mechanismsList
	mechanism := "PLAIN"
	for _, m := range mechanismsList {
		if m == "PLAIN" {
			mechanism = m
			break
		}
	}
	fmt.Printf("mechanism: %s\n", mechanism)
	// find 'en_US' in localesList
	locale := "en_US"
	for _, l := range localesList {
		if l == "en_US" {
			locale = l
			break
		}
	}
	startOk := ConnectionStartOkFrame{
		ClientProperties: clientProperties,
		Mechanism:        mechanism,
		Response:         securityStr,
		Locale:           locale,
	}
	return startOk
}

func createConnectionStartOkFrame(startOk *ConnectionStartOkFrame) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_START_OK

	clientProperties := startOk.ClientProperties
	encodedProperties := EncodeTable(clientProperties)
	payloadBuf.Write(EncodeLongStr(encodedProperties))

	payloadBuf.Write(EncodeShortStr(startOk.Mechanism))
	payloadBuf.Write(EncodeSecurityPlain(startOk.Response))

	payloadBuf.Write(EncodeShortStr(startOk.Locale))

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())

	return frame
}

func FormatMethodFrame(channelNum uint16, class constants.TypeClass, method constants.TypeMethod, methodPayload []byte) []byte {
	var payloadBuf bytes.Buffer

	binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	// payloadBuf.WriteByte(binary.BigEndian.AppendUint16()[])

	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(constants.TYPE_METHOD) // METHOD frame type
	headerBuf := FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame
}

func createConnectionTuneOkFrame(tune *ConnectionTuneOkFrame) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_TUNE_OK

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func CreateConnectionOpenFrame(vhost string) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_OPEN

	// Vhost - short string
	payloadBuf.Write(EncodeShortStr(vhost))

	// Reserved-1 (bit) - set to 0
	payloadBuf.WriteByte(0)

	// Reserved-2 (shortstr) - typically empty
	payloadBuf.Write(EncodeShortStr(""))

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func CreateConnectionOpenOkFrame() []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_OPEN_OK

	// Reserved-1 (bit) - set to 0
	payloadBuf.WriteByte(0)

	frame := FormatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func fineTune(tune *ConnectionTuneOkFrame) *ConnectionTuneOkFrame {
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
