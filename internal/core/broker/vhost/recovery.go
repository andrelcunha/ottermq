package vhost

import (
	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/persistdb/persistence"
)

// RecoverExchange recreates an exchange from its persisted state
func (vh *VHost) RecoverExchange(persistedEx *persistence.PersistedExchange) {
	ex := &Exchange{
		Name:     persistedEx.Name,
		Typ:      ExchangeType(persistedEx.Type),
		Queues:   make(map[string]*Queue),
		Bindings: make(map[string][]*Queue),
		Props: &ExchangeProperties{
			Durable:    persistedEx.Properties.Durable,
			AutoDelete: persistedEx.Properties.AutoDelete,
			Internal:   persistedEx.Properties.Internal,
			NoWait:     persistedEx.Properties.NoWait,
			Arguments:  persistedEx.Properties.Arguments,
		},
	}
	vh.Exchanges[ex.Name] = ex
	// TODO Recreate bindings
}

// RecoverQueue recreates a queue from its persisted state
func (vh *VHost) RecoverQueue(persistedQ *persistence.PersistedQueue) {
	q := &Queue{
		Name: persistedQ.Name,
		Props: &QueueProperties{
			Passive:    persistedQ.Properties.Passive,
			Durable:    persistedQ.Properties.Durable,
			AutoDelete: persistedQ.Properties.AutoDelete,
			Exclusive:  persistedQ.Properties.Exclusive,
			NoWait:     persistedQ.Properties.NoWait,
			Arguments:  persistedQ.Properties.Arguments,
		},
	}
	vh.Queues[q.Name] = q
	// Recreate messages
	for _, msgData := range persistedQ.Messages {
		msg := &amqp.Message{
			ID:   msgData.ID,
			Body: msgData.Body,
		}
		q.messages <- *msg
		q.count++
	}
}
