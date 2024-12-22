package basic

type BasicMethod int

const (
	QOS           BasicMethod = 10
	QOS_OK        BasicMethod = 11
	CONSUME       BasicMethod = 20
	CONSUME_OK    BasicMethod = 21
	CANCEL        BasicMethod = 30
	CANCEL_OK     BasicMethod = 31
	PUBLISH       BasicMethod = 40
	RETURN        BasicMethod = 50
	DELIVER       BasicMethod = 60
	GET           BasicMethod = 70
	GET_OK        BasicMethod = 71
	GET_EMPTY     BasicMethod = 72
	ACK           BasicMethod = 80
	REJECT        BasicMethod = 90
	RECOVER_ASYNC BasicMethod = 100
	RECOVER       BasicMethod = 110
	RECOVER_OK    BasicMethod = 111
)
