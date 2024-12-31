package common

import "time"

type ConnectionInfoDTO struct {
	VHost         string    `json:"vhost"`
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
	Vhost string `json:"vhost"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}
