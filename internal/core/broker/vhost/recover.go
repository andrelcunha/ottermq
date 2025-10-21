package vhost

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
)

// HandleBasicRecover processes a Basic.Recover AMQP frame, handling message requeueing if necessary.
func (vh *VHost) HandleBasicRecover(conn net.Conn, channel uint16, requeue bool) error {
	key := ConnectionChannelKey{conn, channel}

	vh.mu.Lock()
	ch := vh.ChannelDeliveries[key]
	vh.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("no channel delivery state for channel %d", channel)
	}

	ch.mu.Lock()
	unackedMessages := make([]*DeliveryRecord, 0, len(ch.Unacked))
	for _, record := range ch.Unacked {
		unackedMessages = append(unackedMessages, record)
	}
	ch.Unacked = make(map[uint64]*DeliveryRecord)
	ch.mu.Unlock()

	for _, record := range unackedMessages {
		if requeue {
			log.Debug().Msgf("Requeuing message with delivery tag %d on channel %d\n", record.DeliveryTag, channel)
			vh.mu.Lock()
			queue := vh.Queues[record.QueueName]
			vh.mu.Unlock()
			if queue != nil {
				// mark as redelivered for next delivery
				vh.markAsRedelivered(record.Message.ID)
				queue.Push(record.Message)
			}
		} else {
			vh.mu.Lock()
			consumerKey := ConnectionChannelKey{conn, channel}
			consumers := vh.ConsumersByChannel[consumerKey]
			vh.mu.Unlock()
			var targetConsumer *Consumer
			if len(consumers) > 0 {
				for _, c := range consumers {
					if c.Tag == record.ConsumerTag {
						targetConsumer = c
						break
					}
				}
			}
			if targetConsumer != nil {
				err := vh.deliverToConsumer(targetConsumer, record.Message, true)
				if err != nil {
					// Delivery failed, requeue to avoid message loss
					log.Debug().Err(err).Msg("Failed to redeliver recovered message, requeuing")
					vh.mu.Lock()
					q := vh.Queues[record.QueueName]
					vh.mu.Unlock()
					if q != nil {
						q.Push(record.Message)
					}
				}
			} else {
				// consumer no longer exists, requeue
				log.Debug().Str("consumer", record.ConsumerTag).Msgf("Consumer not found for recovered message, requeuing")
				vh.mu.Lock()
				q := vh.Queues[record.QueueName]
				vh.mu.Unlock()
				if q != nil {
					vh.markAsRedelivered(record.Message.ID)
					q.Push(record.Message)
				}
			}
		}
	}
	return nil
}

func (vh *VHost) markAsRedelivered(msgID string) {
	vh.redeliveredMu.Lock()
	vh.redeliveredMessages[msgID] = struct{}{}
	vh.redeliveredMu.Unlock()
}
