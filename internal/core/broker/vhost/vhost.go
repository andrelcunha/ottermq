package vhost

import (
	"fmt"
	"net"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/persistdb"
	"github.com/andrelcunha/ottermq/internal/core/persistdb/persistence"
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
	MsgCtrlr          MessageController
	QueueBufferSize   int `json:"-"`
	persist           persistence.Persistence
}

type Consumer struct {
	ID        string `json:"id"`
	Queue     string `json:"queue"`
	SessionID string `json:"session_id"`
}

func NewVhost(vhostName string, queueBufferSize int, persist persistence.Persistence) *VHost {
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
		QueueBufferSize:   queueBufferSize,
		persist:           persist,
	}
	vh.MsgCtrlr = &DefaultMessageController{vh}
	vh.createMandatoryStructure()
	// Admin expecific
	// vh.CreateQueue(ADMIN_QUEUES)
	// vh.CreateQueue(ADMIN_EXCHANGES)
	// vh.CreateQueue(ADMIN_BINDINGS)
	// vh.CreateQueue(ADMIN_CONNECTIONS)
	// vh.bindAdminQueues()
	return vh
}

func (vh *VHost) createMandatoryStructure() {
	vh.createMandatoryExchanges()
}

func (vh *VHost) CleanupConnection(conn net.Conn) error {
	log.Debug().Msg("Cleaning vhost connection")
	vh.mu.Lock()
	defer vh.mu.Unlock()

	if sessionId, ok := vh.getSessionID(conn); ok {
		vh.handleConsumerDisconnection(sessionId)
		return nil
	}
	return fmt.Errorf("session not found")
	// publishConnectionUpdate(nil) // Trigger update afte cleanup
}

func (vh *VHost) handleConsumerDisconnection(sessionID string) {
	consumerID, ok := vh.ConsumerSessions[sessionID]
	if !ok {
		log.Debug().Str("session_id", sessionID).Msg("Session not found")
		return
	}

	consumer, ok := vh.Consumers[consumerID]
	if !ok {
		log.Debug().Str("consumer_id", consumerID).Msg("Consumer not found")
		return
	}

	for _, msg := range vh.ConsumerUnackMsgs[consumerID] {
		if queue, ok := vh.Queues[consumer.Queue]; ok {
			queue.ReQueue(msg)
		}
	}

	delete(vh.ConsumerUnackMsgs, consumerID)
	delete(vh.ConsumerSessions, sessionID)
	delete(vh.Consumers, consumerID)
	log.Debug().Str("consumer_id", consumerID).Str("session_id", sessionID).Msg("Cleaned up consumer")
}

func (vh *VHost) getSessionID(conn net.Conn) (string, bool) {
	// vh.mu.Lock()
	// defer vh.mu.Unlock()
	for sessionID, consumerID := range vh.ConsumerSessions {
		if conn.RemoteAddr().String() == consumerID {
			return sessionID, true
		}
	}
	return "", false
}
