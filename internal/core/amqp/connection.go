package amqp

import (
	"bytes"
	"log"

	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

func createConnectionCloseFrame(channel uint16) []byte {
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  uint16(CONNECTION),
		MethodID: uint16(CONNECTION_CLOSE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}

func createConnectionStartFrame(configurations *map[string]any) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_START

	payloadBuf.WriteByte(0) // version-major
	payloadBuf.WriteByte(9) // version-minor

	serverPropsRaw, ok := (*configurations)["serverProperties"]
	if !ok {
		log.Fatalf("serverProperties not found in configurations")
	}
	serverProperties, ok := serverPropsRaw.(map[string]any)
	if !ok {
		log.Fatalf("serverProperties is not a map[string]any")
	}
	encodedProperties := utils.EncodeTable(serverProperties)

	payloadBuf.Write(utils.EncodeLongStr(encodedProperties))

	// Extract mechanisms
    mechanismsRaw, ok := (*configurations)["mechanisms"]
    if !ok {
        log.Fatalf("mechanisms not found in configurations")
    }
    mechanismsSlice, ok := mechanismsRaw.([]string)
    if !ok || len(mechanismsSlice) == 0 {
        log.Fatalf("mechanisms is not a non-empty []string")
    }
    payloadBuf.Write(utils.EncodeLongStr([]byte(mechanismsSlice[0])))

	// Extract locales
    localesRaw, ok := (*configurations)["locales"]
    if !ok {
        log.Fatalf("locales not found in configurations")
    }
    localesSlice, ok := localesRaw.([]string)
    if !ok || len(localesSlice) == 0 {
        log.Fatalf("locales is not a non-empty []string")
    }
    payloadBuf.Write(utils.EncodeLongStr([]byte(localesSlice[0])))

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	log.Printf("[DEBUG] Sending CONNECTION_START frame: %v", frame)
	return frame
}
