package vhost

import (
	"sync"

	"github.com/andrelcunha/ottermq/pkg/persistdb"
)

const default_exchange = "_OTTERMQ_DEFAULT_EXCHANGE_"

type VHost struct {
	Name      string                     `json:"name"`
	Exchanges map[string]*Exchange       `json:"exchanges"`
	Queues    map[string]*Queue          `json:"queues"`
	Users     map[string]*persistdb.User `json:"users"`

	// UnackMsgs         map[string]map[string]bool `json:"unacked_messages"`
	Consumers         map[string]*Consumer       `json:"consumers"`
	ConsumerSessions  map[string]string          `json:"consumer_sessions"`
	ConsumerUnackMsgs map[string]map[string]bool `json:"consumer_unacked_messages"`
	mu                sync.Mutex                 `json:"-"`
}

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

type Consumer struct {
	ID        string `json:"id"`
	Queue     string `json:"queue"`
	SessionID string `json:"session_id"`
}

func NewVhost(vhostName string) *VHost {
	b := &VHost{
		Name:      vhostName,
		Exchanges: make(map[string]*Exchange),
		Queues:    make(map[string]*Queue),
		Users:     make(map[string]*persistdb.User),
		// UnackMsgs:         make(map[string]map[string]bool),
		Consumers:         make(map[string]*Consumer),
		ConsumerSessions:  make(map[string]string),
		ConsumerUnackMsgs: make(map[string]map[string]bool),
		// config:            config,
	}
	b.createExchange(default_exchange, DIRECT)
	return b
}
