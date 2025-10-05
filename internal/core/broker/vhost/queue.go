package vhost

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/persistdb/persistence"
)

type QueueArgs map[string]any

type Queue struct {
	Name     string            `json:"name"`
	Props    *QueueProperties  `json:"properties"`
	messages chan amqp.Message `json:"-"`
	count    int               `json:"-"`
	mu       sync.Mutex        `json:"-"`
}

type QueueProperties struct {
	Passive    bool      `json:"passive"`
	Durable    bool      `json:"durable"`
	AutoDelete bool      `json:"auto_delete"`
	Exclusive  bool      `json:"exclusive"`
	NoWait     bool      `json:"no_wait"`
	Arguments  QueueArgs `json:"arguments"`
}

func NewQueue(name string, bufferSize int) *Queue {
	return &Queue{
		Name:     name,
		Props:    &QueueProperties{},
		messages: make(chan amqp.Message, bufferSize),
		count:    0,
	}
}

func (vh *VHost) CreateQueue(name string, props *QueueProperties) (*Queue, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	if queue, ok := vh.Queues[name]; ok {
		log.Debug().Str("queue", name).Msg("Queue already exists")
		return queue, nil
	}
	if props == nil {
		props = &QueueProperties{
			Passive:    false,
			Durable:    false,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments:  make(map[string]any),
		}
	}
	queue := &Queue{
		Name:     name,
		Props:    props,
		messages: make(chan amqp.Message, vh.QueueBufferSize),
		count:    0,
	}

	vh.Queues[name] = queue
	if props.Durable {
		vh.persist.SaveQueue(vh.Name, queue.ToPersistence())
	}

	log.Debug().Str("queue", name).Msg("Created queue")
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

func (q *Queue) ToPersistence() *persistence.PersistedQueue {
	messages := make([]persistence.PersistedMessage, 0)
	return &persistence.PersistedQueue{
		Name:       q.Name,
		Properties: q.Props.ToPersistence(),
		Messages:   messages,
	}
}

func (qp *QueueProperties) ToPersistence() persistence.QProps {
	return persistence.QProps{
		Passive:    qp.Passive,
		Durable:    qp.Durable,
		AutoDelete: qp.AutoDelete,
		Exclusive:  qp.Exclusive,
		NoWait:     qp.NoWait,
		Arguments:  qp.Arguments,
	}
}

func (q *Queue) Push(msg amqp.Message) {
	// queue.messages <- msg
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.messages <- msg:
		q.count++
		log.Debug().Str("queue", q.Name).Str("id", msg.ID).Bytes("body", msg.Body).Msg("Pushed message to queue")
	default:
		log.Debug().Str("queue", q.Name).Str("id", msg.ID).Msg("Queue channel full, dropping message")
	}
}

func (q *Queue) Pop() *amqp.Message {
	// return <-queue.messages
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case msg := <-q.messages:
		q.count--
		log.Debug().Str("queue", q.Name).Str("id", msg.ID).Str("body", string(msg.Body)).Msg("Popped message from queue")
		return &msg
	default:
		log.Debug().Str("queue", q.Name).Msg("Queue is empty")
		return nil
	}
}

func (q *Queue) ReQueue(msg amqp.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.messages <- msg:
		q.count++
		log.Debug().Str("queue", q.Name).Str("id", msg.ID).Bytes("body", msg.Body).Msg("Pushed message to queue")
	default:
		log.Debug().Str("queue", q.Name).Str("id", msg.ID).Msg("Queue channel full, dropping message")
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
	log.Debug().Str("queue", name).Msg("Deleted queue")
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
