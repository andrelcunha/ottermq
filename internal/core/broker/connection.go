package broker

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
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

func (b *Broker) handleConnection(configurations *map[string]any, conn net.Conn) {
	defer func() {
		defer b.ActiveConns.Done()
		b.cleanupConnection(conn)
		// log.Fatalf("Connection closed by server")
	}()
	channelNum := uint16(0) //initial
	connInfo, err := b.framer.Handshake(configurations, conn)
	if err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	b.registerConnection(conn, connInfo)
	// TODO: create a goroutine to monitor heartbeat timeout
	go b.sendHeartbeats(conn, connInfo.Client)
	go b.monitorHeartbeatTimeout(conn, connInfo.Client)
	// keep reading commands in loop
	for {
		frame, err := b.framer.ReadFrame(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("[DEBUG] Connection timeout: %v", err)
			}
			if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
				b.cleanupConnection(conn)
				log.Printf("[DEBUG] Connection closed by client: %v", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading frame: %v", err)
			return
		}
		if len(frame) > 0 { // any octet shall be valid as heartbeat #AMQP_compliance

			b.handleHeartbeat(conn)
		}

		//Process frame
		newInterface, err := b.framer.ParseFrame(frame)
		if err != nil {
			log.Fatalf("ERROR parsing frame: %v", err)
		}
		if _, ok := newInterface.(*amqp.Heartbeat); ok {
			continue
		}
		if newInterface != nil {
			newState, ok := newInterface.(*amqp.ChannelState)
			if !ok {
				log.Fatalf("Failed to cast request to amqp.ChannelState")
			}
			fmt.Printf("[DEBUG] New State: %+v\n", newState)

			if newState.MethodFrame != nil {
				request := newState.MethodFrame
				if channelNum != request.Channel {
					channelNum = newState.MethodFrame.Channel
					fmt.Printf("[DEBUG] Newchannel shall be added: %d\n", request.Channel)
				}
			} else {
				if newState.HeaderFrame != nil {
					log.Printf("[DEBUG] HeaderFrame: %+v\n", newState.HeaderFrame)
				} else if newState.Body != nil {
					log.Printf("[DEBUG] Body: %+v\n", newState.Body)
				}
				newState.MethodFrame = b.Connections[conn].Channels[channelNum].MethodFrame
				fmt.Printf("[DEBUG] Request: %+v\n", newState.MethodFrame)
			}
			b.processRequest(conn, newState)
		}
	}
}

func (b *Broker) monitorHeartbeatTimeout(conn net.Conn, client *amqp.AmqpClient) {
	maxTime := time.Duration(b.config.HeartbeatIntervalMax << 1)
	if client.LastHeartbeat.Add(maxTime * time.Second).Before(time.Now()) {
		b.connectionCloseOk(conn)
	}
}

func (b *Broker) registerConnection(conn net.Conn, connInfo *amqp.ConnectionInfo) {
	b.mu.Lock()
	previousCount := len(b.Connections)
	b.Connections[conn] = connInfo
	delta := len(b.Connections) - previousCount
	b.ActiveConns.Add(delta)
	log.Printf("[DEBUG] Connection delta changed: %d", delta)
	b.mu.Unlock()
}

func (b *Broker) cleanupConnection(conn net.Conn) {
	log.Println("Cleaning connection")
	if connInfo, ok := b.Connections[conn]; ok {
		connInfo.Client.Done <- struct{}{} // should stop heartbeat verification
		vhName := connInfo.VHostName
		vh := b.GetVHost(vhName)
		vh.CleanupConnection(conn)
		b.mu.Lock()
		previousCount := len(b.Connections)
		delete(b.Connections, conn)
		delta := len(b.Connections) - previousCount
		b.ActiveConns.Add(delta)
		log.Printf("[DEBUG] Connection delta changed: %d", delta)
		b.mu.Unlock()
	}
}

// closeConnectionRequested closes a connection and sends a CONNECTION_CLOSE_OK frame
func (b *Broker) closeConnectionRequested(conn net.Conn, channel uint16) (any, error) {
	frame := b.framer.CreateConnectionCloseOkFrame(channel)
	err := b.framer.SendFrame(conn, frame)
	b.cleanupConnection(conn)
	return nil, err
}

// closeConnection sends `connection.close` when the server needs to shutdown for some reason
func (b *Broker) sendCloseConnection(conn net.Conn, channel uint16, replyCode uint16, replyText string, methodId uint16, classId uint16) (any, error) {
	frame := b.framer.CreateConnectionCloseFrame(channel, replyCode, replyText, methodId, classId)
	err := b.framer.SendFrame(conn, frame)

	return nil, err
}

func (b *Broker) connectionCloseOk(conn net.Conn) {
	b.cleanupConnection(conn)
	conn.Close()
	// TODO: Verify if heartbeater already stopped for this conn
}

// openChannel executes the AMQP command CHANNEL_OPEN
func (b *Broker) openChannel(request *amqp.RequestMethodMessage, conn net.Conn, channel uint16) (any, error) {
	fmt.Printf("[DEBUG] Received channel open request: %+v\n", request)

	// Check if the channel is already open
	if b.checkChannel(conn, channel) {
		fmt.Printf("[DEBUG] Channel %d already open\n", channel)
		return nil, fmt.Errorf("channel already open")
	}
	b.registerChannel(conn, request)
	fmt.Printf("[DEBUG] New state added: %+v\n", b.Connections[conn].Channels[request.Channel])

	frame := b.framer.CreateChannelOpenOkFrame(channel, request)

	b.framer.SendFrame(conn, frame)
	return nil, nil
}

// checkChannel checks if a channel is already open
func (b *Broker) checkChannel(conn net.Conn, channel uint16) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.Connections[conn].Channels[channel]
	return ok
}

func (b *Broker) closeChannel(conn net.Conn, channel uint16) (any, error) {
	if b.checkChannel(conn, channel) {
		fmt.Printf("[DEBUG] Channel %d already open\n", channel)
		return nil, fmt.Errorf("channel already open")
	}
	b.removeChannel(conn, channel)
	frame := b.framer.CreateChannelCloseFrame(channel)
	b.framer.SendFrame(conn, frame)
	return nil, nil
}

// registerChannel register a new channel to the connection
func (b *Broker) registerChannel(conn net.Conn, frame *amqp.RequestMethodMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Connections[conn].Channels[frame.Channel] = &amqp.ChannelState{MethodFrame: frame}
	fmt.Printf("[DEBUG] New channel added: %d\n", frame.Channel)
}

// removeChannel removes a channel from the connection
func (b *Broker) removeChannel(conn net.Conn, channel uint16) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Connections[conn].Channels, channel)
}

func (b *Broker) sendHeartbeats(conn net.Conn, client *amqp.AmqpClient) {
	b.mu.Lock()
	heartbeatInterval := int(client.HeartbeatInterval >> 1)
	done := client.Done
	b.mu.Unlock()

	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			shuttingDown := b.ShuttingDown.Load()
			_, exists := b.Connections[conn]
			b.mu.Unlock()
			if shuttingDown {
				return
			}
			if !exists {
				log.Println("[TRACE] Connection no longer exists in broker")
				return
			}

			err := b.framer.SendHearbeat(conn)
			if err != nil {
				log.Printf("[DEBUG] Failed to send heartbeat: %v", err)
				return
			}

		case <-done:
			log.Println("[TRACE] Stopping heartbeat goroutine for closed connection")
			return
		}
	}
}

func (b *Broker) handleHeartbeat(conn net.Conn) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Connections[conn].Client.LastHeartbeat = time.Now()

	return nil
}

func (b *Broker) BroadcastConnectionClose() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for conn := range b.Connections {
		b.sendCloseConnection(conn, 0, 320, "Server shutting down", 0, 0)
	}
}
