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

	ch.mu.Lock()
	defer ch.mu.Unlock()

	if multiple {
		for tag := range ch.Unacked {
			if tag <= deliveryTag {
				delete(ch.Unacked, tag)
			}
		}
	} else {
		delete(ch.Unacked, deliveryTag)
	}

	return nil
}
