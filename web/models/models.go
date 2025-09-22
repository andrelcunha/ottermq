package models

type FiberMap map[string]interface{}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BindQueueRequest struct {
	ExchangeName string `json:"exchange_name"`
	QueueName    string `json:"queue_name"`
	RoutingKey   string `json:"routing_key"`
}

type CreateExchangeRequest struct {
	ExchangeName string `json:"exchange_name"`
	ExchangeType string `json:"exchange_type"`
	// VhostId      string `json:"vhost_id"`
}

type CreateQueueRequest struct {
	QueueName string `json:"queue_name"`
}

type DeleteBindingRequest struct {
	ExchangeName string `json:"exchange_name"`
	QueueName    string `json:"queue_name"`
	RoutingKey   string `json:"routing_key"`
}

type PublishMessageRequest struct {
	ExchangeName string `json:"exchange_name"`
	RoutingKey   string `json:"routing_key"`
	Message      string `json:"message"`
}
