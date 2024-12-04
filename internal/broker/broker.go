package broker

import (
	"bufio"
	"log"
	"net"
	"sync"
)

type Broker struct {
	connections map[net.Conn]bool
	queues      map[string]*Queue
	mu          sync.Mutex
}

type Queue struct {
	name     string
	messages chan string
}

func NewBroker() *Broker {
	return &Broker{
		connections: make(map[net.Conn]bool),
		queues:      make(map[string]*Queue),
	}
}

func (b *Broker) handleConnection(conn net.Conn) {
	defer conn.Close()
	b.mu.Lock()
	b.connections[conn] = true
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
		// Parse and handle the message
		// Example: if message == "CREATE_QUEUE testQueue\n" {
		// Call b.CreateQueue("testQueue")
		// }
	}
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

func (b *Broker) CreateQueue(name string) *Queue {
	b.mu.Lock()
	defer b.mu.Unlock()
	queue := &Queue{
		name:     name,
		messages: make(chan string, 100), // Buffer size can be adjusted
	}
	b.queues[name] = queue
	return queue
}

func (b *Broker) Publish(queueName, message string) {
	b.mu.Lock()
	queue, ok := b.queues[queueName]
	b.mu.Unlock()
	if !ok {
		log.Printf("Queue %s not found", queueName)
		return
	}
	queue.messages <- message
}

func (b *Broker) Consume(queueName string) <-chan string {
	b.mu.Lock()
	queue, ok := b.queues[queueName]
	b.mu.Unlock()
	if !ok {
		log.Printf("Queue %s not found", queueName)
		return nil
	}
	return queue.messages
}
