package broker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

func (b *Broker) connectionHandshake(conn net.Conn) error {
	// read the protocol header from the client
	clientHeader, err := shared.ReadProtocolHeader(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", clientHeader)

	expectedHeader := []byte(constants.AMQP_PROTOCOL_HEADER)
	if !bytes.Equal(clientHeader, expectedHeader) {
		err := shared.SendProtocolHeader(conn)
		if err != nil {
			return err
		}

		return fmt.Errorf("bad protocol: %x (%s -> %s)\n", clientHeader, conn.RemoteAddr().String(), conn.LocalAddr().String())
	}
	log.Printf("accepting AMQP connection (%s -> %s)\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	// send connection.start frame
	startFrame := createConnectionStartFrame()
	if err := b.sendFrame(conn, startFrame); err != nil {
		return err
	}

	// read connecion.start-ok frame
	frame, err := shared.ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", frame)
	startOk, err := b.ParseFrame(conn, frame)
	fmt.Printf("Received connection.start-ok: %+v\n", startOk)

	// send connection.tune frame
	tuneFrame := createConnectionTuneFrame(0, 0, 0)
	if err := b.sendFrame(conn, tuneFrame); err != nil {
		return err
	}

	// read connection.tune-ok frame
	frame, err = shared.ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", frame)
	tuneOk, err := b.ParseFrame(conn, frame)
	if err != nil {
		return err
	}
	fmt.Printf("Received connection.tune-ok: %+v\n", tuneOk)

	// read connection.open frame
	frame, err = shared.ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", frame)
	open, err := b.ParseFrame(conn, frame)
	fmt.Printf("Received connection.open: %+v\n", open)

	//send connection.open-ok frame
	openOkFrame := createConnectionOpenOkFrame()
	if err := b.sendFrame(conn, openOkFrame); err != nil {
		return err
	}

	return nil
}

func (b *Broker) sendFrame(conn net.Conn, frame []byte) error {
	_, err := conn.Write(frame)
	return err
}

func createConnectionStartFrame() []byte {
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

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())

	return frame
}

func createConnectionTuneFrame(channelMax uint16, frameMax uint32, heartbeat uint16) []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_TUNE

	binary.Write(&payloadBuf, binary.BigEndian, channelMax)
	binary.Write(&payloadBuf, binary.BigEndian, frameMax)
	binary.Write(&payloadBuf, binary.BigEndian, heartbeat)

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

func createConnectionOpenOkFrame() []byte {
	var payloadBuf bytes.Buffer
	channelNum := uint16(0)
	classID := constants.CONNECTION
	methodID := constants.CONNECTION_OPEN_OK

	frame := formatMethodFrame(channelNum, classID, methodID, payloadBuf.Bytes())
	return frame
}

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
