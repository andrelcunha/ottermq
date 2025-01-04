package constants

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
