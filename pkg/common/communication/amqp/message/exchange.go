package message

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
