package broker

import (
	"fmt"
	"net"

	"github.com/andrelcunha/ottermq/pkg/common/communication/amqp"
)

func (b *Broker) checkChannel(conn net.Conn, channel uint16) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.Connections[conn].Channels[channel]
	return ok
}

// Add new Channel
func (b *Broker) addChannel(conn net.Conn, frame *amqp.RequestMethodMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// add new channel to the connectionInfo
	b.Connections[conn].Channels[frame.Channel] = &amqp.ChannelState{MethodFrame: frame}
	fmt.Printf("[DEBUG] New channel added: %d\n", frame.Channel)
}

func (b *Broker) removeChannel(conn net.Conn, channel uint16) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.Connections[conn].Channels, channel)
}
