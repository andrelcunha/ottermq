package amqp

const (
	INT_OCTET     = "octet"
	INT_SHORT     = "short"
	INT_LONG      = "long"
	INT_LONG_LONG = "long-long"
	BIT           = "bit"
	STRING_SHORT  = "short-str"
	STRING_LONG   = "long-str"
	TIMESTAMP     = "timestamp"
	TABLE         = "table"
)

const (
	AMQP_PROTOCOL_HEADER = "AMQP\x00\x00\x09\x01"
)

type TypeMethod uint16

type QueueMethod int

const (
	QUEUE_DECLARE    QueueMethod = 10
	QUEUE_DECLARE_OK QueueMethod = 11
	QUEUE_BIND       QueueMethod = 20
	QUEUE_BIND_OK    QueueMethod = 21
	QUEUE_UNBIND     QueueMethod = 50
	QUEUE_UNBIND_OK  QueueMethod = 51
	QUEUE_DELETE     QueueMethod = 40
	QUEUE_DELETE_OK  QueueMethod = 41
)

type FrameType uint8

const (
	TYPE_METHOD    FrameType = 1
	TYPE_HEADER    FrameType = 2
	TYPE_BODY      FrameType = 3
	TYPE_HEARTBEAT FrameType = 8
)

type ExchangeMethod int

const (
	EXCHANGE_DECLARE    ExchangeMethod = 10
	EXCHANGE_DECLARE_OK ExchangeMethod = 11
	EXCHANGE_DELETE     ExchangeMethod = 20
	EXCHANGE_DELETE_OK  ExchangeMethod = 21
)

const (
	CONNECTION_START     TypeMethod = 10
	CONNECTION_START_OK  TypeMethod = 11
	CONNECTION_SECURE    TypeMethod = 20
	CONNECTION_SECURE_OK TypeMethod = 21
	CONNECTION_TUNE      TypeMethod = 30
	CONNECTION_TUNE_OK   TypeMethod = 31
	CONNECTION_OPEN      TypeMethod = 40
	CONNECTION_OPEN_OK   TypeMethod = 41
	CONNECTION_CLOSE     TypeMethod = 50
	CONNECTION_CLOSE_OK  TypeMethod = 51
)

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

type ChannelMethod int

const (
	CHANNEL_OPEN     ChannelMethod = 10
	CHANNEL_OPEN_OK  ChannelMethod = 11
	CHANNEL_FLOW     ChannelMethod = 20
	CHANNEL_FLOW_OK  ChannelMethod = 21
	CHANNEL_CLOSE    ChannelMethod = 40
	CHANNEL_CLOSE_OK ChannelMethod = 41
)

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

type TxMethod int

const (
	SELECT      TxMethod = 10
	SELECT_OK   TxMethod = 11
	COMMIT      TxMethod = 20
	COMMIT_OK   TxMethod = 21
	ROLLBACK    TxMethod = 30
	ROLLBACK_OK TxMethod = 31
)
