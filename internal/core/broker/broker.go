package broker

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/andrelcunha/ottermq/config"
	// "github.com/andrelcunha/ottermq/internal/connection"
	"github.com/andrelcunha/ottermq/internal/core/vhost"
	"github.com/andrelcunha/ottermq/pkg/connection/constants"
	"github.com/andrelcunha/ottermq/pkg/connection/shared"
	_ "github.com/andrelcunha/ottermq/pkg/persistdb"
)

type Broker struct {
	VHosts            map[string]*vhost.VHost
	config            *config.Config         `json:"-"`
	HeartbeatInterval time.Duration          `json:"-"`
	Connections       map[net.Conn]bool      `json:"-"`
	ConnectedAt       map[net.Conn]time.Time `json:"-"`
	LastHeartbeat     map[net.Conn]time.Time `json:"-"`
	mu                sync.Mutex             `json:"-"`
}

func NewBroker(config *config.Config) *Broker {
	b := &Broker{
		VHosts:            make(map[string]*vhost.VHost),
		Connections:       make(map[net.Conn]bool),
		config:            config,
		LastHeartbeat:     make(map[net.Conn]time.Time),
		ConnectedAt:       make(map[net.Conn]time.Time),
		HeartbeatInterval: time.Duration(config.HeartBeatInterval) * time.Second,
	}
	b.VHosts["/"] = vhost.NewVhost("/")
	return b
}

func (b *Broker) Start() {
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
		go b.handleConnection(conn)
	}
}

func (b *Broker) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		b.cleanupConnection(conn)
	}()

	// handler := connection.NewConnectionHandler(conn)
	if err := b.connectionHandshake(conn); err != nil {
		log.Printf("Handshake failed: %v", err)
		return
	}

	log.Println("Handshake successful")

	// keep reading commands in loop
	for {
		frame, err := shared.ReadFrame(conn)
		if err != nil {
			log.Fatalf("ERROR: %v", err.Error())
			return
		}
		fmt.Printf("received: %+v\n", frame)
		_, err = b.ParseFrame(conn, frame)
		if err != nil {
			log.Fatalf("ERROR: %v", err.Error())
		}
	}
}

func (b *Broker) registerConnection(conn net.Conn) {
	b.mu.Lock()
	b.Connections[conn] = true
	b.LastHeartbeat[conn] = time.Now()
	b.ConnectedAt[conn] = time.Now()
	b.mu.Unlock()
}

// func (b *Broker) registerSessionAndConsummer(conn net.Conn) string {
// 	sessionID := generateSessionID()
// 	consumerID := conn.RemoteAddr().String()

// 	b.registerConsumer(consumerID, "default", sessionID)
// 	log.Println("New connection registered")
// 	return consumerID
// }

func (b *Broker) cleanupConnection(conn net.Conn) {
	log.Println("Cleaning connection")
	b.mu.Lock()
	delete(b.Connections, conn)
	delete(b.LastHeartbeat, conn)
	delete(b.ConnectedAt, conn)
	b.mu.Unlock()
	for _, vhost := range b.VHosts {
		vhost.CleanupConnection(conn)
	}
}

func (b *Broker) sendHeartbeat(conn net.Conn) {
	ticker := time.NewTicker(b.HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		if _, ok := b.Connections[conn]; !ok {
			b.mu.Unlock()
			return
		}
		b.mu.Unlock()

		_, err := conn.Write([]byte("HEARTBEAT\n"))
		if err != nil {
			log.Println("Failed to send heartbeat:", err)
			return
		}
	}
}

/*
// func (b *Broker) processCommand(command, consumerID string) (common.CommandResponse, error) {
// 	parts := strings.Fields(command)
// 	if len(parts) == 0 {
// 		return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 	}
//
// 	switch parts[0] {
// 	// case "AUTH":
// 	// 	if len(parts) != 3 {
// 	// 		return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 	// 	}
//
// 	// 	username, password := parts[1], parts[2]
// 	// 	isAuthenticated := b.authenticate(username, password)
// 	// 	var response common.CommandResponse
// 	// 	if isAuthenticated {
// 	// 		response = common.CommandResponse{
// 	// 			Status:  "OK",
// 	// 			Message: fmt.Sprintf("User '%s' authenticated successfully", username),
// 	// 		}
// 	// 	} else {
// 	// 		response = common.CommandResponse{
// 	// 			Status:  "ERROR",
// 	// 			Message: "Invalid credentials",
// 	// 		}
// 	// 	}
// 	// 	return response, nil
//
// 	case "CREATE_EXCHANGE":
// 		if len(parts) != 3 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchangeName := parts[1]
// 		typ := parts[2]
// 		err := b.createExchange(exchangeName, ExchangeType(typ))
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s of type %s created", exchangeName, typ)}, nil
//
// 	case "CREATE_QUEUE":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		queueName := parts[1]
// 		_, err := b.createQueue(queueName)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		// Bind the queue to the default exchange with the same name as the queue
// 		err = b.bindToDefaultExchange(queueName)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s created", queueName)}, nil
//
// 	case "BIND_QUEUE":
// 		if len(parts) < 3 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchangeName := parts[1]
// 		queueName := parts[2]
// 		routingKey := ""
// 		if len(parts) == 4 {
// 			routingKey = parts[3]
// 		}
// 		err := b.bindQueue(exchangeName, queueName, routingKey)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s bound to exchange %s", queueName, exchangeName)}, err
//
// 	case "PUBLISH":
// 		if len(parts) < 4 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchangeName := parts[1]
// 		routingKey := parts[2]
// 		message := strings.Join(parts[3:], " ")
// 		msgId, err := b.publish(exchangeName, routingKey, message)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		var data struct {
// 			MessageID string `json:"message_id"`
// 		}
// 		data.MessageID = msgId
// 		return common.CommandResponse{Status: "OK", Message: "Message sent", Data: data}, nil
//
// 	case "CONSUME":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		queueName := parts[1]
// 		fmt.Println("Consuming from queue:", queueName)
// 		msg := b.consume(queueName, consumerID)
// 		if msg == nil {
// 			return common.CommandResponse{Status: "OK", Message: "No messages available", Data: ""}, nil
// 		}
// 		return common.CommandResponse{Status: "OK", Data: msg}, nil

// 	case "ACK":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		msgID := parts[1]
// 		// get queue from consumerID
// 		consumer, ok := b.Consumers[consumerID]
// 		if !ok {
// 			return common.CommandResponse{Status: "ERROR", Message: "Consumer not found"}, nil
// 		}
// 		queue := consumer.Queue
// 		b.acknowledge(queue, consumerID, msgID)
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Message ID %s acknowledged", msgID)}, nil

// 	case "DELETE_QUEUE":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		queueName := parts[1]
// 		err := b.deleteQueue(queueName)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s deleted", queueName)}, nil

// 	case "LIST_QUEUES":
// 		if len(parts) != 1 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		queues := b.listQueues()
// 		return common.CommandResponse{Status: "OK", Data: queues}, nil

// 	case "COUNT_MESSAGES":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		queueName := parts[1]
// 		count, err := b.countMessages(queueName)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		countResponse := struct {
// 			Count int `json:"count"`
// 		}{Count: count}
// 		return common.CommandResponse{Status: "OK", Data: countResponse}, nil

// 	case "LIST_EXCHANGES":
// 		if len(parts) != 1 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchanges := b.listExchanges()
// 		return common.CommandResponse{Status: "OK", Data: exchanges}, nil

// 	case "LIST_BINDINGS":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchangeName := parts[1]
// 		bindings := b.listBindings(exchangeName)
// 		return common.CommandResponse{Status: "OK", Data: bindings}, nil

// 	case "DELETE_BINDING":
// 		if len(parts) != 4 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchangeName := parts[1]
// 		queueName := parts[2]
// 		routingKey := parts[3]
// 		b.DeletBinding(exchangeName, queueName, routingKey)
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Binding deleted")}, nil

// 	case "DELETE_EXCHANGE":
// 		if len(parts) != 2 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		exchangeName := parts[1]
// 		err := b.deleteExchange(exchangeName)
// 		if err != nil {
// 			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
// 		}
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s deleted", exchangeName)}, nil

// 	case "SUBSCRIBE":
// 		if len(parts) != 3 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		consumerID := parts[1]
// 		queueName := parts[2]
// 		b.subscribe(consumerID, queueName)
// 		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Consumer %s subscribed to queue %s", consumerID, queueName)}, nil

// 	case "LIST_CONNECTIONS":
// 		if len(parts) != 1 {
// 			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
// 		}
// 		connections := b.listConnections()
// 		return common.CommandResponse{Status: "OK", Data: connections}, nil

// 	default:
// 		return common.CommandResponse{Status: "ERROR", Message: fmt.Sprintf("Unknown command '%s'", parts[0])}, nil
// 	}
// }
*/

func (b *Broker) ParseFrame(conn net.Conn, frame []byte) (interface{}, error) {
	if len(frame) < 7 {
		return nil, fmt.Errorf("frame too short")
	}

	frameType := frame[0]
	channel := binary.BigEndian.Uint16(frame[1:3])
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if len(frame) < int(7+payloadSize) {
		return nil, fmt.Errorf("frame too short")
	}
	payload := frame[7:]

	switch frameType {
	case byte(constants.TYPE_METHOD):
		fmt.Printf("Received METHOD frame on channel %d\n", channel)
		return shared.ParseMethodFrame(channel, payload)

	case byte(constants.TYPE_HEARTBEAT):
		log.Printf("Received HEARTBEAT frame on channel %d\n", channel)
		err := b.handleHeartbeat(conn, 0, frame)
		return nil, err

	default:
		fmt.Printf("Received: %x\n", frame)
		return nil, fmt.Errorf("unknown frame type: %d", frameType)
	}
}

func (b *Broker) handleHeartbeat(conn net.Conn, channel int, frame []byte) error {
	if err := b.sendFrame(conn, frame); err != nil {
		log.Printf("Error sending heartbeat response: %v", err)
		return err
	}
	return nil
}
