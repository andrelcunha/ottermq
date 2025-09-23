package amqp

type AMQPConfig struct {
	Mechanisms        []string
	Locales           []string
	HeartbeatInterval int
	FrameMax          int
	ChannelMax        int
}
