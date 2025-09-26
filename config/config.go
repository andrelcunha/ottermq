package config

type Config struct {
	Port                 string
	Host                 string
	Username             string
	Password             string
	HeartbeatIntervalMax uint16
	ChannelMax           uint16
	FrameMax             uint32
	Version              string
	Ssl                  bool
}
