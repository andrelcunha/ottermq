package message

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
