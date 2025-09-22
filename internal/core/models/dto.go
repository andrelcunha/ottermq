package models

import (
	"time"
)

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

func MapListConnectionsDTO(connections []ConnectionInfo) []ConnectionInfoDTO {
	listConnectonsDTO := make([]ConnectionInfoDTO, len(connections))
	for i, connection := range connections {
		state := "disconnected"
		if connection.Done == nil {
			state = "running"
		}
		channels := len(connection.Channels)
		listConnectonsDTO[i] = ConnectionInfoDTO{
			VHostName:     connection.VHostName,
			VHostId:       connection.VHostId,
			Name:          connection.Name,
			Username:      connection.User,
			State:         state,
			SSL:           false,
			Protocol:      "AMQP 0-9-1",
			Channels:      channels,
			LastHeartbeat: connection.LastHeartbeat,
			ConnectedAt:   connection.ConnectedAt,
		}
	}
	return listConnectonsDTO
}

type ExchangeDTO struct {
	VHostName string `json:"vhost"`
	VHostId   string `json:"vhost_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}

type QueueDTO struct {
	VHostName string `json:"vhost"`
	VHostId   string `json:"vhost_id"`
	Name      string `json:"name"`
	Messages  int    `json:"messages"`
}

type BindingDTO struct {
	VHostName string              `json:"vhost"`
	VHostId   string              `json:"vhost_id"`
	Exchange  string              `json:"exchange"`
	Bindings  map[string][]string `json:"bidings"`
}
