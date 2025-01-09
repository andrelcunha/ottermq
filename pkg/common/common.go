package common

import (
	"net"
	"time"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
)

type FiberMap map[string]interface{}

type ConnectionInfo struct {
	Name              string                        `json:"name"`
	User              string                        `json:"user"`
	VHostName         string                        `json:"vhost"`
	VHostId           string                        `json:"vhost_id"`
	HeartbeatInterval uint16                        `json:"heartbeat_interval"`
	LastHeartbeat     time.Time                     `json:"last_heartbeat"`
	ConnectedAt       time.Time                     `json:"connected_at"`
	Conn              net.Conn                      `json:"-"`
	Channels          map[uint16]*amqp.ChannelState `json:"-"`
	Done              chan struct{}                 `json:"-"`
}
