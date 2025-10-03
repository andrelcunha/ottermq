package vhost

import (
	"fmt"
	"log"
	"sync"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

type QueueArgs map[string]any

type Queue struct {
	Name       string            `json:"name"`
	Durable    bool              `json:"durable"`
	Exclusive  bool              `json:"exclusive"`
	AutoDelete bool              `json:"auto_delete"`
	MessageTTL int               `json:"message_ttl"`
	Arguments  QueueArgs         `json:"arguments"`
	messages   chan amqp.Message `json:"-"`
	count      int               `json:"-"`
	mu         sync.Mutex        `json:"-"`
}

func NewQueue(name string) *Queue {
	return &Queue{
		Name:       name,
		Durable:    false,
		Exclusive:  false,
		AutoDelete: false,
		MessageTTL: 0,
		Arguments:  make(QueueArgs),
		messages:   make(chan amqp.Message, 100000), // Adjustable buffer size
		count:      0,
	}
}

func (vh *VHost) CreateQueue(name string) (*Queue, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	if queue, ok := vh.Queues[name]; ok {
		log.Printf("[DEBUG] Queue %s already exists", name)
		return queue, nil
	}
	queue := NewQueue(name)
	vh.Queues[name] = queue
	log.Printf("[DEBUG] Created queue %s", name)
	// vh.saveBrokerState() // TODO: persist state
	// adminQueues := make(map[string]bool)
	// queues := []string{ADMIN_QUEUES, ADMIN_EXCHANGES, ADMIN_BINDINGS, ADMIN_CONNECTIONS}
	// for _, queueName := range queues {
	// 	adminQueues[queueName] = true
	// }

	// if _, ok := adminQueues[name]; !ok {
	// 	vh.publishQueueUpdate()
	// }
	return queue, nil
}

func (q *Queue) Push(msg amqp.Message) {
	// queue.messages <- msg
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.messages <- msg:
		q.count++
		log.Printf("[DEBUG] Pushed message to queue %s: ID=%s, Body=%s", q.Name, msg.ID, msg.Body)
	default:
		log.Printf("[DEBUG] Queue %s channel full, dropping message: ID=%s", q.Name, msg.ID)
	}
}

func (q *Queue) Pop() *amqp.Message {
	// return <-queue.messages
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case msg := <-q.messages:
		q.count--
		log.Printf("[DEBUG] Popped message from queue %s: ID=%s, Body=%s", q.Name, msg.ID, string(msg.Body))
		return &msg
	default:
		log.Printf("[DEBUG] Queue %s is empty", q.Name)
		return nil
	}
}

func (q *Queue) ReQueue(msg amqp.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.messages <- msg:
		q.count++
		log.Printf("[DEBUG] Pushed message to queue %s: ID=%s, Body=%s", q.Name, msg.ID, msg.Body)
	default:
		log.Printf("[DEBUG] Queue %s channel full, dropping message: ID=%s", q.Name, msg.ID)
	}
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count
}

func (vh *VHost) DeleteQueue(name string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, exists := vh.Queues[name]
	if !exists {
		return fmt.Errorf("queue %s not found", name)
	}
	close(queue.messages)
	delete(vh.Queues, name)
	log.Printf("[DEBUG] Deleted queue %s", name)
	// vh.publishQueueUpdate()
	return nil
}

// func (vh *VHost) subscribe(consumerID, queueName string) {
// 	vh.mu.Lock()
// 	defer vh.mu.Unlock()
// 	if _, ok := vh.Queues[queueName]; !ok {
// 		vh.CreateQueue(queueName)
// 	}
// 	if consumer, ok := vh.Consumers[consumerID]; ok {
// 		consumer.Queue = queueName
// 		log.Printf("[DEBUG] Subscribed consumer %s to queue %s", consumerID, queueName)
// 	}
// }
