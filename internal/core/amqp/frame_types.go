package amqp

type Heartbeat struct{}

type AMQP_Key struct {
	Key  string
	Type string
}

type AMQP_Type struct {
}

type ConnectionStartFrame struct {
	VersionMajor     byte
	VersionMinor     byte
	ServerProperties map[string]interface{}
	Mechanisms       string
	Locales          string
}

type ConnectionStartOkFrame struct {
	ClientProperties map[string]interface{}
	Mechanism        string
	Response         string
	Locale           string
}

type ConnectionTuneFrame struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

type ConnectionOpenFrame struct {
	VirtualHost string
}
