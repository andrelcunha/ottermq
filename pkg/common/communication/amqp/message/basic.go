package message

type BasicPublishMessage struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
}
