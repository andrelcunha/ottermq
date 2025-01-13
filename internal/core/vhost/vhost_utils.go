package vhost

import (
	"log"
	"net"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
)

func (vh *VHost) CleanupConnection(conn net.Conn) {
	log.Println("Cleaning vhost connection")
	vh.mu.Lock()

	vh.mu.Unlock()
	consumerID, ok := vh.getSessionID(conn)
	if ok {
		vh.handleConsumerDisconnection(consumerID)
	}
}

func (b *VHost) handleConsumerDisconnection(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	consumerID, ok := b.ConsumerSessions[sessionID]
	if !ok {
		log.Printf("Session %s not found\n", sessionID)
		return
	}

	consumer, ok := b.Consumers[consumerID]
	if !ok {
		log.Printf("Consumer %s not found\n", consumerID)
		return
	}

	// requeue unacknowledged messages
	for msgID := range b.ConsumerUnackMsgs[consumerID] {
		if queue, ok := b.Queues[consumer.Queue]; ok {
			// queue.messages <- Message{ID: msgID}
			queue.ReQueue(amqp.Message{ID: msgID})
		}
	}

	delete(b.ConsumerUnackMsgs, consumerID)
	delete(b.ConsumerSessions, sessionID)
	delete(b.Consumers, consumerID)
}
