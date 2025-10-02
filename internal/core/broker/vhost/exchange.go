package vhost

import (
	"fmt"
)

type Exchange struct {
	Name     string              `json:"name"`
	Queues   map[string]*Queue   `json:"queues"`
	Typ      ExchangeType        `json:"type"`
	Bindings map[string][]*Queue `json:"bindings"`
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
		vh.CreateExchange(mandatoryExchange.Name, mandatoryExchange.Type)
	}
	vh.mu.Lock()
	defer vh.mu.Unlock()
	if defaultExchange, exists := vh.Exchanges[DEFAULT_EXCHANGE]; exists {
		vh.Exchanges[EMPTY_EXCHANGE] = defaultExchange
	}
}

func (vh *VHost) CreateExchange(name string, typ ExchangeType) error {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	// Check if the exchange already exists
	if _, ok := vh.Exchanges[name]; ok {
		return fmt.Errorf("exchange %s already exists", name)
	}

	exchange := &Exchange{
		Name:     name,
		Typ:      typ,
		Queues:   make(map[string]*Queue),
		Bindings: make(map[string][]*Queue),
	}
	vh.Exchanges[name] = exchange
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
