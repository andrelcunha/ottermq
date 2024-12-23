package connection

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

type ServerConnectionHandler struct {
	conn net.Conn
}

func NewServerConnectionHandler(conn net.Conn) *ServerConnectionHandler {
	return &ServerConnectionHandler{conn: conn}
}

func (ph *ServerConnectionHandler) SendFrame(frame []byte) error {
	_, err := ph.conn.Write(frame)
	return err
}

func (ph *ServerConnectionHandler) ConnectionHandshake() error {
	// read the protocol header from the client
	clientHeader, err := shared.ReadProtocolHeader(ph.conn)
	if err != nil {
		return err
	}

	expectedHeader := []byte(constants.AMQP_PROTOCOL_HEADER)
	if !bytes.Equal(clientHeader, expectedHeader) {
		err := shared.SendProtocolHeader(ph.conn)
		if err != nil {
			return err
		}

		return fmt.Errorf("bad protocol: %x (%s -> %s)\n", clientHeader, ph.conn.RemoteAddr().String(), ph.conn.LocalAddr().String())
	}
	log.Printf("accepting AMQP connection (%s -> %s)\n", ph.conn.RemoteAddr().String(), ph.conn.LocalAddr().String())

	// send connection.start frame
	startFrame := createConnectionStartFrame()
	if err := ph.SendFrame(startFrame); err != nil {
		return err
	}

	// read connecion.start-ok frame
	frame, err := shared.ReadFrame(ph.conn)
	if err != nil {
		return err
	}
	shared.ParseFrame(frame)
	// fmt.Printf("Received connection.start-ok: %x\n", frame)

	// send connection.tune frame
	tuneFrame := createConnectionTuneFrame()
	if err := ph.SendFrame(tuneFrame); err != nil {
		return err
	}

	// read connection.tune-ok frame
	frame, err = shared.ReadFrame(ph.conn)
	if err != nil {
		return err
	}
	fmt.Printf("Received connection.tune-ok: %x\n", frame)

	//send connection.open frame
	frame = createConnectionOpenFrame()
	if err := ph.SendFrame(frame); err != nil {
		return err
	}

	// read connection.open-ok frame
	frame, err = shared.ReadFrame(ph.conn)
	if err != nil {
		return err
	}
	fmt.Printf("Received connection.open-ok: %x\n", frame)

	return nil
}

func createConnectionTuneFrame() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{1, 0, 0, 0, 0, 0, 4, 0, 10, 0, 30})
	buf.WriteByte(0xCE)
	return buf.Bytes()
}

func createConnectionOpenFrame() []byte {
	var buf bytes.Buffer
	buf.Write([]byte{1, 0, 0, 0, 0, 0, 4, 0, 10, 0, 40})
	buf.WriteByte(0xCE)
	return buf.Bytes()
}

func createConnectionStartFrame() []byte {
	// class := constants.CONNECTION
	// method := constants.CONNECTION_START

	// var payloadBuf bytes.Buffer

	// serverProperties := map[string]interface{}{
	// 	"product": "OtterMQ",
	// }

	// binary.Write(&payloadBuf, binary.BigEndian, uint16(class))
	// binary.Write(&payloadBuf, binary.BigEndian, uint16(method))

	// payloadBuf.WriteByte(0) // version-major
	// payloadBuf.WriteByte(9) // version-minor

	// propertiesBuf := shared.EncodeTable(serverProperties)
	// payloadBuf.Write(propertiesBuf)

	// // Payload: Mechanisms
	// payloadBuf.Write([]byte{'P', 'L', 'A', 'I', 'N', 0})

	// // Payload: Locales
	// payloadBuf.Write([]byte("en_US"))

	// // Calculate the size of the payload
	// payloadSize := uint32(payloadBuf.Len())

	// // Buffer for the frame header
	// frameType := uint8(constants.TYPE_METHOD) // METHOD frame type
	// channelNum := uint16(0)                   // Channel number
	// headerBuf := shared.FormatHeader(frameType, channelNum, payloadSize)

	// frame := append(headerBuf, payloadBuf.Bytes()...)

	// frame = append(frame, 0xCE) // frame-end

	// return frame
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_START

	payloadBuf.WriteByte(0) // version-major
	payloadBuf.WriteByte(9) // version-minor

	serverProperties := map[string]interface{}{
		"product": "OtterMQ",
	}
	encodedProperties := shared.EncodeTable(serverProperties)
	payloadBuf.Write(shared.EncodeLongStr(encodedProperties))

	payloadBuf.Write(shared.EncodeLongStr([]byte("PLAIN")))

	payloadBuf.Write(shared.EncodeLongStr([]byte("en_US")))

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
