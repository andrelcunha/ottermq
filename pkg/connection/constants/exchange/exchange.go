package exchange

type ExchangeMethod int

const (
	DECLARE    ExchangeMethod = 10
	DECLARE_OK ExchangeMethod = 11
	DELETE     ExchangeMethod = 20
	DELETE_OK  ExchangeMethod = 21
)
