package vhost

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
)

func (vh *VHost) HandleBasicAck(conn net.Conn, channel uint16, deliveryTag uint64, multiple bool) error {
	key := ConnectionChannelKey{conn, channel}

	vh.mu.Lock()
	ch := vh.ChannelDeliveries[key]
	vh.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("no channel delivery state for channel %d", channel)
	}

	var toDelete []*DeliveryRecord

	ch.mu.Lock()
	if multiple {
		for tag := range ch.Unacked {
			if tag <= deliveryTag {
				toDelete = append(toDelete, ch.Unacked[tag])
				delete(ch.Unacked, tag)
			}
		}
	} else {
		if rec, exists := ch.Unacked[deliveryTag]; exists {
			toDelete = append(toDelete, rec)
			delete(ch.Unacked, deliveryTag)
		}
	}
	ch.mu.Unlock()

	if vh.persist != nil {
		for _, rec := range toDelete {
			if rec.Persistent {
				_ = vh.persist.DeleteMessage(vh.Name, rec.QueueName, rec.Message.ID)
			}
		}
	}

	return nil
}

func (vh *VHost) HandleBasicReject(conn net.Conn, channel uint16, deliveryTag uint64, requeue bool) error {
	return vh.HandleBasicNack(conn, channel, deliveryTag, false, requeue)
}

func (vh *VHost) HandleBasicNack(conn net.Conn, channel uint16, deliveryTag uint64, multiple, requeue bool) error {
	key := ConnectionChannelKey{conn, channel}

	vh.mu.Lock()
	ch := vh.ChannelDeliveries[key]
	vh.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("no channel delivery state for channel %d", channel)
	}

	ch.mu.Lock()
	recordsToNack := make([]*DeliveryRecord, 0)
	if multiple {
		for tag, record := range ch.Unacked {
			if tag <= deliveryTag {
				recordsToNack = append(recordsToNack, record)
				delete(ch.Unacked, tag)
			}
		}
	} else {
		if record, exists := ch.Unacked[deliveryTag]; exists {
			recordsToNack = append(recordsToNack, record)
			delete(ch.Unacked, deliveryTag)
		}
	}
	ch.mu.Unlock()
	if len(recordsToNack) == 0 {
		return nil
	}

	if requeue {
		for _, record := range recordsToNack {
			log.Debug().Msgf("Requeuing message with delivery tag %d on channel %d\n", record.DeliveryTag, channel)
			vh.markAsRedelivered(record.Message.ID)
			vh.mu.Lock()
			queue := vh.Queues[record.QueueName]
			vh.mu.Unlock()
			if queue != nil {
				queue.Push(record.Message)
			}
		}
		return nil
	}
	// TODO: implement dead-lettering
	for _, record := range recordsToNack {
		log.Debug().Uint64("delivery_tag", record.DeliveryTag).Uint16("channel", channel).Msg("Nack: discarding message")
		if record.Persistent {
			_ = vh.persist.DeleteMessage(vh.Name, record.QueueName, record.Message.ID)
		}
	}
	return nil
}

func (vh *VHost) HandleBasicQos(conn net.Conn, channel uint16, prefetchCount uint16, global bool) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	key := ConnectionChannelKey{conn, channel}
	state := vh.ChannelDeliveries[key]
	// fetch the channel delivery state if global is true
	if state == nil {
		state = &ChannelDeliveryState{
			Unacked: make(map[uint64]*DeliveryRecord),
		}
		vh.ChannelDeliveries[key] = state
	}
	if global {
		state.GlobalPrefetchCount = prefetchCount
		state.PrefetchGlobal = true
	} else {
		// Store the prefetch count to apply to new consumers if global is false
		state.NextPrefetchCount = prefetchCount
	}
	return nil
}
