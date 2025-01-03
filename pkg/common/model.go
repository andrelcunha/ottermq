package common

import "time"

type ConnectionInfoDTO struct {
	VHostName     string    `json:"vhost"`
	VHostId       string    `json:"vhost_id"`
	Name          string    `json:"name"`
	Username      string    `json:"user_name"`
	State         string    `json:"state"`
	SSL           bool      `json:"ssl"`
	Protocol      string    `json:"protocol"`
	Channels      int       `json:"channels"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	ConnectedAt   time.Time `json:"connected_at"`

	Done chan struct{} `json:"-"`
}

type ExchangeDTO struct {
	VHostName string `json:"vhost"`
	VHostId   string `json:"vhost_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}
