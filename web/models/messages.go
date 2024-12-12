package models

type PublishMessageRequest struct {
	ExchangeName string `json:"exchange_name"`
	RoutingKey   string `json:"routing_key"`
	Message      string `json:"message"`
}
