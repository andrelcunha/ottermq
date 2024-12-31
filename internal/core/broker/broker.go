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
	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/constants/basic"
	"github.com/andrelcunha/ottermq/pkg/connection/constants/exchange"
	"github.com/andrelcunha/ottermq/pkg/connection/constants/queue"
	"github.com/andrelcunha/ottermq/pkg/connection/constants/tx"
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

		fmt.Printf("[DEBUG]received: %x\n", frame)

		//Process frame
		requestInterface, err := b.ParseFrame(configurations, conn, channelNum, frame)
		if err != nil {
			log.Fatalf("ERROR parsing frame: %v", err)
		}
		if requestInterface != nil {
			request, ok := requestInterface.(*amqp.RequestMethodMessage)
			if !ok {
				log.Fatalf("Failed to cast request to amqp.Message")
			}
			if request.ClassID == uint16(constants.CONNECTION) {
				switch request.MethodID {
				case uint16(constants.CONNECTION_CLOSE):
					b.cleanupConnection(conn)
					// send close-ok
					// closeOk := amqp.NewFrame(amqp.MethodMessage{
					// 	Channel:  0,
					// 	ClassID:  uint16(constants.CONNECTION),
					// 	MethodID: uint16(constants.CONNECTION_CLOSE_OK),
					// 	Content:  nil,
					// })
					frame := amqp.ResponseMethodMessage{
						Channel:  0,
						ClassID:  uint16(constants.CONNECTION),
						MethodID: uint16(constants.CONNECTION_CLOSE_OK),
						Content:  amqp.ContentList{},
					}.FormatMethodFrame()
					shared.SendFrame(conn, frame)
				}
				continue
			}
			b.executeRequest(request)
		}
	}
}

// func sendResponse(conn net.Conn, response amqp.ResponseMethodMessage) error {
// 	frame := []byte{}
// 	err :=
// 	if err != nil {
// 		return fmt.Errorf("failed to send response: %v", err)
// 	}

// 	return shared.SendFrame(conn, frame)
// }

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
		err := b.handleHeartbeat(conn)
		if err != nil {
			log.Printf("Error handling heartbeat: %v", err)
		}
		return nil, nil

	default:
		fmt.Printf("Received: %x\n", frame)
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func (b *Broker) handleHeartbeat(conn net.Conn) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Connections[conn].LastHeartbeat = time.Now()
	return nil
}

func (b *Broker) sendHeartbeats(conn net.Conn) {
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
				log.Println("Connection no longer exists in broker")
				return
			}
			b.mu.Unlock()

			// sendHearbeat(conn)
			heartbeatFrame := shared.CreateHeartbeatFrame()
			err := shared.SendFrame(conn, heartbeatFrame)
			if err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Println("Heartbeat sent")

		case <-done:
			log.Println("Stopping heartbeat  goroutine for closed connection")
			return
		}

	}
}

// func sendHearbeat(conn net.Conn) error {
// 	heartbeatFrame := shared.CreateHeartbeatFrame()
// 	if heartbeatFrame == nil {
// 		return fmt.Errorf("Failed to create heartbeat frame")
// 	}
// 	fmt.Printf("Sending heartbeat: %x\n", heartbeatFrame)
// 	err := shared.SendFrame(conn, heartbeatFrame)
// 	if err != nil {
// 		return fmt.Errorf("Fail sending heartbeat: %s\n", err)
// 	}
// 	return nil
// }

func (b *Broker) executeRequest(request *amqp.RequestMethodMessage) (interface{}, error) {
	switch request.ClassID {

	case uint16(constants.CHANNEL):
		// Handle channel-related commands
	case uint16(constants.EXCHANGE):
		switch request.MethodID {
		case uint16(exchange.DECLARE):
			// // Handle exchange declaration
			// exchangeName := parts[1]
			// typ := parts[2]
			// err := b.createExchange(exchangeName, ExchangeType(typ))
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s of type %s created", exchangeName, typ)}, nil
		case uint16(exchange.DELETE):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// err := b.deleteExchange(exchangeName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s deleted", exchangeName)}, nil

		default:
			return nil, fmt.Errorf("unsupported command")
		}
		// Handle exchange-related commands
	case uint16(constants.QUEUE):
		switch request.MethodID {
		case uint16(queue.DECLARE):
			// Handle queue declaration
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// _, err := b.createQueue(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// // Bind the queue to the default exchange with the same name as the queue
			// err = b.bindToDefaultExchange(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s created", queueName)}, nil
		case uint16(queue.BIND):
			// if len(parts) < 3 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// queueName := parts[2]
			// routingKey := ""
			// if len(parts) == 4 {
			// 	routingKey = parts[3]
			// }
			// err := b.bindQueue(exchangeName, queueName, routingKey)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s bound to exchange %s", queueName, exchangeName)}, err
		case uint16(queue.DELETE):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// err := b.deleteQueue(queueName)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s deleted", queueName)}, nil

		case uint16(queue.UNBIND):
			// if len(parts) != 4 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// queueName := parts[2]
			// routingKey := parts[3]
			// b.DeletBinding(exchangeName, queueName, routingKey)
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Binding deleted")}, nil
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	case uint16(constants.BASIC):
		switch request.MethodID {
		case uint16(basic.QOS):
		case uint16(basic.CONSUME):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// queueName := parts[1]
			// fmt.Println("Consuming from queue:", queueName)
			// msg := b.consume(queueName, consumerID)
			// if msg == nil {
			// 	return common.CommandResponse{Status: "OK", Message: "No messages available", Data: ""}, nil
			// }
			// return common.CommandResponse{Status: "OK", Data: msg}, nil
		case uint16(basic.CANCEL):
		case uint16(basic.PUBLISH):
			// if len(parts) < 4 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// exchangeName := parts[1]
			// routingKey := parts[2]
			// message := strings.Join(parts[3:], " ")
			// msgId, err := b.publish(exchangeName, routingKey, message)
			// if err != nil {
			// 	return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
			// }
			// var data struct {
			// 	MessageID string `json:"message_id"`
			// }
			// data.MessageID = msgId
			// return common.CommandResponse{Status: "OK", Message: "Message sent", Data: data}, nil
		case uint16(basic.RETURN):
		case uint16(basic.DELIVER):
		case uint16(basic.GET):
		case uint16(basic.GET_EMPTY):
			// Handle message retrieval
		case uint16(basic.ACK):
			// if len(parts) != 2 {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
			// }
			// msgID := parts[1]
			// // get queue from consumerID
			// consumer, ok := b.Consumers[consumerID]
			// if !ok {
			// 	return common.CommandResponse{Status: "ERROR", Message: "Consumer not found"}, nil
			// }
			// queue := consumer.Queue
			// b.acknowledge(queue, consumerID, msgID)
			// return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Message ID %s acknowledged", msgID)}, nil

		case uint16(basic.REJECT):
		case uint16(basic.RECOVER_ASYNC):
		case uint16(basic.RECOVER):
		case uint16(basic.RECOVER_OK):
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	case uint16(constants.TX):
		// Handle transaction-related commands
		switch request.MethodID {
		case uint16(tx.SELECT):
			// Handle transaction selection
		case uint16(tx.COMMIT):
			// Handle transaction commit
		case uint16(tx.ROLLBACK):
			// Handle transaction rollback
		default:
			return nil, fmt.Errorf("unsupported command")
		}
	default:
		return nil, fmt.Errorf("unsupported command")
	}
	return nil, nil
}
