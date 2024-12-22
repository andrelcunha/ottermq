package connection

type ConnectionMethod int

const (
	START     ConnectionMethod = 10
	START_OK  ConnectionMethod = 11
	SECURE    ConnectionMethod = 20
	SECURE_OK ConnectionMethod = 21
	TUNE      ConnectionMethod = 30
	TUNE_OK   ConnectionMethod = 31
	OPEN      ConnectionMethod = 40
	OPEN_OK   ConnectionMethod = 41
	CLOSE     ConnectionMethod = 50
	CLOSE_OK  ConnectionMethod = 51
)
