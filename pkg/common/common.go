package common

import "time"

type CommandResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type FiberMap map[string]interface{}

type ConnectionInfo struct {
	Name          string    `json:"name"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	ConnectedAt   time.Time `json:"connected_at"`
}
