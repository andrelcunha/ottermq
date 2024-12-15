package broker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/andrelcunha/ottermq/pkg/common"
)

type Broker struct {
	Connections       map[net.Conn]bool          `json:"-"`
	Exchanges         map[string]*Exchange       `json:"exchanges"`
	Queues            map[string]*Queue          `json:"queues"`
	UnackMsgs         map[string]map[string]bool `json:"unacked_messages"`
	Consumers         map[string]*Consumer       `json:"consumers"`
	ConsumerSessions  map[string]string          `json:"consumer_sessions"`
	ConsumerUnackMsgs map[string]map[string]bool `json:"consumer_unacked_messages"`
	mu                sync.Mutex                 `json:"-"`
}

type Exchange struct {
	Name     string              `json:"name"`
	Queues   map[string]*Queue   `json:"queues"`
	Typ      ExchangeType        `json:"type"`
	Bindings map[string][]*Queue `json:"bindings"`
}

type ExchangeType string

const (
	DIRECT ExchangeType = "direct"
	FANOUT ExchangeType = "fanout"
)

type Queue struct {
	Name     string       `json:"name"`
	messages chan Message `json:"-"`
}

type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type Consumer struct {
	ID        string `json:"id"`
	Queue     string `json:"queue"`
	SessionID string `json:"session_id"`
}

func NewBroker() *Broker {
	b := &Broker{
		Connections:       make(map[net.Conn]bool),
		Exchanges:         make(map[string]*Exchange),
		Queues:            make(map[string]*Queue),
		UnackMsgs:         make(map[string]map[string]bool),
		Consumers:         make(map[string]*Consumer),
		ConsumerSessions:  make(map[string]string),
		ConsumerUnackMsgs: make(map[string]map[string]bool),
	}
	b.loadBrokerState()
	b.createExchange("default", DIRECT)
	b.createQueue("default")
	b.bindQueue("default", "default", "default")
	return b
}

func (b *Broker) Start(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start broker: %v", err)
	}
	defer listener.Close()
	log.Printf("Broker listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go b.handleConnection(conn)
	}
}

func (b *Broker) handleConnection(conn net.Conn) {
	defer conn.Close()
	b.mu.Lock()
	b.Connections[conn] = true
	b.mu.Unlock()
	log.Println("New connection established")

	sessionID := generateSessionID()
	consumerID := conn.RemoteAddr().String()

	b.registerConsumer(consumerID, "default", sessionID)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Connection closed")
			break
		}

		log.Printf("Received: %s\n", msg)
		response, err := b.processCommand(msg, consumerID)
		if err != nil {
			log.Println("ERROR: ", err)
			response = common.CommandResponse{
				Status:  "error",
				Message: err.Error(),
			}
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Println("Failed to serialize response:", err)
			continue
		}
		_, err = conn.Write(append(responseJSON, '\n'))
		if err != nil {
			log.Println("Failed to write response:", err)
			break
		}
	}
	b.mu.Lock()
	delete(b.Connections, conn)
	b.handleConsumerDisconnection(sessionID)
	b.mu.Unlock()
}

func (b *Broker) registerConsumer(consumerID, queue, sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	consumer := &Consumer{
		ID:        consumerID,
		Queue:     queue,
		SessionID: sessionID,
	}
	b.Consumers[consumerID] = consumer
	b.ConsumerSessions[sessionID] = consumerID
	b.ConsumerUnackMsgs[consumerID] = make(map[string]bool)
}

func (b *Broker) handleConsumerDisconnection(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	consumerID, ok := b.ConsumerSessions[sessionID]
	if !ok {
		log.Printf("Session %s not found\n", sessionID)
		return
	}

	consumer, ok := b.Consumers[consumerID]
	if !ok {
		log.Printf("Consumer %s not found\n", consumerID)
		return
	}

	for msgID := range b.ConsumerUnackMsgs[consumerID] {
		if queue, ok := b.Queues[consumer.Queue]; ok {
			queue.messages <- Message{ID: msgID}
		}
	}

	delete(b.ConsumerUnackMsgs, consumerID)
	delete(b.ConsumerSessions, sessionID)
	delete(b.Consumers, consumerID)
}

func (b *Broker) processCommand(command, consumerID string) (common.CommandResponse, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		// return "", fmt.Errorf("Invalid command")
		return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
	}

	switch parts[0] {
	case "CREATE_EXCHANGE":
		if len(parts) != 3 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchangeName := parts[1]
		typ := parts[2]
		err := b.createExchange(exchangeName, ExchangeType(typ))
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s of type %s created", exchangeName, typ)}, nil

	case "CREATE_QUEUE":
		if len(parts) != 2 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		_, err := b.createQueue(queueName)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		// Bind the queue to the default exchange with the same name as the queue
		err = b.bindQueue("default", queueName, queueName)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s created", queueName)}, nil

	case "BIND_QUEUE":
		if len(parts) < 3 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchangeName := parts[1]
		queueName := parts[2]
		routingKey := ""
		if len(parts) == 4 {
			routingKey = parts[3]
		}
		err := b.bindQueue(exchangeName, queueName, routingKey)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s bound to exchange %s", queueName, exchangeName)}, err

	case "PUBLISH":
		if len(parts) < 4 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchangeName := parts[1]
		routingKey := parts[2]
		message := strings.Join(parts[3:], " ")
		msgId, err := b.publish(exchangeName, routingKey, message)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		// create type message response with message id
		var data struct {
			MessageID string `json:"message_id"`
		}
		data.MessageID = msgId
		return common.CommandResponse{Status: "OK", Message: "Message sent", Data: data}, nil

	case "CONSUME":
		if len(parts) != 2 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		fmt.Println("Consuming from queue:", queueName)
		msg := b.consume(queueName, consumerID)
		if msg == nil {
			return common.CommandResponse{Status: "OK", Message: "No messages available", Data: ""}, nil
		}
		return common.CommandResponse{Status: "OK", Data: msg}, nil

	case "ACK":
		if len(parts) != 2 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		msgID := parts[1]
		// get queue from consumerID
		consumer, ok := b.Consumers[consumerID]
		if !ok {
			return common.CommandResponse{Status: "ERROR", Message: "Consumer not found"}, nil
		}
		queue := consumer.Queue
		b.acknowledge(queue, consumerID, msgID)
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Message ID %s acknowledged", msgID)}, nil

	case "DELETE_QUEUE":
		if len(parts) != 2 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		err := b.deleteQueue(queueName)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s deleted", queueName)}, nil

	case "LIST_QUEUES":
		if len(parts) != 1 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queues := b.listQueues()
		return common.CommandResponse{Status: "OK", Data: queues}, nil

	case "COUNT_MESSAGES":
		if len(parts) != 2 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		count, err := b.countMessages(queueName)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		countResponse := struct {
			Count int `json:"count"`
		}{Count: count}
		return common.CommandResponse{Status: "OK", Data: countResponse}, nil

	case "LIST_EXCHANGES":
		if len(parts) != 1 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchanges := b.listExchanges()
		return common.CommandResponse{Status: "OK", Data: exchanges}, nil

	case "LIST_BINDINGS":
		if len(parts) != 2 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchangeName := parts[1]
		bindings := b.listBindings(exchangeName)
		return common.CommandResponse{Status: "OK", Data: bindings}, nil

	case "DELETE_BINDING":
		if len(parts) != 4 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchangeName := parts[1]
		queueName := parts[2]
		routingKey := parts[3]
		b.DeletBinding(exchangeName, queueName, routingKey)
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Binding deleted")}, nil

	case "DELETE_EXCHANGE":
		if len(parts) != 2 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		exchangeName := parts[1]
		err := b.deleteExchange(exchangeName)
		if err != nil {
			return common.CommandResponse{Status: "ERROR", Message: err.Error()}, nil
		}
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s deleted", exchangeName)}, nil

	case "SUBSCRIBE":
		if len(parts) != 3 {
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		consumerID := parts[1]
		queueName := parts[2]
		b.subscribe(consumerID, queueName)
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Consumer %s subscribed to queue %s", consumerID, queueName)}, nil

	default:
		return common.CommandResponse{Status: "ERROR", Message: fmt.Sprintf("Unknown command '%s'", parts[0])}, nil
	}
}
