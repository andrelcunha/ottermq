package vhost

import (
	"encoding/json"
	"log"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/models"
)

func (vh *VHost) bindAdminQueues() {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	exchange, ok := vh.Exchanges[DEFAULT_EXCHANGE]
	if !ok {
		log.Printf("[ERROR] Default exchange not found")
		return
	}
	for _, queueName := range []string{
		ADMIN_QUEUES,
		ADMIN_EXCHANGES,
		ADMIN_BINDINGS,
		ADMIN_CONNECTIONS,
	} {
		exchange.Bindings[queueName] = append(exchange.Bindings[queueName], vh.Queues[queueName])
	}
	log.Printf("[DEBUG] Bound admin queues to default exchange")
}

// publishQueueUpdate sends queue state to admin queue
func (vh *VHost) publishQueueUpdate() {
	queues := make([]models.QueueDTO, 0)
	for _, q := range vh.Queues {
		queues = append(queues, models.QueueDTO{
			VHostName: vh.Name,
			VHostId:   vh.Id,
			Name:      q.Name,
			Messages:  q.Len(),
			// Durable: q.Durable,
			// Exclusive: q.Exclusive,
			// AutoDelete: q.AutoDelete,
			// MessageTTL: q.MessageTTL,
			// Arguments: q.Arguments,
		})
	}
	body, err := json.Marshal(queues)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal queue update: %v", err)
	}
	props := &amqp.BasicProperties{ContentType: "application/json"}
	_, err = vh.publish(DEFAULT_EXCHANGE, ADMIN_QUEUES, body, props)
	if err != nil {
		log.Printf("[ERROR] Failed to publish queue update: %v", err)
	}
}

// publishExchangeUpdate sends exchange state to admin queue
func (vh *VHost) publishExchangeUpdate() {
	exchanges := make([]models.ExchangeDTO, 0)
	for _, e := range vh.Exchanges {
		exchanges = append(exchanges, models.ExchangeDTO{
			VHostName: vh.Name,
			VHostId:   vh.Id,
			Name:      e.Name,
			Type:      string(e.Typ),
		})
	}
	body, err := json.Marshal(exchanges)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal excange update: %v", err)
		return
	}
	props := &amqp.BasicProperties{ContentType: "application/json"}
	_, err = vh.publish(DEFAULT_EXCHANGE, ADMIN_EXCHANGES, body, props)
	if err != nil {
		log.Printf("[ERROR] Failed to publish exchage update: %v", err)
	}
}

// publishBindingUpdate sends binding state to admin queue
func (vh *VHost) publishBindingUpdate(exchangeName string) {
	exchange, ok := vh.Exchanges[exchangeName]
	if !ok {
		return
	}
	bindings := make(map[string][]string)
	switch exchange.Typ {
	case DIRECT:
		for routingKey, queues := range exchange.Bindings {
			var queueNames []string
			for _, q := range queues {
				queueNames = append(queueNames, q.Name)
			}
			bindings[routingKey] = queueNames
		}
	case FANOUT:
		var queueNames []string
		for qNames := range exchange.Queues {
			queueNames = append(queueNames, qNames)
		}
		bindings["fanout"] = queueNames
	}
	body, err := json.Marshal(models.BindingDTO{
		VHostName: vh.Name,
		VHostId:   vh.Id,
		Exchange:  exchangeName,
		Bindings:  bindings,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to marshal binding update: %v", err)
		return
	}
	props := &amqp.BasicProperties{ContentType: "application/json"}
	_, err = vh.publish(DEFAULT_EXCHANGE, ADMIN_BINDINGS, body, props)
	if err != nil {
		log.Printf("[ERROR] Failed to publish binding update: %v", err)
	}
}
