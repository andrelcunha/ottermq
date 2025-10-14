package vhost

import (
	"sync"

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
	queueBufferSize    int                        `json:"-"`
	persist            persistence.Persistence
	Consumers          map[ConsumerKey]*Consumer            `json:"consumers"`          // <- Primary registry
	ConsumersByQueue   map[string][]*Consumer               `json:"consumers_by_queue"` // <- Delivery Index
	ConsumersByChannel map[ConnectionChannelKey][]*Consumer // Index consumers by connection+channel
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
		ConsumersByChannel: make(map[ConnectionChannelKey][]*Consumer),
	}
	vh.createMandatoryStructure()
	return vh
}

func (vh *VHost) createMandatoryStructure() {
	vh.createMandatoryExchanges()
}
