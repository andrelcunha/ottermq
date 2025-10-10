package vhost

import (
	"fmt"

	"github.com/andrelcunha/ottermq/pkg/persistence"
	"github.com/rs/zerolog/log"
)

type Exchange struct {
	Name     string              `json:"name"`
	Queues   map[string]*Queue   `json:"queues"` // Queues bound directly to this exchange (e.g., for fanout)
	Typ      ExchangeType        `json:"type"`
	Bindings map[string][]*Queue `json:"bindings"` // RoutingKey to Queues bindings
	Props    *ExchangeProperties `json:"properties"`
}

type ExchangeProperties struct {
	Passive    bool           `json:"passive"`
	Durable    bool           `json:"durable"`
	AutoDelete bool           `json:"auto_delete"`
	Internal   bool           `json:"internal"`
	NoWait     bool           `json:"no_wait"`
	Arguments  map[string]any `json:"arguments"`
}

type ExchangeType string

const (
	DIRECT ExchangeType = "direct"
	FANOUT ExchangeType = "fanout"
	TOPIC  ExchangeType = "topic"
)

type MandatoryExchange struct {
	Name string       `json:"name"`
	Type ExchangeType `json:"type"`
}

const (
	DEFAULT_EXCHANGE = "amq.default"
	EMPTY_EXCHANGE   = ""
	MANDATORY_TOPIC  = "amq.topic"
	MANDATORY_DIRECT = "amq.direct"
	MANDATORY_FANOUT = "amq.fanout"
)

var mandatoryExchanges = []MandatoryExchange{
	{Name: DEFAULT_EXCHANGE, Type: DIRECT},
	{Name: MANDATORY_TOPIC, Type: DIRECT},
	{Name: MANDATORY_DIRECT, Type: DIRECT},
	{Name: MANDATORY_FANOUT, Type: FANOUT},
}

// Candidate to be on an ExchangeManager interface
func ParseExchangeType(s string) (ExchangeType, error) {
	switch s {
	case string(DIRECT):
		return DIRECT, nil
	case string(FANOUT):
		return FANOUT, nil
	default:
		return "", fmt.Errorf("invalid exchange type: %s", s)
	}
}

func (vh *VHost) createMandatoryExchanges() {
	for _, mandatoryExchange := range mandatoryExchanges {
		vh.CreateExchange(mandatoryExchange.Name, mandatoryExchange.Type, &ExchangeProperties{
			Durable:    false,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		})
	}
	vh.mu.Lock()
	defer vh.mu.Unlock()
	if defaultExchange, exists := vh.Exchanges[DEFAULT_EXCHANGE]; exists {
		vh.Exchanges[EMPTY_EXCHANGE] = defaultExchange
	}
}

// CreateExchange creates a new exchange with the given name, type, and properties.
func (vh *VHost) CreateExchange(name string, typ ExchangeType, props *ExchangeProperties) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	// Check if the exchange already exists
	if _, ok := vh.Exchanges[name]; ok {
		return fmt.Errorf("exchange %s already exists", name)
	}
	// Deal with passive property
	if props != nil && props.Passive {
		if _, ok := vh.Exchanges[name]; !ok {
			return fmt.Errorf("exchange %s does not exist", name)
		}
	}

	if props == nil {
		props = &ExchangeProperties{
			Passive:    false,
			Durable:    false,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		}
	}
	exchange := &Exchange{
		Name:     name,
		Typ:      typ,
		Queues:   make(map[string]*Queue),
		Bindings: make(map[string][]*Queue),
		Props:    props,
	}
	vh.Exchanges[name] = exchange
	// Handle durable property
	if props.Durable {
		vh.persist.SaveExchangeMetadata(vh.Name, name, string(typ), props.ToPersistence())
	}
	return nil
}

// Internal helper: assumes vh.mu is already locked
func (vh *VHost) deleteExchangeUnlocked(name string) error {
	// If the exchange is the default exchange, return an error
	for _, mandatoryExchange := range mandatoryExchanges {
		if name == mandatoryExchange.Name {
			return fmt.Errorf("cannot delete default exchange")
		}
	}
	// Check if the exchange exists
	_, ok := vh.Exchanges[name]
	if !ok {
		return fmt.Errorf("exchange %s not found", name)
	}

	delete(vh.Exchanges, name)
	// Handle durable property
	if err := vh.persist.DeleteExchangeMetadata(vh.Name, name); err != nil {
		return fmt.Errorf("failed to delete exchange from persistence: %v", err)
	}
	log.Debug().Str("exchange", name).Msg("Deleted exchange")
	return nil
}

func (vh *VHost) DeleteExchange(name string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	return vh.deleteExchangeUnlocked(name)
}

// checkAutoDeleteExchange checks if an exchange is auto-delete and has no bindings, and deletes it if so.
// It returns true if the exchange was deleted, false otherwise.
// Internal helper: assumes vh.mu is already locked
func (vh *VHost) checkAutoDeleteExchangeUnlocked(name string) (bool, error) {
	exchange, ok := vh.Exchanges[name]
	if !ok {
		return false, fmt.Errorf("exchange %s not found", name)
	}

	if exchange.Props.AutoDelete && len(exchange.Bindings) == 0 && len(exchange.Queues) == 0 {
		log.Debug().Str("exchange", name).Msg("Auto-deleting exchange")
		if err := vh.deleteExchangeUnlocked(name); err != nil {
			return false, fmt.Errorf("failed to auto-delete exchange %s: %v", name, err)
		}
		return true, nil
	}
	return false, nil
}

// CheckAutoDeleteExchange checks if an exchange is auto-delete and has no bindings, and deletes it if so.
func (vh *VHost) CheckAutoDeleteExchange(name string) (bool, error) {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	return vh.checkAutoDeleteExchangeUnlocked(name)
}

// ToPersistence convert Exchange to persistence format
// func (e *Exchange) ToPersistence() *persistence.PersistedExchange {
// 	bindings := make([]persistence.PersistedBinding, 0)
// 	for routingKey, queues := range e.Bindings {
// 		for _, queue := range queues {
// 			bindings = append(bindings, persistence.PersistedBinding{
// 				QueueName:  queue.Name,
// 				RoutingKey: routingKey,
// 				Arguments:  nil, // TODO: Add support for binding arguments
// 			})
// 		}
// 	}
// 	return &persistence.PersistedExchange{
// 		Name:       e.Name,
// 		Type:       string(e.Typ),
// 		Properties: e.Props.ToPersistence(),
// 		Bindings:   bindings,
// 	}
// }

// ToPersistence convert ExchangeProperties to persistence format
func (ep *ExchangeProperties) ToPersistence() persistence.ExchangeProperties {
	return persistence.ExchangeProperties{
		// Passive:    ep.Passive, // Not needed in persistence
		Durable:    ep.Durable,
		AutoDelete: ep.AutoDelete,
		Internal:   ep.Internal,
		// NoWait:     ep.NoWait, // Not needed in persistence
		Arguments: ep.Arguments,
	}
}
