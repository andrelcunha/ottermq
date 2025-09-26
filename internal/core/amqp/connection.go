package amqp

func closeConnectionFrame(channel uint16) []byte {
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  uint16(CONNECTION),
		MethodID: uint16(CONNECTION_CLOSE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}
