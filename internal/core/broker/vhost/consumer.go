package vhost

import (
	"fmt"
	"net"

	"github.com/google/uuid"
)

type ConsumerKey struct {
	Channel uint16
	Tag     string
}

// ConnectionChannelKey uniquely identifies a channel within a connection
type ConnectionChannelKey struct {
	Connection net.Conn
	Channel    uint16
}

type Consumer struct {
	Tag         string
	Channel     uint16
	QueueName   string
	Connection  net.Conn
	DeliveryTag uint64 // ?? What is this for?
	Active      bool
	Props       *ConsumerProperties
}

type ConsumerProperties struct {
	NoAck     bool           `json:"no_ack"`
	Exclusive bool           `json:"exclusive"`
	Arguments map[string]any `json:"arguments"`
}

func NewConsumer(conn net.Conn, channel uint16, queueName, consumerTag string, props *ConsumerProperties) *Consumer {
	if consumerTag == "" {
		consumerTag = uuid.New().String() // Verify if isn't it overkill -- maybe a small random string would be enough, like 8 chars
	}
	return &Consumer{
		Tag:        consumerTag,
		Channel:    channel,
		QueueName:  queueName,
		Connection: conn,
		Props:      props,
	}
}

func (vh *VHost) RegisterConsumer(consumer *Consumer) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	key := ConsumerKey{consumer.Channel, consumer.Tag}
	// Check if queue exists
	_, ok := vh.Queues[consumer.QueueName]
	if !ok {
		// TODO: return a channel exception error - 404
		return fmt.Errorf("queue %s does not exist", consumer.QueueName)
	}

	// Check for duplicates
	if _, exists := vh.Consumers[key]; exists {
		// TODO: return a channel exception error
		return fmt.Errorf("consumer with tag %s already exists on channel %d", key.Tag, key.Channel)
	}

	// Check if exclusive consumer already exists for this queue
	// or if trying to add exclusive when others exist
	if existingConsumers, exists := vh.ConsumersByQueue[consumer.QueueName]; exists {
		for _, c := range existingConsumers {
			if c.Props.Exclusive {
				// TODO: return a channel exception error - 405
				return fmt.Errorf("exclusive consumer already exists for queue %s", consumer.QueueName)
			}
		}
		if consumer.Props.Exclusive && len(existingConsumers) > 0 {
			// TODO: return a channel exception error - 405
			return fmt.Errorf("cannot add exclusive consumer when other consumers exist for queue %s", consumer.QueueName)
		}
	}

	// Activate consumer and register
	consumer.Active = true
	vh.Consumers[key] = consumer

	// Index for delivery
	// verify if map entry exists
	if _, exists := vh.ConsumersByQueue[consumer.QueueName]; !exists {
		vh.ConsumersByQueue[consumer.QueueName] = []*Consumer{}
	}
	vh.ConsumersByQueue[consumer.QueueName] = append(
		vh.ConsumersByQueue[consumer.QueueName],
		consumer,
	)
	// Index for channel (connection-scoped)
	channelKey := ConnectionChannelKey{consumer.Connection, consumer.Channel}
	if _, exists := vh.ConsumersByChannel[channelKey]; !exists {
		vh.ConsumersByChannel[channelKey] = []*Consumer{}
	}
	vh.ConsumersByChannel[channelKey] = append(
		vh.ConsumersByChannel[channelKey],
		consumer,
	)

	return nil
}

func (vh *VHost) CancelConsumer(channel uint16, tag string) error {
	key := ConsumerKey{channel, tag}
	vh.mu.Lock()
	defer vh.mu.Unlock()
	consumer, exists := vh.Consumers[key]
	if !exists {
		return fmt.Errorf("consumer with tag %s on channel %d does not exist", tag, channel)
	}

	consumer.Active = false
	delete(vh.Consumers, key)

	// Remove from ConsumersByQueue
	consumersForQueue := vh.ConsumersByQueue[consumer.QueueName]
	for i, c := range consumersForQueue {
		if c.Tag == tag && c.Channel == channel {
			vh.ConsumersByQueue[consumer.QueueName] = append(consumersForQueue[:i], consumersForQueue[i+1:]...)
			break
		}
	}
	if len(vh.ConsumersByQueue[consumer.QueueName]) == 0 {
		delete(vh.ConsumersByQueue, consumer.QueueName)
	}

	// Remove from ConsumersByChannel (connection-scoped)
	channelKey := ConnectionChannelKey{consumer.Connection, consumer.Channel}
	consumersForChannel := vh.ConsumersByChannel[channelKey]
	for i, c := range consumersForChannel {
		if c.Tag == tag {
			vh.ConsumersByChannel[channelKey] = append(consumersForChannel[:i], consumersForChannel[i+1:]...)
			break
		}
	}
	// Clean up empty channel entries
	if len(vh.ConsumersByChannel[channelKey]) == 0 {
		delete(vh.ConsumersByChannel, channelKey)
	}

	return nil
}

func (vh *VHost) CleanupChannel(connection net.Conn, channel uint16) {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	channelKey := ConnectionChannelKey{connection, channel}
	// Get copy of consumers to avoid modification during iteration
	consumersForChannel, exists := vh.ConsumersByChannel[channelKey]
	if !exists {
		return // No consumers on this channel
	}

	consumersCopy := make([]*Consumer, len(consumersForChannel))
	copy(consumersCopy, consumersForChannel)

	// Cancel each consumer (this will modify the maps)
	for _, c := range consumersCopy {
		// Unlock to call CancelConsumer (which also locks)
		vh.mu.Unlock()
		vh.CancelConsumer(c.Channel, c.Tag)
		vh.mu.Lock()
	}
}

func (vh *VHost) CleanupConnection(connection net.Conn) {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	// Find all consumers for this connection
	var consumersToRemove []*Consumer
	for _, consumer := range vh.Consumers {
		if consumer.Connection == connection {
			consumersToRemove = append(consumersToRemove, consumer)
		}
	}

	// Cancel each consumer
	for _, consumer := range consumersToRemove {
		// Unlock to call CancelConsumer (which also locks)
		vh.mu.Unlock()
		vh.CancelConsumer(consumer.Channel, consumer.Tag)
		vh.mu.Lock()
	}
}
