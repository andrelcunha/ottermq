package amqp

func createExchangeDeclareFrame(channel uint16, request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  request.ClassID,
		MethodID: uint16(EXCHANGE_DECLARE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}
