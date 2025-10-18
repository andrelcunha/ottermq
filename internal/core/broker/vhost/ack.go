package vhost

import (
	"fmt"
	"net"
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
