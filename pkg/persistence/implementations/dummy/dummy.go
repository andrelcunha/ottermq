package dummy

import "github.com/andrelcunha/ottermq/pkg/persistence"

// DummyPersistence implements persistence.Persistence with no-ops for testing
type DummyPersistence struct{}

func (d *DummyPersistence) SaveExchangeMetadata(vhost, name, exchangeType string, props persistence.ExchangeProperties) error {
	return nil
}
func (d *DummyPersistence) LoadExchangeMetadata(vhost, name string) (string, persistence.ExchangeProperties, error) {
	return "", persistence.ExchangeProperties{}, nil
}
func (d *DummyPersistence) DeleteExchangeMetadata(vhost, name string) error { return nil }
func (d *DummyPersistence) SaveQueueMetadata(vhost, name string, props persistence.QueueProperties) error {
	return nil
}
func (d *DummyPersistence) LoadQueueMetadata(vhost, name string) (persistence.QueueProperties, error) {
	return persistence.QueueProperties{}, nil
}
func (d *DummyPersistence) DeleteQueueMetadata(vhost, name string) error { return nil }

func (d *DummyPersistence) Initialize() error { return nil }
func (d *DummyPersistence) Close() error      { return nil }

// LoadMessages returns an empty slice
func (d *DummyPersistence) LoadMessages(vhost, queue string) ([]struct {
	ID         string                        `json:"id"`
	Body       []byte                        `json:"body"`
	Properties persistence.MessageProperties `json:"properties"`
}, error) {
	return []struct {
		ID         string                        `json:"id"`
		Body       []byte                        `json:"body"`
		Properties persistence.MessageProperties `json:"properties"`
	}{}, nil
}

// SaveMessage
func (d *DummyPersistence) SaveMessage(vhost, queue, msgId string, msgBody []byte, msgProps persistence.MessageProperties) error {
	return nil
}
