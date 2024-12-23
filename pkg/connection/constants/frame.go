package constants

type FrameType uint8

const (
	TYPE_METHOD    FrameType = 1
	TYPE_HEADER    FrameType = 2
	TYPE_BODY      FrameType = 3
	TYPE_HEARTBEAT FrameType = 8
)
