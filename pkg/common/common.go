package common

import (
	"net"
	"time"
)

type FiberMap map[string]interface{}

type ConnectionInfo struct {
	User              string    `json:"user"`
	VHost             string    `json:"vhost"`
	HeartbeatInterval uint16    `json:"heartbeat_interval"`
	LastHeartbeat     time.Time `json:"last_heartbeat"`
	ConnectedAt       time.Time `json:"connected_at"`
	Conn              net.Conn  `json:"-"`
	Done              chan struct{}
}
