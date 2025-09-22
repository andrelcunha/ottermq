package amqp

type ChannelCloseMessage struct {
	ReplyCode uint16
	ReplyText string
	ClassID   uint16
	MethodID  uint16
}

type ChannelOpenMessage struct {
}

// ClassID: 10 (connection) |
// MethodID: 40 (close)
type ConnectionCloseMessage struct {
	ReplyCode uint16
	ReplyText string
	ClassID   uint16
	MethodID  uint16
}

// ClassID: 10 (connection) |
// MethodID: 41 (close-Ok)
type ConnectionCloseOkMessage struct {
}

type ExchangeDeclareMessage struct {
	ExchangeName string
	ExchangeType string
	Durable      bool
	AutoDelete   bool
	Internal     bool
	NoWait       bool
	Arguments    map[string]interface{}
}

type ExchangeDeleteMessage struct {
	ExchangeName string
	IfUnused     bool
	NoWait       bool
}

type QueueDeclareMessage struct {
	QueueName string
	Durable   bool
	IfUnused  bool
	Exclusive bool
	NoWait    bool
	Arguments map[string]interface{}
}

type QueueDeleteMessage struct {
	QueueName string
	IfUnused  bool
	NoWait    bool
}

type QueueBindMessage struct {
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
	Arguments  map[string]interface{}
}
