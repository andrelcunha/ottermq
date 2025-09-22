package broker

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/amqp/shared"
	"github.com/andrelcunha/ottermq/internal/core/models"
	_ "github.com/andrelcunha/ottermq/internal/core/persistdb"
)

func (b *Broker) handleConnection(configurations *map[string]any, conn net.Conn) {
	defer func() {
		conn.Close()
		b.cleanupConnection(conn)
	}()
	channelNum := uint16(0)

	if err := shared.ServerHandshake(configurations, conn); err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}
	username := (*configurations)["username"].(string)
	vhost := (*configurations)["vhost"].(string)
	heartbeatInterval := (*configurations)["heartbeatInterval"].(uint16)

	b.registerConnection(conn, username, vhost, heartbeatInterval)
	go b.sendHeartbeats(conn)
	log.Println("Handshake successful")

	// keep reading commands in loop
	for {
		frame, err := shared.ReadFrame(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("Connection timeout: %v", err)
			}
			if err == io.EOF {
				b.cleanupConnection(conn)
				log.Printf("Connection closed by client: %v", conn.RemoteAddr())
				return
			}
			log.Printf("Error reading frame: %v", err)
			return
		}

		log.Printf("[DEBUG] received: %x\n", frame)

		//Process frame
		newInterface, err := b.ParseFrame(configurations, conn, channelNum, frame)
		if err != nil {
			log.Fatalf("ERROR parsing frame: %v", err)
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

func (b *Broker) registerConnection(conn net.Conn, username, vhostName string, heartbeatInterval uint16) {
	vhost := b.GetVHostFromName(vhostName)
	if vhost == nil {
		log.Fatalf("VHost not found: %s", vhostName)
	}

	b.mu.Lock()

	b.Connections[conn] = &models.ConnectionInfo{
		Name:              conn.RemoteAddr().String(),
		User:              username,
		VHostName:         vhost.Name,
		VHostId:           vhost.Id,
		HeartbeatInterval: heartbeatInterval,
		ConnectedAt:       time.Now(),
		LastHeartbeat:     time.Now(),
		Conn:              conn,
		Channels:          make(map[uint16]*amqp.ChannelState),
		Done:              make(chan struct{}),
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
