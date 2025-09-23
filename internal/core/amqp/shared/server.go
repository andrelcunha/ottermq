package shared

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/persistdb"
)

// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func Handshake(configurations *map[string]any, conn net.Conn) (*AmqpClient, error) {
	// read the protocol header from the client
	clientHeader, err := ReadProtocolHeader(conn)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Handshake - Received: %x", clientHeader)
	protocol := (*configurations)["protocol"].(string)
	protocolHeader, err := BuildProtocolHeader(protocol)
	if err != nil {
		msg := "error parsing protocol"
		log.Printf("[ERROR] %s", msg)
		return nil, fmt.Errorf(msg)
	}
	if !bytes.Equal(clientHeader, protocolHeader) {
		err := SendProtocolHeader(conn, protocolHeader)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("bad protocol: %x (%s -> %s)", clientHeader, conn.RemoteAddr().String(), conn.LocalAddr().String())
	}
	log.Printf("[DEBUG] Handshake - Accepting AMQP connection (%s -> %s)\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	/** connection.start **/
	// send connection.start frame
	startFrame := CreateConnectionStartFrame()
	if err := SendFrame(conn, startFrame); err != nil {
		return nil, err
	}

	/** R connection.start-ok -> W connection.tune **/
	// read connecion.start-ok frame
	frame, err := ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	log.Printf("\n[DEBUG] - Handshake - Received: %x\n", frame)
	response, err := ParseFrame(configurations, conn, 0, frame)
	if err != nil {
		return nil, err
	}
	state, ok := response.(*amqp.ChannelState)
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
	tuneFrame := CreateConnectionTuneFrame(tune)
	if err := SendFrame(conn, tuneFrame); err != nil {
		return nil, err
	}

	// read connection.tune-ok frame
	frame, err = ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	response, err = ParseFrame(configurations, conn, 0, frame)
	if err != nil {
		return nil, err
	}
	state, ok = response.(*amqp.ChannelState)
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
	frame, err = ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	// set vhost on configurations

	response, err = ParseFrame(configurations, conn, 0, frame)
	if err != nil {
		return nil, err
	}
	state, ok = response.(*amqp.ChannelState)
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
	openOkFrame := CreateConnectionOpenOkFrame()
	if err := SendFrame(conn, openOkFrame); err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Handshake - connection (%s -> %s) Opened\n", conn.RemoteAddr().String(), conn.LocalAddr().String())
	log.Printf("[DEBUG] Handshake - connection (%s -> %s) Opened\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	return client, nil
}

func processStartOkContent(configurations *map[string]interface{}, startOkFrame *ConnectionStartOkFrame) error {
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

func BuildProtocolHeader(protocol string) ([]byte, error) {
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
