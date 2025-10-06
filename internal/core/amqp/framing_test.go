package amqp

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestParseFrame_MethodFrame tests parsing of METHOD frames
func TestParseFrame_MethodFrame(t *testing.T) {
	// Create a valid METHOD frame for Channel.Open
	frame := make([]byte, 0)

	// Frame header
	frame = append(frame, byte(TYPE_METHOD))      // frame type
	frame = append(frame, 0x00, 0x01)             // channel 1
	frame = append(frame, 0x00, 0x00, 0x00, 0x08) // payload size (8 bytes)

	// Payload: class ID (CHANNEL=20) + method ID (CHANNEL_OPEN=10) + reserved (4 bytes)
	frame = append(frame, 0x00, 0x14)             // class ID (20 = CHANNEL)
	frame = append(frame, 0x00, 0x0A)             // method ID (10 = CHANNEL_OPEN)
	frame = append(frame, 0x00, 0x00, 0x00, 0x00) // reserved short string (empty)

	result, err := parseFrame(frame)
	if err != nil {
		t.Fatalf("parseFrame failed: %v", err)
	}

	state, ok := result.(*ChannelState)
	if !ok {
		t.Fatalf("Expected *ChannelState, got %T", result)
	}

	if state.MethodFrame == nil {
		t.Fatal("MethodFrame is nil")
	}

	if state.MethodFrame.Channel != 1 {
		t.Errorf("Expected channel 1, got %d", state.MethodFrame.Channel)
	}

	if state.MethodFrame.ClassID != uint16(CHANNEL) {
		t.Errorf("Expected class ID %d, got %d", uint16(CHANNEL), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(CHANNEL_OPEN) {
		t.Errorf("Expected method ID %d, got %d", uint16(CHANNEL_OPEN), state.MethodFrame.MethodID)
	}
}

// TestParseFrame_HeartbeatFrame tests parsing of HEARTBEAT frames
func TestParseFrame_HeartbeatFrame(t *testing.T) {
	// Create a valid HEARTBEAT frame
	frame := make([]byte, 0)

	// Frame header
	frame = append(frame, byte(TYPE_HEARTBEAT))   // frame type
	frame = append(frame, 0x00, 0x00)             // channel 0
	frame = append(frame, 0x00, 0x00, 0x00, 0x00) // payload size (0 bytes)

	result, err := parseFrame(frame)
	if err != nil {
		t.Fatalf("parseFrame failed: %v", err)
	}

	_, ok := result.(*Heartbeat)
	if !ok {
		t.Fatalf("Expected *Heartbeat, got %T", result)
	}
}

// TestParseFrame_TooShort tests error handling for frames that are too short
func TestParseFrame_TooShort(t *testing.T) {
	// Create a frame that's too short (less than 7 bytes header)
	frame := []byte{0x01, 0x00, 0x01}

	_, err := parseFrame(frame)
	if err == nil {
		t.Fatal("Expected error for frame that's too short, got nil")
	}

	expectedErr := "frame too short"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestParseFrame_InvalidPayloadSize tests error handling for invalid payload size
func TestParseFrame_InvalidPayloadSize(t *testing.T) {
	// Create a frame with invalid payload size
	frame := make([]byte, 0)
	frame = append(frame, byte(TYPE_METHOD))      // frame type
	frame = append(frame, 0x00, 0x01)             // channel 1
	frame = append(frame, 0x00, 0x00, 0x00, 0xFF) // payload size (255 bytes)
	// But only provide 10 bytes total (header is 7)
	frame = append(frame, 0x00, 0x00, 0x00)

	_, err := parseFrame(frame)
	if err == nil {
		t.Fatal("Expected error for invalid payload size, got nil")
	}

	expectedErr := "frame too short"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestParseFrame_UnknownFrameType tests error handling for unknown frame types
func TestParseFrame_UnknownFrameType(t *testing.T) {
	// Create a frame with unknown frame type (99)
	frame := make([]byte, 0)
	frame = append(frame, 99)                     // unknown frame type
	frame = append(frame, 0x00, 0x00)             // channel 0
	frame = append(frame, 0x00, 0x00, 0x00, 0x00) // payload size (0 bytes)

	_, err := parseFrame(frame)
	if err == nil {
		t.Fatal("Expected error for unknown frame type, got nil")
	}
}

// TestFormatMethodFrame tests creating a METHOD frame
func TestFormatMethodFrame(t *testing.T) {
	// Create a simple method frame for Channel.Close
	channelNum := uint16(1)
	class := TypeClass(CHANNEL)
	method := TypeMethod(CHANNEL_CLOSE)

	// Create method payload for Channel.Close
	// reply-code (short), reply-text (shortstr), class-id (short), method-id (short)
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, uint16(200)) // reply code
	payload.WriteByte(0)                                  // empty shortstr (length 0)
	binary.Write(&payload, binary.BigEndian, uint16(0))   // class id
	binary.Write(&payload, binary.BigEndian, uint16(0))   // method id

	frame := formatMethodFrame(channelNum, class, method, payload.Bytes())

	// Verify frame structure
	if len(frame) < 8 {
		t.Fatalf("Frame too short: %d bytes", len(frame))
	}

	// Check frame type
	if frame[0] != byte(TYPE_METHOD) {
		t.Errorf("Expected frame type %d, got %d", byte(TYPE_METHOD), frame[0])
	}

	// Check channel number
	channel := binary.BigEndian.Uint16(frame[1:3])
	if channel != channelNum {
		t.Errorf("Expected channel %d, got %d", channelNum, channel)
	}

	// Check payload size
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	expectedPayloadSize := uint32(4 + len(payload.Bytes())) // 4 bytes for class+method IDs
	if payloadSize != expectedPayloadSize {
		t.Errorf("Expected payload size %d, got %d", expectedPayloadSize, payloadSize)
	}

	// Check frame-end octet
	if frame[len(frame)-1] != 0xCE {
		t.Errorf("Expected frame-end 0xCE, got 0x%02X", frame[len(frame)-1])
	}

	// Check class and method IDs in payload
	classID := binary.BigEndian.Uint16(frame[7:9])
	if classID != uint16(class) {
		t.Errorf("Expected class ID %d, got %d", uint16(class), classID)
	}

	methodID := binary.BigEndian.Uint16(frame[9:11])
	if methodID != uint16(method) {
		t.Errorf("Expected method ID %d, got %d", uint16(method), methodID)
	}
}

// TestFormatHeader tests creating frame headers
func TestFormatHeader(t *testing.T) {
	frameType := uint8(TYPE_METHOD)
	channel := uint16(5)
	payloadSize := uint32(100)

	header := formatHeader(frameType, channel, payloadSize)

	// Verify header length
	if len(header) != 7 {
		t.Fatalf("Expected header length 7, got %d", len(header))
	}

	// Verify frame type
	if header[0] != frameType {
		t.Errorf("Expected frame type %d, got %d", frameType, header[0])
	}

	// Verify channel
	ch := binary.BigEndian.Uint16(header[1:3])
	if ch != channel {
		t.Errorf("Expected channel %d, got %d", channel, ch)
	}

	// Verify payload size
	ps := binary.BigEndian.Uint32(header[3:7])
	if ps != payloadSize {
		t.Errorf("Expected payload size %d, got %d", payloadSize, ps)
	}
}

// TestGetSmalestShortInt tests the helper function for getting smallest uint16
func TestGetSmalestShortInt(t *testing.T) {
	tests := []struct {
		name     string
		a, b     uint16
		expected uint16
	}{
		{"Both non-zero, a smaller", 10, 20, 10},
		{"Both non-zero, b smaller", 30, 15, 15},
		{"a is zero", 0, 25, 25},
		{"b is zero", 35, 0, 35},
		{"Both zero", 0, 0, 32767}, // When both are 0, they're treated as MaxInt16 (32767)
		{"Equal values", 50, 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSmalestShortInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("getSmalestShortInt(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestGetSmalestLongInt tests the helper function for getting smallest uint32
func TestGetSmalestLongInt(t *testing.T) {
	tests := []struct {
		name     string
		a, b     uint32
		expected uint32
	}{
		{"Both non-zero, a smaller", 100, 200, 100},
		{"Both non-zero, b smaller", 300, 150, 150},
		{"a is zero", 0, 250, 250},
		{"b is zero", 350, 0, 350},
		{"Both zero", 0, 0, 2147483647}, // When both are 0, they're treated as MaxInt32 (2147483647)
		{"Equal values", 500, 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSmalestLongInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("getSmalestLongInt(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestCreateHeartbeatFrame tests heartbeat frame creation
func TestCreateHeartbeatFrame(t *testing.T) {
	frame := createHeartbeatFrame()

	// Verify frame length (7 byte header + 1 byte frame-end = 8 bytes)
	expectedLength := 8
	if len(frame) != expectedLength {
		t.Fatalf("Expected frame length %d, got %d", expectedLength, len(frame))
	}

	// Verify frame type
	if frame[0] != byte(TYPE_HEARTBEAT) {
		t.Errorf("Expected frame type %d, got %d", byte(TYPE_HEARTBEAT), frame[0])
	}

	// Verify channel (should be 0 for heartbeat)
	channel := binary.BigEndian.Uint16(frame[1:3])
	if channel != 0 {
		t.Errorf("Expected channel 0, got %d", channel)
	}

	// Verify payload size (should be 0 for heartbeat)
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if payloadSize != 0 {
		t.Errorf("Expected payload size 0, got %d", payloadSize)
	}

	// Verify frame-end
	if frame[7] != 0xCE {
		t.Errorf("Expected frame-end 0xCE, got 0x%02X", frame[7])
	}
}

// TestParseFrame_BodyFrame tests parsing of BODY frames
func TestParseFrame_BodyFrame(t *testing.T) {
	// Create a valid BODY frame
	bodyContent := []byte("Hello, AMQP!")
	frame := make([]byte, 0)

	// Frame header
	frame = append(frame, byte(TYPE_BODY)) // frame type
	frame = append(frame, 0x00, 0x01)      // channel 1

	// Payload size (length of body content)
	payloadSize := uint32(len(bodyContent))
	var sizeBuf [4]byte
	binary.BigEndian.PutUint32(sizeBuf[:], payloadSize)
	frame = append(frame, sizeBuf[:]...)

	// Payload (body content)
	frame = append(frame, bodyContent...)

	result, err := parseFrame(frame)
	if err != nil {
		t.Fatalf("parseFrame failed: %v", err)
	}

	state, ok := result.(*ChannelState)
	if !ok {
		t.Fatalf("Expected *ChannelState, got %T", result)
	}

	if state.Body == nil {
		t.Fatal("Body is nil")
	}

	if !bytes.Equal(state.Body, bodyContent) {
		t.Errorf("Expected body %q, got %q", bodyContent, state.Body)
	}
}

// TestParseMethodFrame_TooShort tests error handling for method frames with short payload
func TestParseMethodFrame_TooShort(t *testing.T) {
	// Create a method frame with payload that's too short (less than 4 bytes for class+method)
	frame := make([]byte, 0)

	frame = append(frame, byte(TYPE_METHOD))      // frame type
	frame = append(frame, 0x00, 0x01)             // channel 1
	frame = append(frame, 0x00, 0x00, 0x00, 0x02) // payload size (2 bytes - too short)
	frame = append(frame, 0x00, 0x14)             // only class ID, missing method ID

	_, err := parseFrame(frame)
	if err == nil {
		t.Fatal("Expected error for method frame with short payload, got nil")
	}
}

// TestDefaultFramer_ParseFrame tests the DefaultFramer's ParseFrame method
func TestDefaultFramer_ParseFrame(t *testing.T) {
	framer := &DefaultFramer{}

	// Create a simple heartbeat frame
	frame := make([]byte, 0)
	frame = append(frame, byte(TYPE_HEARTBEAT))   // frame type
	frame = append(frame, 0x00, 0x00)             // channel 0
	frame = append(frame, 0x00, 0x00, 0x00, 0x00) // payload size (0 bytes)

	result, err := framer.ParseFrame(frame)
	if err != nil {
		t.Fatalf("ParseFrame failed: %v", err)
	}

	_, ok := result.(*Heartbeat)
	if !ok {
		t.Fatalf("Expected *Heartbeat, got %T", result)
	}
}

// TestFrameTypeConstants tests that frame type constants are correct per AMQP 0-9-1
func TestFrameTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant FrameType
		expected uint8
	}{
		{"METHOD frame type", TYPE_METHOD, 1},
		{"HEADER frame type", TYPE_HEADER, 2},
		{"BODY frame type", TYPE_BODY, 3},
		{"HEARTBEAT frame type", TYPE_HEARTBEAT, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint8(tt.constant) != tt.expected {
				t.Errorf("%s = %d, expected %d", tt.name, uint8(tt.constant), tt.expected)
			}
		})
	}
}

// TestFrameEndConstant tests that FRAME_END constant is correct (0xCE)
func TestFrameEndConstant(t *testing.T) {
	if FRAME_END != 0xCE {
		t.Errorf("FRAME_END = 0x%02X, expected 0xCE", FRAME_END)
	}
}

// TestFormatMethodPayload tests the method payload formatting
func TestFormatMethodPayload(t *testing.T) {
	tests := []struct {
		name     string
		content  ContentList
		validate func(t *testing.T, payload []byte)
	}{
		{
			name: "Empty content list",
			content: ContentList{
				KeyValuePairs: []KeyValue{},
			},
			validate: func(t *testing.T, payload []byte) {
				if len(payload) != 0 {
					t.Errorf("Expected empty payload, got %d bytes", len(payload))
				}
			},
		},
		{
			name: "Single short int",
			content: ContentList{
				KeyValuePairs: []KeyValue{
					{Key: INT_SHORT, Value: uint16(42)},
				},
			},
			validate: func(t *testing.T, payload []byte) {
				if len(payload) != 2 {
					t.Errorf("Expected 2 bytes, got %d", len(payload))
				}
				value := binary.BigEndian.Uint16(payload)
				if value != 42 {
					t.Errorf("Expected value 42, got %d", value)
				}
			},
		},
		{
			name: "Single octet",
			content: ContentList{
				KeyValuePairs: []KeyValue{
					{Key: INT_OCTET, Value: uint8(7)},
				},
			},
			validate: func(t *testing.T, payload []byte) {
				if len(payload) != 1 {
					t.Errorf("Expected 1 byte, got %d", len(payload))
				}
				if payload[0] != 7 {
					t.Errorf("Expected value 7, got %d", payload[0])
				}
			},
		},
		{
			name: "Boolean bit true",
			content: ContentList{
				KeyValuePairs: []KeyValue{
					{Key: BIT, Value: true},
				},
			},
			validate: func(t *testing.T, payload []byte) {
				if len(payload) != 1 {
					t.Errorf("Expected 1 byte, got %d", len(payload))
				}
				if payload[0] != 1 {
					t.Errorf("Expected value 1, got %d", payload[0])
				}
			},
		},
		{
			name: "Boolean bit false",
			content: ContentList{
				KeyValuePairs: []KeyValue{
					{Key: BIT, Value: false},
				},
			},
			validate: func(t *testing.T, payload []byte) {
				if len(payload) != 1 {
					t.Errorf("Expected 1 byte, got %d", len(payload))
				}
				if payload[0] != 0 {
					t.Errorf("Expected value 0, got %d", payload[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := formatMethodPayload(tt.content)
			tt.validate(t, payload)
		})
	}
}
