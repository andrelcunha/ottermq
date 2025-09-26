package vhost

import (
	"fmt"
)

const DEFAULT_EXCHANGE = "AMQP_default"

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
)

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
	if name == DEFAULT_EXCHANGE {
		return fmt.Errorf("cannot delete default exchange")
	}

	// Check if the exchange exists
	_, ok := vh.Exchanges[name]
	if !ok {
		return fmt.Errorf("exchange %s not found", name)
	}
	delete(vh.Exchanges, name)
	return nil
}
