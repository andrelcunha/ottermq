package models

type BindQueueRequest struct {
	ExchangeName string `json:"exchange_name"`
	QueueName    string `json:"queue_name"`
	RoutingKey   string `json:"routing_key"`
}

type DeleteBindingRequest struct {
	ExchangeName string `json:"exchange_name"`
	QueueName    string `json:"queue_name"`
	RoutingKey   string `json:"routing_key"`
}
