package vhost

import (
	"fmt"

	db "github.com/andrelcunha/ottermq/internal/core/persistdb/persistence"
)

type Exchange struct {
	Name     string              `json:"name"`
	Queues   map[string]*Queue   `json:"queues"`
	Typ      ExchangeType        `json:"type"`
	Bindings map[string][]*Queue `json:"bindings"`
	Props    *ExchangeProperties `json:"properties"`
}

type ExchangeProperties struct {
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

func (vh *VHost) CreateExchange(name string, typ ExchangeType, props *ExchangeProperties) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	// Check if the exchange already exists
	if _, ok := vh.Exchanges[name]; ok {
		return fmt.Errorf("exchange %s already exists", name)
	}

	if props == nil {
		props = &ExchangeProperties{
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
		vh.persist.SaveExchange(vh.Name, exchange.ToPersistence())
	}
	return nil
}

func (vh *VHost) DeleteExchange(name string) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
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
	return nil
}

func (e *Exchange) ToPersistence() *db.PersistedExchange {
	bindings := make([]db.PersistedBinding, 0)
	for routingKey, queues := range e.Bindings {
		for _, queue := range queues {
			bindings = append(bindings, db.PersistedBinding{
				QueueName:  queue.Name,
				RoutingKey: routingKey,
				Arguments:  nil, // TODO: Add support for binding arguments
			})
		}
	}
	return &db.PersistedExchange{
		Name:       e.Name,
		Type:       string(e.Typ),
		Properties: e.Props.ToPersistence(),
		Bindings:   bindings,
	}
}
func (ep *ExchangeProperties) ToPersistence() db.ExchangePropertiesDb {
	return db.ExchangePropertiesDb{
		Durable:    ep.Durable,
		AutoDelete: ep.AutoDelete,
		Internal:   ep.Internal,
		NoWait:     ep.NoWait,
		Arguments:  ep.Arguments,
	}
}
