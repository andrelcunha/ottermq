package vhost

import (
	"context"
	"net"
	"sync"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
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
	activeDeliveries   map[string]context.CancelFunc        // queueName -> cancelFunc
	framer             amqp.Framer
	ChannelDeliveries  map[ConnectionChannelKey]*ChannelDeliveryState
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
		activeDeliveries:   make(map[string]context.CancelFunc),
		ChannelDeliveries:  make(map[ConnectionChannelKey]*ChannelDeliveryState),
	}
	vh.createMandatoryStructure()
	return vh
}

func (vh *VHost) SetFramer(framer amqp.Framer) {
	vh.framer = framer
}

func (vh *VHost) createMandatoryStructure() {
	vh.createMandatoryExchanges()
}

// GetUnackedMessageCountsAllQueues returns a map of queue names to their respective counts of unacknowledged messages.
func (vh *VHost) GetUnackedMessageCountsAllQueues() map[string]int {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	counts := make(map[string]int)
	for _, channelState := range vh.ChannelDeliveries {
		channelState.mu.Lock()
		for _, record := range channelState.Unacked {
			counts[record.QueueName]++
		}
		channelState.mu.Unlock()
	}
	return counts
}

// GetUnackedMessageCountByChannel returns the count of unacknowledged messages for a specific connection and channel.
func (vh *VHost) GetUnackedMessageCountByChannel(conn net.Conn, channel uint16) int {
	vh.mu.Lock()
	defer vh.mu.Unlock()

	key := ConnectionChannelKey{conn, channel}
	channelState := vh.ChannelDeliveries[key]
	if channelState == nil {
		return 0
	}

	channelState.mu.Lock()
	count := len(channelState.Unacked)
	channelState.mu.Unlock()

	return count
}
