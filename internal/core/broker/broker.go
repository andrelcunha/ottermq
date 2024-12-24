package broker

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andrelcunha/ottermq/config"
	"github.com/andrelcunha/ottermq/internal/core/vhost"
	. "github.com/andrelcunha/ottermq/pkg/common"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/server"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
	_ "github.com/andrelcunha/ottermq/pkg/persistdb"
)

var (
	version = "0.6.0-alpha"
)

const (
	platform = "golang"
	product  = "OtterMQ"
)

type Broker struct {
	VHosts      map[string]*vhost.VHost
	config      *config.Config               `json:"-"`
	Connections map[net.Conn]*ConnectionInfo `json:"-"`
	mu          sync.Mutex                   `json:"-"`
}

func NewBroker(config *config.Config) *Broker {
	b := &Broker{
		VHosts:      make(map[string]*vhost.VHost),
		Connections: make(map[net.Conn]*ConnectionInfo),
		config:      config,
	}
	b.VHosts["/"] = vhost.NewVhost("/")
	return b
}

func (b *Broker) Start() {
	capabilities := map[string]interface{}{
		"basic.nack":             true,
		"connection.blocked":     true,
		"consumer_cancel_notify": true,
		"publisher_confirms":     true,
	}

	serverProperties := map[string]interface{}{
		"capabilities": capabilities,
		"product":      product,
		"version":      version,
		"platform":     platform,
	}
	configurations := map[string]interface{}{
		"mechanisms":        []string{"PLAIN"},
		"locales":           []string{"en_US"},
		"serverProperties":  serverProperties,
		"heartbeatInterval": b.config.HeartbeatIntervalMax,
		"frameMax":          b.config.FrameMax,
		"channelMax":        b.config.ChannelMax,
	}

	addr := fmt.Sprintf("%s:%s", b.config.Host, b.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start vhost: %v", err)
	}
	defer listener.Close()
	log.Printf("Started TCP listener on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		log.Println("New client waiting for connection: ", conn.RemoteAddr())
		go b.handleConnection((&configurations), conn)
	}
}

func (b *Broker) handleConnection(configurations *map[string]interface{}, conn net.Conn) {
	defer func() {
		conn.Close()
		b.cleanupConnection(conn)
	}()
	channelNum := uint16(0)
	if err := server.ServerHandshake(configurations, conn); err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}
	username := (*configurations)["username"].(string)
	vhost := (*configurations)["vhost"].(string)
	heartbeatInterval := (*configurations)["heartbeatInterval"].(uint16)
	b.registerConnection(conn, username, vhost, heartbeatInterval)

	go b.sendHeartbeat(conn)
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

		fmt.Printf("[DEBUG]received: %x\n", frame)

		//Process frame
		_, err = b.ParseFrame(configurations, conn, channelNum, frame)
		if err != nil {
			log.Fatalf("ERROR parsing frame: %v", err)
		}
	}
}

func (b *Broker) registerConnection(conn net.Conn, username, vhost string, heartbeatInterval uint16) {
	b.mu.Lock()
	b.Connections[conn] = &ConnectionInfo{
		User:              username,
		VHost:             vhost,
		HeartbeatInterval: heartbeatInterval,
		ConnectedAt:       time.Now(),
		LastHeartbeat:     time.Now(),
		Conn:              conn,
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

func (b *Broker) ParseFrame(configurations *map[string]interface{}, conn net.Conn, currentChannel uint16, frame []byte) (interface{}, error) {
	if len(frame) < 7 {
		return nil, fmt.Errorf("frame too short")
	}

	frameType := frame[0]
	channel := binary.BigEndian.Uint16(frame[1:3])
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if len(frame) < int(7+payloadSize) {
		return nil, fmt.Errorf("frame too short")
	}
	if channel != currentChannel {
		return nil, fmt.Errorf("unexpected channel: %d", channel)
	}
	payload := frame[7:]

	switch frameType {
	case byte(constants.TYPE_METHOD):
		fmt.Printf("Received METHOD frame on channel %d\n", channel)
		return shared.ParseMethodFrame(configurations, channel, payload)

	case byte(constants.TYPE_HEARTBEAT):
		log.Printf("Received HEARTBEAT frame on channel %d\n", channel)
		err := b.handleHeartbeat(conn, frame)
		return nil, err

	default:
		fmt.Printf("Received: %x\n", frame)
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func (b *Broker) handleHeartbeat(conn net.Conn, frame []byte) error {
	if err := shared.SendFrame(conn, frame); err != nil {
		log.Printf("Error sending heartbeat response: %v", err)
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Connections[conn].LastHeartbeat = time.Now()
	return nil
}

func (b *Broker) sendHeartbeat(conn net.Conn) {
	b.mu.Lock()
	connectionInfo, ok := b.Connections[conn]
	if !ok {
		b.mu.Unlock()
		return
	}
	heartbeatInterval := connectionInfo.HeartbeatInterval
	done := connectionInfo.Done
	b.mu.Unlock()

	ticker := time.NewTicker(time.Duration(heartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			if _, ok := b.Connections[conn]; !ok {
				b.mu.Unlock()
				return
			}
			b.mu.Unlock()

			heartbeatFrame := shared.CreateHeartbeatFrame()
			err := shared.SendFrame(conn, heartbeatFrame)
			if err != nil {
				log.Println("Fail sending heartbeat: ", err)
				return
			}

		case <-done:
			log.Println("Stopping heartbeat  goroutine for closed connection")
			return
		}

	}
}
