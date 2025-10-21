package vhost

import (
	"fmt"
	"sync"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/rs/zerolog/log"
)

type DeliveryRecord struct {
	DeliveryTag uint64
	ConsumerTag string
	QueueName   string
	Message     amqp.Message
	Persistent  bool
}

type ChannelDeliveryState struct {
	mu              sync.Mutex
	LastDeliveryTag uint64
	Unacked         map[uint64]*DeliveryRecord // deliveryTag -> DeliveryRecord
}

func (vh *VHost) deliverToConsumer(consumer *Consumer, msg amqp.Message, redelivered bool) error {
	if !redelivered {
		redelivered = vh.shouldRedeliver(msg.ID)
	}

	if !consumer.Active {
		return fmt.Errorf("consumer %s on channel %d is not active", consumer.Tag, consumer.Channel)
	}
	channelKey := ConnectionChannelKey{consumer.Connection, consumer.Channel}
	vh.mu.Lock()
	ch := vh.ChannelDeliveries[channelKey]
	if ch == nil {
		ch = &ChannelDeliveryState{Unacked: make(map[uint64]*DeliveryRecord)}
		vh.ChannelDeliveries[channelKey] = ch
	}
	vh.mu.Unlock()

	ch.mu.Lock()
	ch.LastDeliveryTag++
	tag := ch.LastDeliveryTag

	track := !consumer.Props.NoAck // only track when manual ack is required
	if track {
		ch.Unacked[tag] = &DeliveryRecord{
			DeliveryTag: tag,
			ConsumerTag: consumer.Tag,
			QueueName:   consumer.QueueName,
			Message:     msg,
			Persistent:  msg.Properties.DeliveryMode == amqp.PERSISTENT,
		}
	}
	ch.mu.Unlock()

	deliverFrame := vh.framer.CreateBasicDeliverFrame(
		consumer.Channel,
		consumer.Tag,
		msg.Exchange,
		msg.RoutingKey,
		tag,
		redelivered,
	)
	headerFrame := vh.framer.CreateHeaderFrame(consumer.Channel, uint16(amqp.BASIC), msg)
	bodyFrame := vh.framer.CreateBodyFrame(consumer.Channel, msg.Body)

	if err := vh.framer.SendFrame(consumer.Connection, deliverFrame); err != nil {
		log.Error().Err(err).Msg("Failed to send deliver frame")
		if track {
			ch.mu.Lock()
			delete(ch.Unacked, tag)
			ch.mu.Unlock()
		}
		return err
	}
	if err := vh.framer.SendFrame(consumer.Connection, headerFrame); err != nil {
		log.Error().Err(err).Msg("Failed to send header frame")
		if track {
			ch.mu.Lock()
			delete(ch.Unacked, tag)
			ch.mu.Unlock()
		}
		return err
	}
	if err := vh.framer.SendFrame(consumer.Connection, bodyFrame); err != nil {
		log.Error().Err(err).Msg("Failed to send body frame")
		if track {
			ch.mu.Lock()
			delete(ch.Unacked, tag)
			ch.mu.Unlock()
		}
		return err
	}

	if redelivered {
		vh.clearRedeliveredMark(msg.ID)
	}

	// Persistence
	if consumer.Props.NoAck && vh.persist != nil && msg.Properties.DeliveryMode == amqp.PERSISTENT {
		if err := vh.persist.DeleteMessage(vh.Name, consumer.QueueName, msg.ID); err != nil {
			log.Error().Err(err).Msg("Failed to delete persisted message after delivery with no-ack")
		}
	}
	return nil
}

func (vh *VHost) shouldRedeliver(msgID string) bool {
	vh.redeliveredMu.Lock()
	_, exists := vh.redeliveredMessages[msgID]
	vh.redeliveredMu.Unlock()
	return exists
}

func (vh *VHost) clearRedeliveredMark(msgID string) {
	vh.redeliveredMu.Lock()
	delete(vh.redeliveredMessages, msgID)
	vh.redeliveredMu.Unlock()
}
