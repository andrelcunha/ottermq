package vhost

import (
	"log"
	"net"
	"sync"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
	"github.com/andrelcunha/ottermq/pkg/persistdb"
	"github.com/google/uuid"
)

type VHost struct {
	Name      string                     `json:"name"`
	Id        string                     `json:"id"`
	Exchanges map[string]*Exchange       `json:"exchanges"`
	Queues    map[string]*Queue          `json:"queues"`
	Users     map[string]*persistdb.User `json:"users"`

	// UnackMsgs         map[string]map[string]bool `json:"unacked_messages"`
	Consumers         map[string]*Consumer       `json:"consumers"`
	ConsumerSessions  map[string]string          `json:"consumer_sessions"`
	ConsumerUnackMsgs map[string]map[string]bool `json:"consumer_unacked_messages"`
	mu                sync.Mutex                 `json:"-"`
}

type Consumer struct {
	ID        string `json:"id"`
	Queue     string `json:"queue"`
	SessionID string `json:"session_id"`
}

func NewVhost(vhostName string) *VHost {
	// generate a random id
	id := uuid.New().String()
	vh := &VHost{
		Name:      vhostName,
		Id:        id,
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
		Users:     make(map[string]*persistdb.User),
		// UnackMsgs:         make(map[string]map[string]bool),
		Consumers:         make(map[string]*Consumer),
		ConsumerSessions:  make(map[string]string),
		ConsumerUnackMsgs: make(map[string]map[string]bool),
		// config:            config,
	}
	vh.CreateExchange(default_exchange, DIRECT)
	return vh
}

func (vh *VHost) getSessionID(conn net.Conn) (string, bool) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	for sessionID, consumerID := range vh.ConsumerSessions {
		if conn.RemoteAddr().String() == consumerID {
			return sessionID, true
		}
	}
	return "", false
}

func (vh *VHost) CleanupConnection(conn net.Conn) {
	log.Println("Cleaning vhost connection")
	vh.mu.Lock()
	defer vh.mu.Unlock()

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
