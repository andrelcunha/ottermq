package persistence

// Universal models that all implementations can use
type MessageProperties struct {
	ContentType     string         `json:"content_type,omitempty"`
	ContentEncoding string         `json:"content_encoding,omitempty"`
	Headers         map[string]any `json:"headers,omitempty"`
	DeliveryMode    uint8          `json:"delivery_mode,omitempty"`
	Priority        uint8          `json:"priority,omitempty"`
	CorrelationID   string         `json:"correlation_id,omitempty"`
	ReplyTo         string         `json:"reply_to,omitempty"`
	Expiration      string         `json:"expiration,omitempty"`
	MessageID       string         `json:"message_id,omitempty"`
	Timestamp       int64          `json:"timestamp,omitempty"`
	Type            string         `json:"type,omitempty"`
	UserID          string         `json:"user_id,omitempty"`
	AppID           string         `json:"app_id,omitempty"`
}

// Basic queue/exchange properties - universal concepts
type QueueProperties struct {
	Passive    bool           `json:"passive"` // Not needed in persistence*
	Durable    bool           `json:"durable"`
	Exclusive  bool           `json:"exclusive"`
	AutoDelete bool           `json:"auto_delete"`
	NoWait     bool           `json:"no_wait"` // Not needed in persistence*
	Arguments  map[string]any `json:"arguments"`
}

type ExchangeProperties struct {
	Passive    bool           `json:"passive"` // Not needed in persistence*
	Durable    bool           `json:"durable"`
	AutoDelete bool           `json:"auto_delete"`
	Internal   bool           `json:"internal"`
	NoWait     bool           `json:"no_wait"` // Not needed in persistence*
	Arguments  map[string]any `json:"arguments"`
}

// * Passive and NoWait are not needed in persistence
// because they are only relevant during declaration time
// Even though they are included here for completeness,
// and it could make sense to persist them if we decide to
// track declaration history
