package models

import (
	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/andrelcunha/ottermq/internal/core/amqp/shared"
)

type ConnectionInfo struct {
	Client   *shared.AmqpClient
	Channels map[uint16]*amqp.ChannelState `json:"-"`
}
