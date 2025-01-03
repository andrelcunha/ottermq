package constants

type ChannelMethod int

const (
	CHANNEL_OPEN     ChannelMethod = 10
	CHANNEL_OPEN_OK  ChannelMethod = 11
	CHANNEL_FLOW     ChannelMethod = 20
	CHANNEL_FLOW_OK  ChannelMethod = 21
	CHANNEL_CLOSE    ChannelMethod = 40
	CHANNEL_CLOSE_OK ChannelMethod = 41
)
