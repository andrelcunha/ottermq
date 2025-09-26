package amqp

type Heartbeat struct{}

type AMQP_Key struct {
	Key  string
	Type string
}

type AMQP_Type struct {
}

// Received by client
// -- this is legacy from client implementation
// -- Dont use it!
type ConnectionStartFrame struct {
	VersionMajor     byte
	VersionMinor     byte
	ServerProperties map[string]interface{}
	Mechanisms       string
	Locales          string
}

type ConnectionStartOk struct {
	ClientProperties map[string]any
	Mechanism        string
	Response         string
	Locale           string
}

type ConnectionTune struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

type ConnectionOpen struct {
	VirtualHost string
}
