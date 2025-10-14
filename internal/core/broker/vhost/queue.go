package vhost

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/pkg/persistence"
)

type QueueArgs map[string]any

type Queue struct {
	Name     string            `json:"name"`
	Props    *QueueProperties  `json:"properties"`
	messages chan amqp.Message `json:"-"`
	count    int               `json:"-"`
	mu       sync.Mutex        `json:"-"`
	/* Delivery */
	deliveryCtx    context.Context    `json:"-"`
	deliveryCancel context.CancelFunc `json:"-"`
	delivering     bool
}

type QueueProperties struct {
	Passive    bool      `json:"passive"` // TODO: #126 remove Passive Property and instead implement it at DeclareQueue
	Durable    bool      `json:"durable"`
	AutoDelete bool      `json:"auto_delete"`
	Exclusive  bool      `json:"exclusive"` // not implemented yet
	Arguments  QueueArgs `json:"arguments"`
}

func NewQueue(name string, bufferSize int) *Queue {

	return &Queue{
		Name:       name,
		Props:      &QueueProperties{},
		messages:   make(chan amqp.Message, bufferSize),
		count:      0,
		delivering: false,
	}
}

func (q *Queue) startDeliveryLoop(vh *VHost) {
	if q.delivering {
		return // already running
	}
	q.deliveryCtx, q.deliveryCancel = context.WithCancel(context.Background())
	q.delivering = true
	go func() {
		for {
			select {
			case <-q.deliveryCtx.Done():
				log.Debug().Str("queue", q.Name).Msg("Stopping delivery loop")
				q.delivering = false
				return
			case msg := <-q.messages:
				// Here we would deliver the message to consumers
				log.Debug().Str("queue", q.Name).Str("id", msg.ID).Msg("Delivering message to consumers")
				consumers := vh.GetActiveConsumersForQueue(q.Name)
				if len(consumers) > 0 {
					// Simple round-robin delivery
					consumer := consumers[0]
					vh.ConsumersByQueue[q.Name] = append(consumers[1:], consumer)
					// TODO: improve delivery strategy using basic.qos and manual ack
					if err := vh.deliverToConsumer(consumer, msg); err != nil {
						log.Error().Err(err).Str("consumer", consumer.Tag).Msg("Delivery failed, removing consumer")
						vh.CancelConsumer(consumer.Channel, consumer.Tag)
						// Requeue the message
						// Requeue just if rejected and `requeue` is true
					}
				}
			}
		}
	}()
}

func (q *Queue) stopDeliveryLoop() {
	if !q.delivering {
		return
	}
	q.deliveryCancel()
	q.delivering = false
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
			Arguments:  make(map[string]any),
		}
	}
	queue := NewQueue(name, vh.queueBufferSize)
	queue.Props = props

	vh.Queues[name] = queue
	if props.Durable {
		if err := vh.persist.SaveQueueMetadata(vh.Name, name, props.ToPersistence()); err != nil {
			log.Error().Err(err).Str("queue", name).Msg("Failed to save queue metadata")
		}
	}

	log.Debug().Str("queue", name).Msg("Created queue")

	return queue, nil
}

func (vh *VHost) DeleteQueue(name string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	queue, exists := vh.Queues[name]
	if !exists {
		return fmt.Errorf("queue %s not found", name)
	}
	close(queue.messages)
	// verify if there are any bindings to this queue and remove them
	for _, exchange := range vh.Exchanges {
		// Routing keys to bindings
		switch exchange.Typ {
		case DIRECT:
			for rk, queues := range exchange.Bindings {
				for i, q := range queues {
					if q.Name == name {
						exchange.Bindings[rk] = append(queues[:i], queues[i+1:]...)
						break
					}
				}
				if len(exchange.Bindings[rk]) == 0 {
					delete(exchange.Bindings, rk)
					// Check if the exchange can be auto-deleted
					if deleted, err := vh.checkAutoDeleteExchangeUnlocked(exchange.Name); err != nil {
						log.Printf("Failed to check auto-delete exchange: %v", err)
					} else if deleted {
						log.Printf("Exchange %s was auto-deleted", exchange.Name)
					}
				}
			}
		case FANOUT:
			if _, ok := exchange.Queues[name]; ok {
				delete(exchange.Queues, name)
				// Check if the exchange can be auto-deleted
				if deleted, err := vh.checkAutoDeleteExchangeUnlocked(exchange.Name); err != nil {
					log.Printf("Failed to check auto-delete exchange: %v", err)
				} else if deleted {
					log.Printf("Exchange %s was auto-deleted", exchange.Name)
				}
			}
		}
	}

	// Remove the queue from the VHost's queue map
	delete(vh.Queues, name)

	log.Debug().Str("queue", name).Msg("Deleted queue")
	// Call persistence layer to delete the queue
	if queue.Props.Durable {
		if err := vh.persist.DeleteQueueMetadata(vh.Name, name); err != nil {
			log.Error().Err(err).Str("queue", name).Msg("Failed to delete queue from persistence")
			return err
		}
	}
	// vh.publishQueueUpdate()
	return nil
}

func (qp *QueueProperties) ToPersistence() persistence.QueueProperties {
	return persistence.QueueProperties{
		Passive:    qp.Passive, //TODO(#126) remove Passive from persistence
		Durable:    qp.Durable,
		AutoDelete: qp.AutoDelete,
		Exclusive:  qp.Exclusive,
		Arguments:  qp.Arguments,
	}
}

func (q *Queue) Push(msg amqp.Message) {
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
