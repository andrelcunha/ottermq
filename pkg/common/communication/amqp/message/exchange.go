package message

type ExchangeDeclareMessage struct {
	ExchangeName string
	Type         string
	Durable      bool
	AutoDelete   bool
	Internal     bool
	NoWait       bool
	Arguments    map[string]interface{}
}
