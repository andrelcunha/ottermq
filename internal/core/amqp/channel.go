package amqp

func closeChannelFrame(channel uint16) []byte {
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  uint16(CHANNEL),
		MethodID: uint16(CHANNEL_CLOSE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}

func createChannelOpenOkFrame(channel uint16, request *RequestMethodMessage) []byte {
	frame := ResponseMethodMessage{
		Channel:  channel,
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
