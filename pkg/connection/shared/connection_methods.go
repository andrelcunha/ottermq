package shared

import (
	"bytes"
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
)

type ConnectionStartOkFrame struct {
	ClientProperties map[string]interface{}
	Mechanism        string
	Security         string
	Locale           string
}

type ConnectionTuneOkFrame struct {
	ChannelMax int16
	FrameMax   int32
	Heartbeat  int16
}

type ConnectionOpenFrame struct {
	VirtualHost  string
	Capabilities string
	Properties   map[string]interface{}
}

func parseConnectionMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(constants.CONNECTION_START_OK):
		fmt.Printf("Received CONNECTION_START_OK frame \n")
		return parseConnectionStartOkFrame(payload)
	case uint16(constants.CONNECTION_TUNE_OK):
		fmt.Printf("Received CONNECTION_TUNE_OK frame \n")
		return parseConnectionTuneOkFrame(payload)
	case uint16(constants.CONNECTION_OPEN):
		fmt.Printf("Received CONNECTION_OPEN frame \n")
		return parseConnectionOpenFrame(payload)

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
		Security:         security,
		Locale:           locale,
	}, nil
}
