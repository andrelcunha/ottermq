package common

import (
	"net"
	"time"
)

type FiberMap map[string]interface{}

type ConnectionInfo struct {
	Name              string        `json:"name"`
	User              string        `json:"user"`
	VHostName         string        `json:"vhost"`
	VHostId           string        `json:"vhost_id"`
	HeartbeatInterval uint16        `json:"heartbeat_interval"`
	LastHeartbeat     time.Time     `json:"last_heartbeat"`
	ConnectedAt       time.Time     `json:"connected_at"`
	Conn              net.Conn      `json:"-"`
	Channels          map[int]bool  `json:"-"`
	Done              chan struct{} `json:"-"`
}

type Channel struct {
	ID uint16 `json:"id"`
	// ConsumerTag string `json:"consumer_tag"`
	// Active      bool   `json:"active"`
	// Exclusive   bool   `json:"exclusive"`
	// AutoDelete  bool   `json:"auto_delete"`
	// NoWait      bool   `json:"no_wait"`
	// Arguments   map[string]interface{} `json:"arguments"`
}
