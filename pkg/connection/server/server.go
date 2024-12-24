package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
	"github.com/andrelcunha/ottermq/pkg/persistdb"
)

// Client sends ProtocolHeader
// Server responds with connection.start
// Client responds with connection.start-ok
// Server responds with connection.tune
// Client responds with connection.tune-ok
// Client sends connection.open
// Server responds with connection.open-ok
func ServerHandshake(configurations map[string]interface{}, conn net.Conn) error {
	// read the protocol header from the client
	clientHeader, err := shared.ReadProtocolHeader(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\n[DEBUG] received: %x\n", clientHeader)

	expectedHeader := []byte(constants.AMQP_PROTOCOL_HEADER)
	if !bytes.Equal(clientHeader, expectedHeader) {
		err := shared.SendProtocolHeader(conn)
		if err != nil {
			return err
		}
		return fmt.Errorf("bad protocol: %x (%s -> %s)\n", clientHeader, conn.RemoteAddr().String(), conn.LocalAddr().String())
	}
	log.Printf("accepting AMQP connection (%s -> %s)\n", conn.RemoteAddr().String(), conn.LocalAddr().String())

	/** connection.start **/
	// send connection.start frame
	startFrame := shared.CreateConnectionStartFrame()
	if err := shared.SendFrame(conn, startFrame); err != nil {
		return err
	}

	/** R connection.start-ok -> W connection.tune **/
	// read connecion.start-ok frame
	frame, err := shared.ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\n[DEBUG] received: %x\n", frame)
	response, err := shared.ParseFrame(configurations, frame)
	if err != nil {
		return err
	}
	startOkFrame, ok := response.(*shared.ConnectionStartOkFrame)
	if !ok {
		err = fmt.Errorf("Type assertion ConnectionStartOkFrame failed")
		return err
	}
	err = processStartOkContent(startOkFrame)
	if err != nil {
		return err
	}

	tune := &shared.ConnectionTuneFrame{
		ChannelMax: 2047,
		FrameMax:   131072,
		Heartbeat:  10,
	}
	// create tune frame
	tuneFrame := shared.CreateConnectionTuneFrame(tune)
	if err := shared.SendFrame(conn, tuneFrame); err != nil {
		return err
	}

	// read connection.tune-ok frame
	frame, err = shared.ReadFrame(conn)
	if err != nil {
		return err
	}
	fmt.Printf("\nreceived: %x\n", frame)
	tuneOk, err := shared.ParseFrame(configurations, frame)
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
	open, err := shared.ParseFrame(configurations, frame)
	fmt.Printf("Received connection.open: %+v\n", open)

	//send connection.open-ok frame
	openOkFrame := shared.CreateConnectionOpenOkFrame()
	if err := shared.SendFrame(conn, openOkFrame); err != nil {
		return err
	}

	return nil
}

func processStartOkContent(startOkFrame *shared.ConnectionStartOkFrame) error {
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

	return nil
}
