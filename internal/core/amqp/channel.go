package amqp

func closeChannelFrame(channel, replycode, classID, methodID uint16, replyText string) []byte {
	replyCodeKv := KeyValue{
		Key:   INT_SHORT,
		Value: replycode,
	}
	replyTextKv := KeyValue{
		Key:   STRING_SHORT,
		Value: replyText,
	}
	classIDKv := KeyValue{
		Key:   INT_SHORT,
		Value: classID,
	}
	methodIDKv := KeyValue{
		Key:   INT_SHORT,
		Value: methodID,
	}
	content := ContentList{
		KeyValuePairs: []KeyValue{replyCodeKv, replyTextKv, classIDKv, methodIDKv},
	}
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  uint16(CHANNEL),
		MethodID: uint16(CHANNEL_CLOSE_OK),
		Content:  content,
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

func createChannelCloseOkFrame(channel uint16) []byte {
	frame := ResponseMethodMessage{
		Channel:  channel,
		ClassID:  uint16(CHANNEL),
		MethodID: uint16(CHANNEL_CLOSE_OK),
		Content:  ContentList{},
	}.FormatMethodFrame()
	return frame
}
