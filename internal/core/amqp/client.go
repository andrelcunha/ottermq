package amqp

import (
	"log"
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
	interlaval        time.Duration
	maxDelay          time.Duration
	heartbeatChan     chan struct{}
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
		interlaval:        time.Duration(config.HeartbeatInterval>>1) * time.Second,
		maxDelay:          time.Duration(config.HeartbeatInterval<<1) * time.Second,
		heartbeatChan:     make(chan struct{}, 1),
	}
	client.startHeartbeat()
	client.startHeartbeatMonitor()
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

func (c *AmqpClient) startHeartbeat() {
	ticker := time.NewTicker(c.interlaval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				heartbeatFrame := createHeartbeatFrame()
				err := sendFrame(c.Conn, heartbeatFrame)
				if err != nil {
					log.Printf("[ERROR] Failed to send heartbeat: %v", err)
					return
				}
				log.Println("[DEBUG] Heartbeat sent")
			case <-c.Done:
				log.Println("[DEBUG] Heartbeat stopped for", c.Name)
				return
			}
		}
	}()
}

func (c *AmqpClient) startHeartbeatMonitor() {
	timer := time.NewTimer(c.maxDelay)

	go func() {
		for {
			select {
			case <-timer.C:
				log.Printf("[DEBUG] Heartbeat overdue for %s", c.Name)
				c.Stop()
				return

			case <-c.heartbeatChan:
				c.LastHeartbeat = time.Now()
				if !timer.Stop() {
					<-timer.C // drain if needed
				}
				timer.Reset(c.maxDelay)

			case <-c.Done:
				log.Printf("[DEBUG] Heartbeat monitor stopped for %s", c.Name)
				timer.Stop()
				return
			}
		}
	}()
}

func (c *AmqpClient) ReceiveHeartbeat() {
	select {
	case c.heartbeatChan <- struct{}{}:
	default: // avoid blocking if no one is listening
	}
}

func (c *AmqpClient) IsAlive() bool {
	select {
	case <-c.Done:
		return false
	default:
		return time.Since(c.LastHeartbeat) <= c.maxDelay

	}
}

func (c *AmqpClient) Stop() {
	close(c.Done)
}
