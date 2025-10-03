package broker

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
	"github.com/rs/zerolog/log"
	// _ "github.com/andrelcunha/ottermq/internal/core/persistdb"
)

type ConnManager interface {
	HandleConnection(configurations *map[string]any, conn net.Conn) error
	ProcessRequest(conn net.Conn, newState *amqp.ChannelState) (any, error)
	GetChannelState(conn net.Conn, channel uint16) *amqp.ChannelState
	UpdateChannelState(conn net.Conn, channel uint16, newState *amqp.ChannelState)
	HandleHeartbeat(conn net.Conn) error
	GetVHostFromName(vhostName string) *vhost.VHost
}

// DefaultConnManager implements ConnManager
type DefaultConnManager struct {
	broker *Broker
	framer amqp.Framer
}

// NewDefaultConnManager creates a new connection manager
func NewDefaultConnManager(broker *Broker, framer amqp.Framer) *DefaultConnManager {
	return &DefaultConnManager{
		broker: broker,
		framer: framer,
	}
}

func (b *Broker) handleConnection(conn net.Conn, connInfo *amqp.ConnectionInfo) {
	b.ActiveConns.Add(1)
	client := connInfo.Client
	// ctx := client.Ctx

	defer func() {
		defer b.ActiveConns.Done()
		if len(b.Connections) == 0 {
			log.Debug().Msg("No connections to clean")
			return
		}
		log.Debug().Msg("Cleaning connection")
		b.cleanupConnection(conn)
	}()

	channelNum := uint16(0)

	for {
		frame, err := b.framer.ReadFrame(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Debug().Err(err).Msg("Connection timeout")
			}
			if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
				log.Debug().Str("client", conn.RemoteAddr().String()).Msg("Connection closed by client")
			} else {
				log.Error().Err(err).Msg("Error reading frame")
			}
			client.Cancel()
			return
		}

		if len(frame) > 0 { // any octet shall be valid as heartbeat #AMQP_compliance
			b.registerHeartbeat(conn)
		}

		//Process frame
		newInterface, err := b.framer.ParseFrame(frame)
		if err != nil {
			log.Error().Err(err).Msg("Failed parsing frame")
			client.Cancel()
			return
		}
		if _, ok := newInterface.(*amqp.Heartbeat); ok {
			continue
		}

		newState, ok := newInterface.(*amqp.ChannelState)
		if !ok {
			log.Error().Msg("Failed to cast request to ChannelState")
			client.Cancel()
			return
		}

		log.Trace().Interface("state", newState).Msg("New State")

		if newState.MethodFrame != nil {
			request := newState.MethodFrame
			if channelNum != request.Channel {
				channelNum = newState.MethodFrame.Channel
				log.Debug().Uint16("channel", request.Channel).Msg("Switching to channel")
			}
		} else {
			if newState.HeaderFrame != nil {
				log.Debug().Interface("header", newState.HeaderFrame).Msg("HeaderFrame")
			} else if newState.Body != nil {
				log.Debug().Interface("body", newState.Body).Msg("Body")
			}
			if previousState, exists := b.Connections[conn].Channels[channelNum]; exists {
				newState.MethodFrame = previousState.MethodFrame
				log.Debug().Interface("method_frame", previousState.MethodFrame).Msg("Recovered method frame")
			} else {
				log.Debug().Uint16("channel", channelNum).Msg("Channel not found")
				continue
			}
		}
		b.processRequest(conn, newState)
	}
}

func (b *Broker) registerConnection(conn net.Conn, connInfo *amqp.ConnectionInfo) {
	b.mu.Lock()
	b.Connections[conn] = connInfo
	b.mu.Unlock()
}

func (b *Broker) cleanupConnection(conn net.Conn) {
	b.mu.Lock()
	connInfo, ok := b.Connections[conn]
	if !ok {
		b.mu.Unlock()
		return
	}
	vhName := connInfo.VHostName
	delete(b.Connections, conn)
	b.mu.Unlock()
	
	connInfo.Client.Ctx.Done()
	vh := b.GetVHost(vhName)
	if vh != nil {
		vh.CleanupConnection(conn)
	}
}

// closeConnectionRequested closes a connection and sends a CONNECTION_CLOSE_OK frame
func (b *Broker) closeConnectionRequested(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	frame := b.framer.CreateConnectionCloseOkFrame(request)
	err := b.framer.SendFrame(conn, frame)
	b.cleanupConnection(conn)
	return nil, err
}

// closeConnection sends `connection.close` when the server needs to shutdown for some reason
func (b *Broker) sendCloseConnection(conn net.Conn, channel, replyCode, methodId, classId uint16, replyText string) (any, error) {
	frame := b.framer.CreateConnectionCloseFrame(channel, replyCode, methodId, classId, replyText)
	err := b.framer.SendFrame(conn, frame)

	return nil, err
}

func (b *Broker) connectionCloseOk(conn net.Conn) {
	b.cleanupConnection(conn)
	conn.Close()
}

// registerHeartbeat registers a heartbeat for a connection
func (b *Broker) registerHeartbeat(conn net.Conn) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if connInfo, exists := b.Connections[conn]; exists {
		connInfo.Client.LastHeartbeat = time.Now()
		return nil
	}
	return fmt.Errorf("connection not found")
}

func (b *Broker) BroadcastConnectionClose() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for conn := range b.Connections {
		b.sendCloseConnection(conn, 0, 320, 0, 0, "Server shutting down")
	}
}

func (b *Broker) connectionHandler(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	switch request.MethodID {
	case uint16(amqp.CONNECTION_CLOSE):
		return b.closeConnectionRequested(request, conn)
	case uint16(amqp.CONNECTION_CLOSE_OK):
		b.connectionCloseOk(conn)
		return nil, nil
	default:
		log.Debug().Uint16("method_id", request.MethodID).Msg("Unknown connection method")
		return nil, fmt.Errorf("unknown connection method: %d", request.MethodID)
	}
}
