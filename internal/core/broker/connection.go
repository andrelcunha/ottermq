package broker

import (
	"bytes"
	"encoding/binary"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

func formatMethodFrame(channelNum uint16, class constants.TypeClass, method constants.TypeMethod, methodPayload []byte) []byte {
	var payloadBuf bytes.Buffer

	binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	// payloadBuf.WriteByte(binary.BigEndian.AppendUint16()[])

	payloadBuf.Write(methodPayload)

	// Calculate the size of the payload
	payloadSize := uint32(payloadBuf.Len())

	// Buffer for the frame header
	frameType := uint8(constants.TYPE_METHOD) // METHOD frame type
	headerBuf := shared.FormatHeader(frameType, channelNum, payloadSize)

	frame := append(headerBuf, payloadBuf.Bytes()...)

	frame = append(frame, 0xCE) // frame-end

	return frame
}

func FormatContentFrame(channelNum uint16, class constants.TypeClass, method constants.TypeMethod, methodPayload []byte) []byte {
	panic("Not implemented")
}

func FormatHeartbeatFrame() []byte {
	panic("Not implemented")
}

func FormatArgument(key, value, string, valueType any) []byte {
	panic("not implemented")
}
