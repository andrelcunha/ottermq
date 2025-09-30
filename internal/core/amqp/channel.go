package amqp

func closeChannelFrame(request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  uint16(CHANNEL),
		MethodID: uint16(CHANNEL_CLOSE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}

func createChannelOpenOkFrame(request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  request.Channel,
		ClassID:  request.ClassID,
		MethodID: uint16(CHANNEL_OPEN_OK),
		Content: ContentList{
			KeyValuePairs: []KeyValue{
				{
					Key:   INT_LONG,
					Value: uint32(0),
				},
			},
		},
	}.FormatMethodFrame()
	return frame
}
