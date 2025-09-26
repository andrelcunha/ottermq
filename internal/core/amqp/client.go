package amqp

import (
	"net"
	"time"
)

type AmqpClient struct {
	Name              string
	User              string
	VHostName         string
	VHostId           string
	ConnectedAt       time.Time
	LastHeartbeat     time.Time
	HeartbeatInterval uint16
	FrameMax          uint32
	ChannelMax        uint16
	Conn              net.Conn
	Protocol          string
	SSL               bool
	Done              chan struct{}
	// interlaval        time.Duration
	// maxDelay          time.Duration
	// heartbeatChan     chan struct{}
}

type AmqpClientConfig struct {
	Username          string
	Vhost             string
	HeartbeatInterval uint16
	FrameMax          uint32
	ChannelMax        uint16
	Protocol          string
	SSL               bool
}

func NewAmqpClient(conn net.Conn, config *AmqpClientConfig) *AmqpClient {

	client := &AmqpClient{
		Name:              conn.RemoteAddr().String(),
		User:              config.Username,
		VHostName:         config.Vhost,
		Protocol:          config.Protocol,
		SSL:               config.SSL,
		ConnectedAt:       time.Now(),
		HeartbeatInterval: config.HeartbeatInterval,
		LastHeartbeat:     time.Now(),
		FrameMax:          config.FrameMax,
		ChannelMax:        config.ChannelMax,
		Conn:              conn,
		Done:              make(chan struct{}),
		// interlaval:        time.Duration(config.HeartbeatInterval>>1) * time.Second,
		// maxDelay:          time.Duration(config.HeartbeatInterval<<1) * time.Second,
		// heartbeatChan:     make(chan struct{}, 1),
	}
	// client.startHeartbeat()
	// client.startHeartbeatMonitor()
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

func (c *AmqpClient) Stop() {
	close(c.Done)
}
