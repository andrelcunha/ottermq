package broker

import (
	"fmt"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
	"github.com/rs/zerolog/log"
)

func (b *Broker) channelHandler(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	switch request.MethodID {
	case uint16(amqp.CHANNEL_OPEN):
		return b.openChannel(request, conn)
	case uint16(amqp.CHANNEL_CLOSE):
		return b.closeChannel(request, conn)
	default:
		log.Debug().Uint16("method_id", request.MethodID).Msg("Unknown channel method")
		return nil, fmt.Errorf("unknown channel method: %d", request.MethodID)
	}
}

// openChannel executes the AMQP command CHANNEL_OPEN
func (b *Broker) openChannel(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	log.Debug().Interface("request", request).Msg("Received channel open request")

	// Check if the channel is already open
	if b.checkChannel(conn, request.Channel) {
		log.Debug().Uint16("channel", request.Channel).Msg("Channel already open")
		return nil, fmt.Errorf("channel already open")
	}
	b.registerChannel(conn, request)
	log.Trace().Interface("state", b.Connections[conn].Channels[request.Channel]).Msg("New state added")

	frame := b.framer.CreateChannelOpenOkFrame(request)
	b.framer.SendFrame(conn, frame)
	return nil, nil
}

func (b *Broker) closeChannel(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	if !b.checkChannel(conn, request.Channel) {
		log.Debug().Uint16("channel", request.Channel).Msg("Channel already closed") // no need to rise an error here
		return nil, nil
	}
	b.removeChannel(conn, request.Channel)
	frame := b.framer.CreateChannelCloseOkFrame(request.Channel)
	err := b.framer.SendFrame(conn, frame)
	return nil, err
}

// registerChannel register a new channel to the connection
func (b *Broker) registerChannel(conn net.Conn, frame *amqp.RequestMethodMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Connections[conn].Channels[frame.Channel] = &amqp.ChannelState{MethodFrame: frame}
	log.Debug().Uint16("channel", frame.Channel).Msg("New channel added")
}

// removeChannel removes a channel from the connection
func (b *Broker) removeChannel(conn net.Conn, channel uint16) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Connections[conn].Channels, channel)
}

// checkChannel checks if a channel is already open
func (b *Broker) checkChannel(conn net.Conn, channel uint16) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.Connections[conn].Channels[channel]
	return ok
}
