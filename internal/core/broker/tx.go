package broker

import (
	"fmt"

	"github.com/andrelcunha/ottermq/internal/core/amqp"
)

// txHandler handles transaction-related commands
func (*Broker) txHandler(request *amqp.RequestMethodMessage) (any, error) {
	switch request.MethodID {
	case uint16(amqp.SELECT):
		// Handle transaction selection
		return nil, fmt.Errorf("not implemented")
	case uint16(amqp.COMMIT):
		// Handle transaction commit
		return nil, fmt.Errorf("not implemented")
	case uint16(amqp.ROLLBACK):
		// Handle transaction rollback
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("unsupported command")
	}
}
