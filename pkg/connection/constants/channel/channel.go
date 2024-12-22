package channel

type ChannelMethod int

const (
	OPEN     ChannelMethod = 10
	OPEN_OK  ChannelMethod = 11
	FLOW     ChannelMethod = 20
	FLOW_OK  ChannelMethod = 21
	CLOSE    ChannelMethod = 40
	CLOSE_OK ChannelMethod = 41
)
