package broker

import (
	"fmt"
	"log"
	"net"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

func (b *Broker) channelHandler(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	switch request.MethodID {
	case uint16(amqp.CHANNEL_OPEN):
		return b.openChannel(request, conn)
	case uint16(amqp.CHANNEL_CLOSE):
		return b.closeChannel(request, conn)
	default:
		log.Printf("[DEBUG] Unknown channel method: %d", request.MethodID)
		return nil, fmt.Errorf("unknown channel method: %d", request.MethodID)
	}
}

// openChannel executes the AMQP command CHANNEL_OPEN
func (b *Broker) openChannel(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	log.Printf("[DEBUG] Received channel open request: %+v\n", request)

	// Check if the channel is already open
	if b.checkChannel(conn, request.Channel) {
		log.Printf("[DEBUG] Channel %d already open\n", request.Channel)
		return nil, fmt.Errorf("channel already open")
	}
	b.registerChannel(conn, request)
	log.Printf("[TRACE] New state added: %+v\n", b.Connections[conn].Channels[request.Channel])

	frame := b.framer.CreateChannelOpenOkFrame(request)
	b.framer.SendFrame(conn, frame)
	return nil, nil
}

func (b *Broker) closeChannel(request *amqp.RequestMethodMessage, conn net.Conn) (any, error) {
	if b.checkChannel(conn, request.Channel) {
		fmt.Printf("[DEBUG] Channel %d already open\n", request.Channel)
		return nil, fmt.Errorf("channel already open")
	}
	b.removeChannel(conn, request.Channel)
	frame := b.framer.CreateChannelCloseFrame(request)
	b.framer.SendFrame(conn, frame)
	return nil, nil
}

// registerChannel register a new channel to the connection
func (b *Broker) registerChannel(conn net.Conn, frame *amqp.RequestMethodMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Connections[conn].Channels[frame.Channel] = &amqp.ChannelState{MethodFrame: frame}
	log.Printf("[DEBUG] New channel added: %d\n", frame.Channel)
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
