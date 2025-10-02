package models

import (
	"time"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
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

func MapListConnectionsDTO(connections []amqp.ConnectionInfo) []ConnectionInfoDTO {
	listConnectonsDTO := make([]ConnectionInfoDTO, len(connections))
	for i, connection := range connections {
		state := "disconnected"
		if connection.Client.Ctx.Err() == nil {
			state = "running"
		}
		channels := len(connection.Channels)
		listConnectonsDTO[i] = ConnectionInfoDTO{
			VHostName:     connection.VHostName,
			Name:          connection.Client.RemoteAddr,
			Username:      connection.Client.Config.Username,
			State:         state,
			SSL:           connection.Client.Config.SSL,
			Protocol:      connection.Client.Config.Protocol,
			Channels:      channels,
			LastHeartbeat: connection.Client.LastHeartbeat,
			ConnectedAt:   connection.Client.ConnectedAt,
		}
	}
	return listConnectonsDTO
}

type ExchangeDTO struct {
	VHostName string `json:"vhost"`
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
	Bindings  map[string][]string `json:"bindings"`
}
