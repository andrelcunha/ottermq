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
	"github.com/google/uuid"
)

type Broker struct {
	Connections     map[net.Conn]bool    `json:"-"`
	Exchanges       map[string]*Exchange `json:"exchanges"`
	Queues          map[string]*Queue    `json:"queues"`
	UnackedMessages map[string]Message   `json:"unacked_messages"`
	mu              sync.Mutex           `json:"-"`
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

func NewBroker() *Broker {
	b := &Broker{
		Connections:     make(map[net.Conn]bool),
		Exchanges:       make(map[string]*Exchange),
		Queues:          make(map[string]*Queue),
		UnackedMessages: make(map[string]Message),
	}
	b.loadBrokerState()
	b.createExchange("default", DIRECT)
	return b
}

func (b *Broker) handleConnection(conn net.Conn) {
	defer conn.Close()
	b.mu.Lock()
	b.Connections[conn] = true
	b.mu.Unlock()
	log.Println("New connection established")

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Connection closed")
			break
		}

		log.Printf("Received: %s\n", msg)
		response, err := b.processCommand(msg)
		if err != nil {
			log.Println("ERROR: ", err)
			response = common.CommandResponse{
				Status:  "error",
				Message: err.Error(),
			}
		}

		// Serialize the response into JSON
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
	b.mu.Unlock()
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

func (b *Broker) processCommand(command string) (common.CommandResponse, error) {
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
		b.createExchange(exchangeName, ExchangeType(typ))
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Exchange %s of type %s created", exchangeName, typ)}, nil

	case "CREATE_QUEUE":
		if len(parts) != 2 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		_ = b.createQueue(queueName)
		// Bind the queue to the default exchange with the same name as the queue
		err := b.bindQueue("default", queueName, queueName)
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
		b.publish(exchangeName, routingKey, message)
		return common.CommandResponse{Status: "OK", Message: "Message sent"}, nil

	case "CONSUME":
		if len(parts) != 2 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		msg := <-b.consume(queueName)
		return common.CommandResponse{Status: "OK", Data: msg.Content}, nil

	case "ACK":
		if len(parts) != 2 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		msgID := parts[1]
		b.acknowledge(msgID)
		// return fmt.Sprintf("Message ID %s acknowledged", msgID), nil
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Message ID %s acknowledged", msgID)}, nil

	case "DELETE_QUEUE":
		if len(parts) != 2 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queueName := parts[1]
		b.deleteQueue(queueName)
		// return fmt.Sprintf("Queue %s deleted", queueName), nil
		return common.CommandResponse{Status: "OK", Message: fmt.Sprintf("Queue %s deleted", queueName)}, nil

	case "LIST_QUEUES":
		if len(parts) != 1 {
			// return "", fmt.Errorf("Invalid %s command", parts[0])
			return common.CommandResponse{Status: "ERROR", Message: "Invalid command"}, nil
		}
		queues := b.listQueues()
		// queues := strings.Join(queueNames, ", ")
		// return fmt.Sprintf("Queues: %s", queues), nil
		return common.CommandResponse{Status: "OK", Data: queues}, nil

	default:
		// return "", fmt.Errorf("Unknown command '%s'", parts[0])
		return common.CommandResponse{Status: "ERROR", Message: fmt.Sprintf("Unknown command '%s'", parts[0])}, nil
	}
}

// acknowledge removes the message with the given ID frrom the unackedMessages map.
func (b *Broker) acknowledge(msgID string) {
	b.mu.Lock()
	delete(b.UnackedMessages, msgID)
	b.mu.Unlock()
	b.saveBrokerState()
}

func (b *Broker) createQueue(name string) *Queue {
	b.mu.Lock()
	defer b.mu.Unlock()
	queue := &Queue{
		Name:     name,
		messages: make(chan Message, 100),
	}
	b.Queues[name] = queue
	b.saveBrokerState()
	return queue
}

func (b *Broker) publish(exchangeName, routingKey, message string) {
	b.mu.Lock()
	exchange, ok := b.Exchanges[exchangeName]
	b.mu.Unlock()
	if !ok {
		log.Printf("Exchange %s not found", exchangeName)
		return
	}
	msgID := uuid.New().String()
	msg := Message{
		ID:      msgID,
		Content: message,
	}

	// Save message to file
	err := b.saveMessage(routingKey, msg)
	if err != nil {
		log.Printf("Failed to save message to file: %v", err)
	}

	switch exchange.Typ {
	case DIRECT:
		queues, ok := exchange.Bindings[routingKey]
		if ok {
			for _, queue := range queues {
				queue.messages <- msg
				log.Printf("Message %s sent to queue %s", msgID, queue.Name)
			}
		} else {
			log.Printf("Queue %s not found for routing key %s", routingKey, exchangeName)
			return
		}
	case FANOUT:
		for _, queue := range exchange.Queues {
			queue.messages <- msg
			log.Printf("Message %s sent to queue %s", msgID, queue.Name)
		}
	}
}

func (b *Broker) consume(queueName string) <-chan Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	queue, ok := b.Queues[queueName]
	if !ok {
		log.Printf("Queue %s not found", queueName)
		return nil
	}
	return queue.messages
}

func (b *Broker) deleteQueue(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Queues, name)
	b.saveBrokerState()
}

func (b *Broker) listQueues() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	queueNames := make([]string, 0, len(b.Queues))
	for name := range b.Queues {
		queueNames = append(queueNames, name)
	}
	return queueNames
}

func (b *Broker) createExchange(name string, typ ExchangeType) {
	b.mu.Lock()
	defer b.mu.Unlock()
	exchange := &Exchange{
		Name:     name,
		Typ:      typ,
		Queues:   make(map[string]*Queue),
		Bindings: make(map[string][]*Queue),
	}
	b.Exchanges[name] = exchange
}

func (b *Broker) bindQueue(exchangeName, queueName, routingKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if exchangeName == "" {
		exchangeName = "default"
	}

	// Find the exchange
	exchange, ok := b.Exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("Exchange %s not found", exchangeName)
	}

	// Find the queue
	queue, ok := b.Queues[queueName]
	if !ok {
		return fmt.Errorf("Queue %s not found", queueName)
	}

	switch exchange.Typ {
	case DIRECT:
		if exchangeName == "default" {
			log.Printf("Binding queue %s to %s exchange", queueName, exchangeName)
			routingKey = queueName
		} else {
			delete(exchange.Bindings, routingKey)
		}
		exchange.Bindings[routingKey] = append(exchange.Bindings[routingKey], queue)
	case FANOUT:
		exchange.Queues[queueName] = queue
	}

	// Persist the state
	err := b.saveBrokerState()
	if err != nil {
		log.Printf("Failed to save broker state: %v", err)
	}

	return nil
}
