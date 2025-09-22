package shared

import (
	"bytes"
	"fmt"
	"log"
	"net"
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
func ServerHandshake(configurations *map[string]interface{}, conn net.Conn) error {
	// read the protocol header from the client
	clientHeader, err := ReadProtocolHeader(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\n[DEBUG] received: %x\n", clientHeader)

	expectedHeader := []byte(amqp.AMQP_PROTOCOL_HEADER)
	if !bytes.Equal(clientHeader, expectedHeader) {
		err := SendProtocolHeader(conn)
		if err != nil {
			return err
		}
		return fmt.Errorf("bad protocol: %x (%s -> %s)\n", clientHeader, conn.RemoteAddr().String(), conn.LocalAddr().String())
	}
	log.Printf("accepting AMQP connection (%s -> %s)\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	/** connection.start **/
	// send connection.start frame
	startFrame := CreateConnectionStartFrame()
	if err := SendFrame(conn, startFrame); err != nil {
		return err
	}

	/** R connection.start-ok -> W connection.tune **/
	// read connecion.start-ok frame
	frame, err := ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\n[DEBUG] received: %x\n", frame)
	response, err := ParseFrame(configurations, frame)
	if err != nil {
		return err
	}
	state, ok := response.(*amqp.ChannelState)
	if !ok {
		err = fmt.Errorf("Type assertion ChannelState failed")
		return err
	}
	if state.MethodFrame == nil {
		err = fmt.Errorf("MethodFrame is empty")
		return err
	}
	startOkFrame := state.MethodFrame.Content.(*ConnectionStartOkFrame)
	if startOkFrame == nil {
		err = fmt.Errorf("Type assertion ConnectionStartOkFrame failed")
	}

	err = processStartOkContent(configurations, startOkFrame)
	if err != nil {
		return err
	}

	tune := &ConnectionTuneFrame{
		ChannelMax: 2047,
		FrameMax:   131072,
		Heartbeat:  10,
	}
	// create tune frame
	tuneFrame := CreateConnectionTuneFrame(tune)
	if err := SendFrame(conn, tuneFrame); err != nil {
		return err
	}

	// read connection.tune-ok frame
	frame, err = ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", frame)
	response, err = ParseFrame(configurations, frame)
	if err != nil {
		return err
	}
	state, ok = response.(*amqp.ChannelState)
	if !ok {
		err = fmt.Errorf("Type assertion ChannelState failed")
		return err
	}
	if state.MethodFrame == nil {
		err = fmt.Errorf("MethodFrame is empty")
		return err
	}
	tuneOkFrame := state.MethodFrame.Content.(*ConnectionTuneFrame)
	if tuneOkFrame == nil {
		err = fmt.Errorf("Type assertion ConnectionTuneOkFrame failed")
	}
	fmt.Printf("Received connection.tune-ok: %+v\n", tuneOkFrame)
	// TODO: save tuneOK data

	// read connection.open frame
	frame, err = ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", frame)
	// set vhost on configurations

	response, err = ParseFrame(configurations, frame)
	if err != nil {
		return err
	}
	state, ok = response.(*amqp.ChannelState)
	if !ok {
		err = fmt.Errorf("Type assertion ConnectionOpen failed")
		return err
	}
	if state.MethodFrame == nil {
		err = fmt.Errorf("MethodFrame is empty")
		return err
	}

	openFrame, ok := state.MethodFrame.Content.(*ConnectionOpenFrame)
	(*configurations)["vhost"] = openFrame.VirtualHost
	fmt.Printf("Received connection.open: %+v\n", openFrame)

	//send connection.open-ok frame
	openOkFrame := CreateConnectionOpenOkFrame()
	if err := SendFrame(conn, openOkFrame); err != nil {
		return err
	}

	return nil
}

func processStartOkContent(configurations *map[string]interface{}, startOkFrame *ConnectionStartOkFrame) error {
	mechanism := startOkFrame.Mechanism
	if mechanism != "PLAIN" {
		return fmt.Errorf("Mechanism invalid or %s not suported", mechanism)
	}
	// parse username and password from startOkFrame.Response
	fmt.Printf("Response: '%s'\n", startOkFrame.Response)
	credentials := strings.Split(strings.Trim(startOkFrame.Response, " "), " ")
	fmt.Printf("Credentials: username: '%s' password: '%s'\n", credentials[0], credentials[1])
	if len(credentials) != 2 {
		return fmt.Errorf("Failed to parse credentials: invalid format")
	}
	username := credentials[0]
	password := credentials[1]
	// ask for persistdb if user and password match
	authOk, err := persistdb.AuthenticateUser(username, password)
	if err != nil {
		return err
	}
	if !authOk {
		return fmt.Errorf("Authentication failed")
	}
	// set username to configurations
	(*configurations)["username"] = username

	return nil
}
