package queue

type QueueMethod int

const (
	DECLARE    QueueMethod = 10
	DECLARE_OK QueueMethod = 11
	BIND       QueueMethod = 20
	BIND_OK    QueueMethod = 21
	UNBIND     QueueMethod = 50
	UNBIND_OK  QueueMethod = 51
	DELETE     QueueMethod = 40
	DELETE_OK  QueueMethod = 41
)
