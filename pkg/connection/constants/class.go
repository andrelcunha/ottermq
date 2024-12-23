package constants

type TypeClass int

// Class constants
const (
	CONNECTION TypeClass = 10
	CHANNEL    TypeClass = 20
	EXCHANGE   TypeClass = 40
	QUEUE      TypeClass = 50
	BASIC      TypeClass = 60
	TX         TypeClass = 90
)
