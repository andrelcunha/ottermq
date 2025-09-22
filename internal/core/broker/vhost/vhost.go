package vhost

import (
	"log"
	"net"
	"sync"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/persistdb"
	"github.com/google/uuid"
)

type VHost struct {
	Name              string                             `json:"name"`
	Id                string                             `json:"id"`
	Exchanges         map[string]*Exchange               `json:"exchanges"`
	Queues            map[string]*Queue                  `json:"queues"`
	Users             map[string]*persistdb.User         `json:"users"`
	Consumers         map[string]*Consumer               `json:"consumers"`
	ConsumerSessions  map[string]string                  `json:"consumer_sessions"`
	ConsumerUnackMsgs map[string]map[string]amqp.Message `json:"consumer_unacked_messages"`
	mu                sync.Mutex                         `json:"-"`
}

type Consumer struct {
	ID        string `json:"id"`
	Queue     string `json:"queue"`
	SessionID string `json:"session_id"`
}

func NewVhost(vhostName string) *VHost {
	id := uuid.New().String()
	vh := &VHost{
		Name:              vhostName,
		Id:                id,
		Exchanges:         make(map[string]*Exchange),
		Queues:            make(map[string]*Queue),
		Users:             make(map[string]*persistdb.User),
		Consumers:         make(map[string]*Consumer),
		ConsumerSessions:  make(map[string]string),
		ConsumerUnackMsgs: make(map[string]map[string]amqp.Message),
	}
	vh.CreateExchange(default_exchange, DIRECT)
	return vh
}

func (vh *VHost) CleanupConnection(conn net.Conn) {
	log.Println(" [DEBUG] Cleaning vhost connection")
	vh.mu.Lock()
	defer vh.mu.Unlock()

	sessionId, ok := vh.getSessionID(conn)
	if ok {
		vh.handleConsumerDisconnection(sessionId)
	}
}

func (vh *VHost) handleConsumerDisconnection(sessionID string) {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	consumerID, ok := vh.ConsumerSessions[sessionID]
	if !ok {
		log.Printf("[DEBUG] Session %s not found\n", sessionID)
		return
	}

	consumer, ok := vh.Consumers[consumerID]
	if !ok {
		log.Printf("[DEBUG] Consumer %s not found\n", consumerID)
		return
	}

	// requeue unacknowledged messages
	for _, msg := range vh.ConsumerUnackMsgs[consumerID] {
		if queue, ok := vh.Queues[consumer.Queue]; ok {
			queue.ReQueue(msg)
		}
	}

	delete(vh.ConsumerUnackMsgs, consumerID)
	delete(vh.ConsumerSessions, sessionID)
	delete(vh.Consumers, consumerID)
	log.Printf("[DEBUG] Cleaned up consumer %s, session %s", consumerID, sessionID)
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
