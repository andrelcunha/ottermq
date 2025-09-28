package amqp

import (
	"context"
	"net"
	"time"
)

type AmqpClient struct {
	RemoteAddr    string
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	Conn          net.Conn
	Ctx           context.Context
	Cancel        context.CancelFunc
	Config        *AmqpClientConfig
}

type AmqpClientConfig struct {
	Username          string
	HeartbeatInterval uint16
	FrameMax          uint32
	ChannelMax        uint16
	Protocol          string
	SSL               bool
}

func NewAmqpClient(conn net.Conn, config *AmqpClientConfig, connCtx context.Context, cancel context.CancelFunc) *AmqpClient {

	client := &AmqpClient{
		RemoteAddr:  conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
		Conn:        conn,
		Ctx:         connCtx,
		Cancel:      cancel,
		Config:      config,
	}

	return client
}

func NewAmqpClientConfig(configurations *map[string]any) *AmqpClientConfig {
	username := (*configurations)["username"].(string)
	heartbeatInterval := (*configurations)["heartbeatInterval"].(uint16)
	frameMax := (*configurations)["frameMax"].(uint32)
	channelMax := (*configurations)["channelMax"].(uint16)
	protocol := (*configurations)["protocol"].(string)
	ssl := (*configurations)["ssl"].(bool)

	return &AmqpClientConfig{
		Username:          username,
		HeartbeatInterval: heartbeatInterval,
		FrameMax:          frameMax,
		ChannelMax:        channelMax,
		Protocol:          protocol,
		SSL:               ssl,
	}
}

// ConnectionInfo represents the information of a connection to the AMQP server
type ConnectionInfo struct {
	VHostName string                   `json:"vhost"`
	Client    *AmqpClient              `json:"client"`
	Channels  map[uint16]*ChannelState `json:"channels"`
}

// NewConnectionInfo creates a new ConnectionInfo, receiving the `vhost` name
func NewConnectionInfo(vhostName string) *ConnectionInfo {
	return &ConnectionInfo{
		VHostName: vhostName,
		Client:    nil,
		Channels:  make(map[uint16]*ChannelState),
	}
}
