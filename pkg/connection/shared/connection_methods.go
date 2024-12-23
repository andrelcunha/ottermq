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

func parseConnectionMethod(methodID uint16, payload []byte) (interface{}, error) {
	switch methodID {
	case uint16(constants.CONNECTION_START_OK):
		fmt.Printf("Received CONNECTION_START_OK frame \n")
		return parseConnectionStartOkFrame(payload)
	default:
		return nil, fmt.Errorf("unknown method ID: %d", methodID)
	}
}

func parseConnectionStartOkFrame(payload []byte) (*ConnectionStartOkFrame, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short")
	}
	fmt.Printf("Payload size :  %d\n", len(payload))

	buf := bytes.NewReader(payload)
	fmt.Printf("Buffer size :  %d\n", buf.Len())

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
