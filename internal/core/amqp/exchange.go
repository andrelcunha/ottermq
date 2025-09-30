package amqp

func createExchangeDeclareFrame(request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(EXCHANGE_DECLARE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}
