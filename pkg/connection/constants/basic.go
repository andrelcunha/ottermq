package constants

type BasicMethod int

const (
	BASIC_QOS           BasicMethod = 10
	BASIC_QOS_OK        BasicMethod = 11
	BASIC_CONSUME       BasicMethod = 20
	BASIC_CONSUME_OK    BasicMethod = 21
	BASIC_CANCEL        BasicMethod = 30
	BASIC_CANCEL_OK     BasicMethod = 31
	BASIC_PUBLISH       BasicMethod = 40
	BASIC_RETURN        BasicMethod = 50
	BASIC_DELIVER       BasicMethod = 60
	BASIC_GET           BasicMethod = 70
	BASIC_GET_OK        BasicMethod = 71
	BASIC_GET_EMPTY     BasicMethod = 72
	BASIC_ACK           BasicMethod = 80
	BASIC_REJECT        BasicMethod = 90
	BASIC_RECOVER_ASYNC BasicMethod = 100
	BASIC_RECOVER       BasicMethod = 110
	BASIC_RECOVER_OK    BasicMethod = 111
)
