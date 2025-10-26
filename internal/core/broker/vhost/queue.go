package vhost

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	Passive    bool      `json:"passive"`
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
				q.mu.Lock()
				q.count--
				q.mu.Unlock()

				log.Debug().Str("queue", q.Name).Str("id", msg.ID).Msg("Delivering message to consumers")
				consumers := vh.GetActiveConsumersForQueue(q.Name)
				notDeliverable := true
				// if there are no consumers, requeue the message and wait
				if len(consumers) == 0 {
					log.Debug().Str("queue", q.Name).Msg("No active consumers for queue, requeuing message")
					q.Push(msg)
					time.Sleep(100 * time.Millisecond) // avoid busy loop
					continue
				}
				for notDeliverable {
					// Round-robin delivery + QoS check
					consumer := q.roundRobinConsumer(consumers, vh)
					if vh.shouldThrottle(consumer, vh.getChannelDeliveryState(consumer.Connection, consumer.Channel)) {
						log.Debug().Str("consumer", consumer.Tag).Msg("Throttling delivery to consumer due to QoS limits")
						continue
					}

					notDeliverable = false
					if err := vh.deliverToConsumer(consumer, msg, false); err != nil {
						log.Error().Err(err).Str("consumer", consumer.Tag).Msg("Delivery failed, removing consumer")
						err := vh.CancelConsumer(consumer.Channel, consumer.Tag)
						if err != nil {
							log.Error().Err(err).Str("consumer", consumer.Tag).Msg("Error cancelling consumer")
						}
						// Requeue the message
						// Requeue just if rejected and `requeue` is true
						q.Push(msg)
					}
				}
			}
		}
	}()
}

func (q *Queue) roundRobinConsumer(consumers []*Consumer, vh *VHost) *Consumer {
	// Simple round-robin delivery
	vh.mu.Lock()
	defer vh.mu.Unlock()
	consumer := consumers[0]
	vh.ConsumersByQueue[q.Name] = append(consumers[1:], consumer)
	return consumer
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

	// Passive declaration: error if queue doesn't exist
	if props != nil && props.Passive {
		if existing, ok := vh.Queues[name]; ok {
			log.Debug().Str("queue", name).Msg("Passive declaration succeeded")
			return existing, nil
		}
		return nil, fmt.Errorf("queue %s does not exist", name)
	}

	// Queue already exists: validate compatibility
	if existing, ok := vh.Queues[name]; ok {
		if existing.Props == nil || props == nil {
			return nil, fmt.Errorf("queue %s already exists with incompatible properties", name)
		}
		if existing.Props.Durable != props.Durable ||
			existing.Props.AutoDelete != props.AutoDelete ||
			existing.Props.Exclusive != props.Exclusive ||
			!equalArgs(existing.Props.Arguments, props.Arguments) {
			return nil, fmt.Errorf("queue %s already exists with different properties", name)
		}
		log.Debug().Str("queue", name).Msg("Queue already exists with matching properties")
		return existing, nil
	}

	// Create new queue
	if props == nil {
		props = NewQueueProperties()
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

func NewQueueProperties() *QueueProperties {
	return &QueueProperties{
		Passive:    false,
		Durable:    false,
		AutoDelete: false,
		Exclusive:  false,
		Arguments:  make(map[string]any),
	}
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

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count
}
