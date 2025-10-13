package vhost

import (
	"net"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/andrelcunha/ottermq/internal/persistdb"
	"github.com/andrelcunha/ottermq/pkg/persistence"
	"github.com/google/uuid"
)

type VHost struct {
	Name               string                     `json:"name"`
	Id                 string                     `json:"id"`
	Exchanges          map[string]*Exchange       `json:"exchanges"`
	Queues             map[string]*Queue          `json:"queues"`
	Users              map[string]*persistdb.User `json:"users"`
	mu                 sync.Mutex                 `json:"-"`
	MsgCtrlr           MessageController
	queueBufferSize    int `json:"-"`
	persist            persistence.Persistence
	Consumers          map[ConsumerKey]*Consumer `json:"consumers"`            // <- Primary registry
	ConsumersByQueue   map[string][]*Consumer    `json:"consumers_by_queue"`   // <- Delivery Index
	ConsumersByChannel map[uint16][]*Consumer    `json:"consumers_by_channel"` // <- Channel Index
}

func NewVhost(vhostName string, queueBufferSize int, persist persistence.Persistence) *VHost {
	id := uuid.New().String()
	vh := &VHost{
		Name:               vhostName,
		Id:                 id,
		Exchanges:          make(map[string]*Exchange),
		Queues:             make(map[string]*Queue),
		Users:              make(map[string]*persistdb.User),
		queueBufferSize:    queueBufferSize,
		persist:            persist,
		Consumers:          make(map[ConsumerKey]*Consumer),
		ConsumersByQueue:   make(map[string][]*Consumer),
		ConsumersByChannel: make(map[uint16][]*Consumer),
	}
	vh.MsgCtrlr = &DefaultMessageController{vh}
	vh.createMandatoryStructure()
	return vh
}

func (vh *VHost) createMandatoryStructure() {
	vh.createMandatoryExchanges()
}

func (vh *VHost) CleanupConnection(conn net.Conn) {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	// Find and cleanup all consumers for this connection
	var channelsToCleanup []uint16
	for key, consumer := range vh.Consumers {
		if consumer.Connection == conn {
			channelsToCleanup = append(channelsToCleanup, key.Channel)
		}
	}

	// Cleanup channels (release lock temporarily to avoid deadlock)
	vh.mu.Unlock()
	for _, channel := range channelsToCleanup {
		vh.CleanupChannel(channel)
	}
	vh.mu.Lock()

	log.Debug().Msg("Cleaned vhost connection")
}
