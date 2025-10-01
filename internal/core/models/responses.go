package models

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ConnectionListResponse struct {
	Connections []ConnectionInfoDTO `json:"connections"`
}

type ExchangeListResponse struct {
	Exchanges []ExchangeDTO `json:"exchanges"`
}

type QueueListResponse struct {
	Queues []QueueDTO `json:"queues"`
}
