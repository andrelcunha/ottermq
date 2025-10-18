package vhost

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
)

func (vh *VHost) HandleBasicReject(conn net.Conn, channel uint16, deliveryTag uint64, requeue bool) error {
	key := ConnectionChannelKey{conn, channel}

	vh.mu.Lock()
	ch := vh.ChannelDeliveries[key]
	vh.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("no channel delivery state for channel %d", channel)
	}

	ch.mu.Lock()
	record, exists := ch.Unacked[deliveryTag]
	if exists {
		delete(ch.Unacked, deliveryTag)
	}
	ch.mu.Unlock()

	if requeue {
		log.Debug().Msgf("Requeuing message with delivery tag %d on channel %d\n", deliveryTag, channel)
		vh.mu.Lock()
		queue := vh.Queues[record.QueueName]
		vh.mu.Unlock()
		if queue != nil {
			queue.Push(record.Message)
		}
	} else {
		// TODO: implement dead-lettering
		log.Debug().Msgf("Discarding message with delivery tag %d on channel %d\n", deliveryTag, channel)
		if record.Persistent {
			// TODO: persist discard if necessary
			log.Debug().Msgf("Message with delivery tag %d was persistent, consider persisting discard\n", deliveryTag)
		}
	}
	return nil
}
