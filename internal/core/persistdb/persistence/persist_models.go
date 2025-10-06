package persistence

type ExchangePropertiesDb struct {
	Durable    bool           `json:"durable"`
	AutoDelete bool           `json:"auto_delete"`
	Internal   bool           `json:"internal"`
	NoWait     bool           `json:"no_wait"`
	Arguments  map[string]any `json:"arguments"`
}

type QProps struct {
	Passive    bool           `json:"passive"`
	Durable    bool           `json:"durable"`
	Exclusive  bool           `json:"exclusive"`
	AutoDelete bool           `json:"auto_delete"`
	NoWait     bool           `json:"no_wait"`
	Arguments  map[string]any `json:"arguments"`
}

type MessageProperties struct {
	ContentType     string         `json:"content_type,omitempty"`
	ContentEncoding string         `json:"content_encoding,omitempty"`
	Headers         map[string]any `json:"headers,omitempty"`
	DeliveryMode    uint8          `json:"delivery_mode,omitempty"` // 1 = non-persistent, 2 = persistent
	Priority        uint8          `json:"priority,omitempty"`
	CorrelationID   string         `json:"correlation_id,omitempty"`
	ReplyTo         string         `json:"reply_to,omitempty"`
	Expiration      string         `json:"expiration,omitempty"`
	MessageID       string         `json:"message_id,omitempty"`
	Timestamp       int64          `json:"timestamp,omitempty"` // Unix timestamp
	Type            string         `json:"type,omitempty"`
	UserID          string         `json:"user_id,omitempty"`
	AppID           string         `json:"app_id,omitempty"`
	ClusterID       string         `json:"cluster_id,omitempty"` // deprecated
}

type PersistedMessage struct {
	ID         string            `json:"id"`
	Body       []byte            `json:"body"`
	Properties MessageProperties `json:"properties"`
}

type PersistedQueue struct {
	Name       string             `json:"name"`
	Properties QProps             `json:"properties"`
	Messages   []PersistedMessage `json:"messages"`
}

type PersistedBinding struct {
	QueueName  string         `json:"queue_name"`
	RoutingKey string         `json:"routing_key"`
	Arguments  map[string]any `json:"arguments"`
}

type PersistedExchange struct {
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	Properties ExchangePropertiesDb `json:"properties"`
	Bindings   []PersistedBinding   `json:"bindings"`
}

type PersistedVHost struct {
	Name      string              `json:"name"`
	Exchanges []PersistedExchange `json:"exchanges"`
	Queues    []PersistedQueue    `json:"queues"`
}

type PersistedBrokerState struct {
	VHosts  []PersistedVHost `json:"vhosts"`
	Version int              `json:"version"`
}
