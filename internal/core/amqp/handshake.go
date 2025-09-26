package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/andrelcunha/ottermq/internal/core/persistdb"
)

// REGION Share
// Probably should be called Handshake or something

// handshake
// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func handshake(configurations *map[string]any, conn net.Conn) (*AmqpClient, error) {
	// read the protocol header from the client
	clientHeader, err := readProtocolHeader(conn)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Handshake - Received: %x", clientHeader)
	protocol := (*configurations)["protocol"].(string)
	protocolHeader, err := buildProtocolHeader(protocol)
	if err != nil {
		msg := "error parsing protocol"
		log.Printf("[ERROR] %s", msg)
		return nil, fmt.Errorf("%s", msg)
	}
	if !bytes.Equal(clientHeader, protocolHeader) {
		err := sendProtocolHeader(conn, protocolHeader)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("bad protocol: %x (%s -> %s)", clientHeader, conn.RemoteAddr().String(), conn.LocalAddr().String())
	}
	log.Printf("[DEBUG] Handshake - Accepting AMQP connection (%s -> %s)\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	/** connection.start **/
	// send connection.start frame
	startFrame := createConnectionStartFrame()
	if err := sendFrame(conn, startFrame); err != nil {
		return nil, err
	}

	/** R connection.start-ok -> W connection.tune **/
	// read connecion.start-ok frame
	frame, err := readFrame(conn)
	if err != nil {
		return nil, err
	}
	log.Printf("\n[DEBUG] - Handshake - Received: %x\n", frame)
	response, err := parseFrame(frame)
	if err != nil {
		return nil, err
	}
	state, ok := response.(*ChannelState)
	if !ok {
		return nil, fmt.Errorf("type assertion ChannelState failed")
	}
	if state.MethodFrame == nil {
		return nil, fmt.Errorf("methodFrame is empty")
	}
	startOkFrame := state.MethodFrame.Content.(*ConnectionStartOkFrame)
	if startOkFrame == nil {
		return nil, fmt.Errorf("type assertion ConnectionStartOkFrame failed")
	}

	err = processStartOkContent(configurations, startOkFrame)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Handshake - connection (%s -> %s) Started\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	heartbeat, _ := (*configurations)["heartbeatInterval"].(uint16)
	frameMax, _ := (*configurations)["frameMax"].(uint32)
	channelMax, _ := (*configurations)["channelMax"].(uint16)

	tune := &ConnectionTuneFrame{
		ChannelMax: uint16(channelMax), //2047,
		FrameMax:   uint32(frameMax),   //131072,
		Heartbeat:  uint16(heartbeat),  //10
	}
	// create tune frame
	tuneFrame := createConnectionTuneFrame(tune)
	if err := sendFrame(conn, tuneFrame); err != nil {
		return nil, err
	}

	// read connection.tune-ok frame
	frame, err = readFrame(conn)
	if err != nil {
		return nil, err
	}
	response, err = parseFrame(frame)
	if err != nil {
		return nil, err
	}
	state, ok = response.(*ChannelState)
	if !ok {
		return nil, fmt.Errorf("type assertion ChannelState failed")
	}
	if state.MethodFrame == nil {
		return nil, fmt.Errorf("MethodFrame is empty")
	}
	tuneOkFrame := state.MethodFrame.Content.(*ConnectionTuneFrame)
	if tuneOkFrame == nil {
		return nil, fmt.Errorf("type assertion ConnectionTuneOkFrame failed")
	}

	(*configurations)["heartbeatInterval"] = tuneOkFrame.Heartbeat
	(*configurations)["frameMax"] = tuneOkFrame.FrameMax
	(*configurations)["channelMax"] = tuneOkFrame.ChannelMax

	config := NewAmqpClientConfig(configurations)
	client := NewAmqpClient(conn, config)
	log.Printf("[DEBUG] Handshake completed")

	// read connection.open frame
	frame, err = readFrame(conn)
	if err != nil {
		return nil, err
	}
	// set vhost on configurations

	response, err = parseFrame(frame)
	if err != nil {
		return nil, err
	}
	state, ok = response.(*ChannelState)
	if !ok {
		return nil, fmt.Errorf("type assertion ConnectionOpen failed")
	}
	if state.MethodFrame == nil {
		return nil, fmt.Errorf("methodFrame is empty")
	}

	openFrame, _ := state.MethodFrame.Content.(*ConnectionOpenFrame)
	(*configurations)["vhost"] = openFrame.VirtualHost
	client.VHostName = openFrame.VirtualHost

	//send connection.open-ok frame
	openOkFrame := createConnectionOpenOkFrame()
	if err := sendFrame(conn, openOkFrame); err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Handshake - connection (%s -> %s) Opened\n", conn.RemoteAddr().String(), conn.LocalAddr().String())
	log.Printf("[DEBUG] Handshake Completed")

	return client, nil
}

func processStartOkContent(configurations *map[string]any, startOkFrame *ConnectionStartOkFrame) error {
	mechanism := startOkFrame.Mechanism
	if mechanism != "PLAIN" {
		return fmt.Errorf("mechanism invalid or %s not suported", mechanism)
	}
	// parse username and password from startOkFrame.Response
	fmt.Printf("Response: '%s'\n", startOkFrame.Response)
	credentials := strings.Split(strings.Trim(startOkFrame.Response, " "), " ")
	fmt.Printf("Credentials: username: '%s' password: '%s'\n", credentials[0], credentials[1])
	if len(credentials) != 2 {
		return fmt.Errorf("failed to parse credentials: invalid format")
	}
	username := credentials[0]
	password := credentials[1]
	// ask for persistdb if user and password match
	authOk, err := persistdb.AuthenticateUser(username, password)
	if err != nil {
		return err
	}
	if !authOk {
		return fmt.Errorf("authentication failed")
	}
	// set username to configurations
	(*configurations)["username"] = username

	return nil
}

func buildProtocolHeader(protocol string) ([]byte, error) {
	parts := strings.Split(protocol, " ")
	if len(parts) != 2 || parts[0] != "AMQP" {
		return nil, fmt.Errorf("invalid protocol format: %s", protocol)
	}

	versionParts := strings.Split(parts[1], "-")
	if len(versionParts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", parts[1])
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %v", err)
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %v", err)
	}
	revision, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid revision version: %v", err)
	}

	header := []byte{
		'A', 'M', 'Q', 'P',
		0x00,
		byte(major),
		byte(minor),
		byte(revision),
	}

	return header, nil
}

func sendProtocolHeader(conn net.Conn, header []byte) error {
	_, err := conn.Write(header)
	return err
}

func readProtocolHeader(conn net.Conn) ([]byte, error) {
	header := make([]byte, 8)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func readFrame(conn net.Conn) ([]byte, error) {
	// all frames starts with a 7-octet header
	frameHeader := make([]byte, 7)
	_, err := io.ReadFull(conn, frameHeader)
	if err != nil {
		return nil, err
	}

	// fist octet is the type of frame
	// frameType := binary.BigEndian.Uint16(frameHeader[0:1])

	// 2nd and 3rd octets (short) are the channel number
	// channelNum := binary.BigEndian.Uint16(frameHeader[1:3])

	// 4th to 7th octets (long) are the size of the payload
	payloadSize := binary.BigEndian.Uint32(frameHeader[3:])

	// read the framePayload
	framePayload := make([]byte, payloadSize)
	_, err = io.ReadFull(conn, framePayload)
	if err != nil {
		return nil, err
	}

	// frame-end is a 1-octet after the payload
	frameEnd := make([]byte, 1)
	_, err = io.ReadFull(conn, frameEnd)
	if err != nil {
		return nil, err
	}

	// check if the frame-end is correct (0xCE)
	if frameEnd[0] != 0xCE {
		// return nil, ErrInvalidFrameEnd
		return nil, fmt.Errorf("invalid frame end octet")
	}

	return append(frameHeader, framePayload...), nil
}

func sendFrame(conn net.Conn, frame []byte) error {
	fmt.Printf("[DEBUG] Sending frame: %x\n", frame)
	_, err := conn.Write(frame)
	return err
}

func createHeartbeatFrame() []byte {
	frame := make([]byte, 8)
	frame[0] = byte(TYPE_HEARTBEAT)
	binary.BigEndian.PutUint16(frame[1:3], 0)
	binary.BigEndian.PutUint32(frame[3:7], 0)
	frame[7] = 0xCE
	return frame
}
