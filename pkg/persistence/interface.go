package persistence

// Persistence defines the interface for all storage backends
type Persistence interface {
	// Queue operations
	SaveQueueMetadata(vhost, name string, props QueueProperties) error

	// LoadQueueMetadata loads a single queue's metadata from a JSON file (props)
	LoadQueueMetadata(vhost, name string) (props QueueProperties, err error)
	DeleteQueueMetadata(vhost, name string) error

	// Exchange operations
	SaveExchangeMetadata(vhost, name, exchangeType string, props ExchangeProperties) error

	// LoadExchangeMetadata loads exchange metadata from a JSON file (type and properties)
	LoadExchangeMetadata(vhost, name string) (exchangeType string, props ExchangeProperties, err error)

	DeleteExchangeMetadata(vhost, name string) error

	// Message operations (to be defined)
	SaveMessage(vhost, queue, msgId string, msgBody []byte, msgProps MessageProperties) error
	LoadMessages(vhostName, queueName string) ([]struct {
		ID         string            `json:"id"`
		Body       []byte            `json:"body"`
		Properties MessageProperties `json:"properties"`
	}, error)

	// Lifecycle
	Initialize() error
	Close() error

	// Health/Status (TBD)
	// HealthCheck() error
}

// Config for persistence implementations
type Config struct {
	Type    string            `json:"type"`     // "json", "memento", etc.
	DataDir string            `json:"data_dir"` // Base directory for storage
	Options map[string]string `json:"options"`  // Implementation-specific options
}
