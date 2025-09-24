package models

import (
	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

type ConnectionInfo struct {
	Client   *amqp.AmqpClient
	Channels map[uint16]*amqp.ChannelState `json:"-"`
}
