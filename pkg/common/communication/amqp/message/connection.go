package message

// ClassID: 10 (connection) |
// MethodID: 40 (close)
type ConnectionCloseMessage struct {
	ReplyCode uint16
	ReplyText string
	ClassID   uint16
	MethodID  uint16
}

// ClassID: 10 (connection) |
// MethodID: 41 (close-Ok)
type ConnectionCloseOkMessage struct {
}
