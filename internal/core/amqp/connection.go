package amqp

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp/utils"
)

func sendHeartbeat(conn net.Conn) error {
	heartbeatFrame := createHeartbeatFrame()
	return sendFrame(conn, heartbeatFrame)
}

// createConnectionCloseFrame creates a connection close frame,
func createConnectionCloseFrame(channel, replyCode, classID, methodID uint16, replyText string) []byte {
	replyCodeKv := KeyValue{
		Key:   INT_SHORT,
		Value: replyCode,
	}
	replyTextKv := KeyValue{
		Key:   STRING_SHORT,
		Value: replyText,
	}
	classIDKv := KeyValue{
		Key:   INT_SHORT,
		Value: classID,
	}
	methodIDKv := KeyValue{
		Key:   INT_SHORT,
		Value: methodID,
	}
	content := ContentList{
		KeyValuePairs: []KeyValue{replyCodeKv, replyTextKv, classIDKv, methodIDKv},
	}
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  uint16(CONNECTION),
		MethodID: uint16(CONNECTION_CLOSE),
		Content:  content,
	}.FormatMethodFrame()
	return frame
}

func createConnectionCloseOkFrame(request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
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

func createConnectionTuneFrame(tune *ConnectionTune) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_TUNE

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func createConnectionTuneOkFrame(tune *ConnectionTune) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_TUNE_OK

	binary.Write(&payloadBuf, binary.BigEndian, tune.ChannelMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.FrameMax)
	binary.Write(&payloadBuf, binary.BigEndian, tune.Heartbeat)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func createConnectionOpenOkFrame() []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := CONNECTION
	methodID := CONNECTION_OPEN_OK

	// Reserved-1 (bit) - set to 0
	payloadBuf.WriteByte(0)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func fineTune(tune *ConnectionTune) *ConnectionTune {
	// TODO: get values from config
	tune.ChannelMax = getSmalestShortInt(2047, tune.ChannelMax)
	tune.FrameMax = getSmalestLongInt(131072, tune.FrameMax)
	tune.Heartbeat = getSmalestShortInt(10, tune.Heartbeat)

	return tune
}
