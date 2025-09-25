package broker

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
	"github.com/andrelcunha/ottermq/internal/core/models"
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
	mu     sync.Mutex
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
		conn.Close()
		b.cleanupConnection(conn)
	}()
	channelNum := uint16(0) //initial
	client, err := b.framer.Handshake(configurations, conn)
	if err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	b.registerConnection(conn, client)
	go b.sendHeartbeats(conn, client)
	// keep reading commands in loop
	for {
		frame, err := b.framer.ReadFrame(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("[DEBUG] Connection timeout: %v", err)
			}
			if err == io.EOF {
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

		log.Printf("[DEBUG] received: %x\n", frame)

		//Process frame
		newInterface, err := b.framer.ParseFrame(configurations, conn, channelNum, frame)
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

func (b *Broker) registerConnection(conn net.Conn, client *amqp.AmqpClient) {
	vhost := b.GetVHostFromName(client.VHostName)
	if vhost == nil {
		log.Fatalf("VHost not found: %s", client.VHostName)
	}
	client.VHostId = vhost.Id
	b.mu.Lock()

	b.Connections[conn] = &models.ConnectionInfo{
		Client:   client,
		Channels: make(map[uint16]*amqp.ChannelState),
	}
	b.mu.Unlock()
}

func (b *Broker) cleanupConnection(conn net.Conn) {
	log.Println("Cleaning connection")
	b.mu.Lock()
	delete(b.Connections, conn)
	b.mu.Unlock()
	for _, vhost := range b.VHosts {
		vhost.CleanupConnection(conn)
	}
}

func (b *Broker) checkChannel(conn net.Conn, channel uint16) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.Connections[conn].Channels[channel]
	return ok
}

// Add new Channel
func (b *Broker) addChannel(conn net.Conn, frame *amqp.RequestMethodMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// add new channel to the connectionInfo
	b.Connections[conn].Channels[frame.Channel] = &amqp.ChannelState{MethodFrame: frame}
	fmt.Printf("[DEBUG] New channel added: %d\n", frame.Channel)
}

func (b *Broker) removeChannel(conn net.Conn, channel uint16) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Connections[conn].Channels, channel)
}

func (b *Broker) sendHeartbeats(conn net.Conn, client *amqp.AmqpClient) {
	b.mu.Lock()
	// connectionInfo, ok := b.Connections[conn]
	// if !ok {
	// 	b.mu.Unlock()
	// 	return
	// }
	heartbeatInterval := int(client.HeartbeatInterval >> 1)
	done := client.Done
	b.mu.Unlock()

	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			if _, ok := b.Connections[conn]; !ok {
				b.mu.Unlock()
				log.Println("Connection no longer exists in broker")
				return
			}
			b.mu.Unlock()

			// sendHearbeat(conn)
			// heartbeatFrame := amqp.CreateHeartbeatFrame()
			// err := shared.SendFrame(conn, heartbeatFrame)
			err := b.framer.SendHearbeat(conn)
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Println("[DEBUG] Heartbeat sent")

		case <-done:
			log.Println("Stopping heartbeat  goroutine for closed connection")
			return
		}
	}
}

func (b *Broker) handleHeartbeat(conn net.Conn) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	// b.Connections[conn].LastHeartbeat = time.Now()
	b.Connections[conn].Client.LastHeartbeat = time.Now()

	return nil
}
