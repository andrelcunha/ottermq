package models

type CreateExchangeRequest struct {
	ExchangeName string `json:"exchange_name"`
	ExchangeType string `json:"exchange_type"`
	// VhostId      string `json:"vhost_id"`
}
