package amqp

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// TestBuildProtocolHeader tests building protocol headers for different AMQP versions
func TestBuildProtocolHeader(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		expected []byte
	}{
		{
			name:     "AMQP 0-9-1",
			protocol: "AMQP 0-9-1",
			expected: []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := buildProtocolHeader(tt.protocol)
			if err != nil {
				t.Fatalf("buildProtocolHeader failed: %v", err)
			}

			if !bytes.Equal(header, tt.expected) {
				t.Errorf("Expected header %v, got %v", tt.expected, header)
			}
		})
	}
}

// TestBuildProtocolHeader_InvalidProtocol tests error handling for invalid protocols
func TestBuildProtocolHeader_InvalidProtocol(t *testing.T) {
	_, err := buildProtocolHeader("INVALID")
	if err == nil {
		t.Fatal("Expected error for invalid protocol, got nil")
	}
}

// TestCreateHeartbeatFrame_Structure tests the structure of heartbeat frames
func TestCreateHeartbeatFrame_Structure(t *testing.T) {
	frame := createHeartbeatFrame()

	// Expected structure:
	// [0] = TYPE_HEARTBEAT (8)
	// [1:3] = channel (0)
	// [3:7] = payload size (0)
	// [7] = frame-end (0xCE)

	expectedLength := 8
	if len(frame) != expectedLength {
		t.Fatalf("Expected frame length %d, got %d", expectedLength, len(frame))
	}

	// Verify frame type
	if frame[0] != byte(TYPE_HEARTBEAT) {
		t.Errorf("Expected frame type %d, got %d", byte(TYPE_HEARTBEAT), frame[0])
	}

	// Verify channel is 0
	channel := binary.BigEndian.Uint16(frame[1:3])
	if channel != 0 {
		t.Errorf("Expected channel 0, got %d", channel)
	}

	// Verify payload size is 0
	payloadSize := binary.BigEndian.Uint32(frame[3:7])
	if payloadSize != 0 {
		t.Errorf("Expected payload size 0, got %d", payloadSize)
	}

	// Verify frame-end is 0xCE
	if frame[7] != 0xCE {
		t.Errorf("Expected frame-end 0xCE, got 0x%02X", frame[7])
	}
}

// mockConn is a mock implementation of net.Conn for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func newMockConn(data []byte) *mockConn {
	return &mockConn{
		readBuf:  bytes.NewBuffer(data),
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5672}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// TestReadProtocolHeader tests reading protocol headers
func TestReadProtocolHeader(t *testing.T) {
	// Create a mock connection with a valid AMQP 0-9-1 protocol header
	header := []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
	conn := newMockConn(header)

	receivedHeader, err := readProtocolHeader(conn)
	if err != nil {
		t.Fatalf("readProtocolHeader failed: %v", err)
	}

	if !bytes.Equal(receivedHeader, header) {
		t.Errorf("Expected header %v, got %v", header, receivedHeader)
	}
}

// TestReadProtocolHeader_Incomplete tests error handling for incomplete headers
func TestReadProtocolHeader_Incomplete(t *testing.T) {
	// Create a mock connection with incomplete header (only 5 bytes)
	incompleteHeader := []byte{'A', 'M', 'Q', 'P', 0}
	conn := newMockConn(incompleteHeader)

	_, err := readProtocolHeader(conn)
	if err == nil {
		t.Fatal("Expected error for incomplete header, got nil")
	}
}

// TestSendProtocolHeader tests sending protocol headers
func TestSendProtocolHeader(t *testing.T) {
	header := []byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1}
	conn := newMockConn(nil)

	err := sendProtocolHeader(conn, header)
	if err != nil {
		t.Fatalf("sendProtocolHeader failed: %v", err)
	}

	sentData := conn.writeBuf.Bytes()
	if !bytes.Equal(sentData, header) {
		t.Errorf("Expected sent data %v, got %v", header, sentData)
	}
}

// TestReadFrame_ValidFrame tests reading a valid frame
func TestReadFrame_ValidFrame(t *testing.T) {
	// Create a valid heartbeat frame
	frameData := make([]byte, 0)
	frameData = append(frameData, byte(TYPE_HEARTBEAT))   // frame type
	frameData = append(frameData, 0x00, 0x00)             // channel 0
	frameData = append(frameData, 0x00, 0x00, 0x00, 0x00) // payload size 0
	frameData = append(frameData, FRAME_END)              // frame-end

	conn := newMockConn(frameData)

	receivedFrame, err := readFrame(conn)
	if err != nil {
		t.Fatalf("readFrame failed: %v", err)
	}

	// readFrame returns header + payload (without frame-end)
	expectedFrame := frameData[:len(frameData)-1] // all except frame-end
	if !bytes.Equal(receivedFrame, expectedFrame) {
		t.Errorf("Expected frame %v, got %v", expectedFrame, receivedFrame)
	}
}

// TestReadFrame_InvalidFrameEnd tests error handling for invalid frame-end
func TestReadFrame_InvalidFrameEnd(t *testing.T) {
	// Create a frame with invalid frame-end (not 0xCE)
	frameData := make([]byte, 0)
	frameData = append(frameData, byte(TYPE_HEARTBEAT))   // frame type
	frameData = append(frameData, 0x00, 0x00)             // channel 0
	frameData = append(frameData, 0x00, 0x00, 0x00, 0x00) // payload size 0
	frameData = append(frameData, 0xFF)                   // invalid frame-end

	conn := newMockConn(frameData)

	_, err := readFrame(conn)
	if err == nil {
		t.Fatal("Expected error for invalid frame-end, got nil")
	}

	expectedErr := "invalid frame end octet"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestReadFrame_WithPayload tests reading a frame with payload
func TestReadFrame_WithPayload(t *testing.T) {
	// Create a method frame with some payload
	payload := []byte{0x00, 0x0A, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00}

	frameData := make([]byte, 0)
	frameData = append(frameData, byte(TYPE_METHOD)) // frame type
	frameData = append(frameData, 0x00, 0x01)        // channel 1

	// Payload size
	var sizeBuf [4]byte
	binary.BigEndian.PutUint32(sizeBuf[:], uint32(len(payload)))
	frameData = append(frameData, sizeBuf[:]...)

	frameData = append(frameData, payload...) // payload
	frameData = append(frameData, FRAME_END)  // frame-end

	conn := newMockConn(frameData)

	receivedFrame, err := readFrame(conn)
	if err != nil {
		t.Fatalf("readFrame failed: %v", err)
	}

	// Verify frame type
	if receivedFrame[0] != byte(TYPE_METHOD) {
		t.Errorf("Expected frame type %d, got %d", byte(TYPE_METHOD), receivedFrame[0])
	}

	// Verify channel
	channel := binary.BigEndian.Uint16(receivedFrame[1:3])
	if channel != 1 {
		t.Errorf("Expected channel 1, got %d", channel)
	}

	// Verify payload size
	payloadSize := binary.BigEndian.Uint32(receivedFrame[3:7])
	if payloadSize != uint32(len(payload)) {
		t.Errorf("Expected payload size %d, got %d", len(payload), payloadSize)
	}

	// Verify payload
	receivedPayload := receivedFrame[7:]
	if !bytes.Equal(receivedPayload, payload) {
		t.Errorf("Expected payload %v, got %v", payload, receivedPayload)
	}
}

// TestSendFrame tests sending a frame
func TestSendFrame(t *testing.T) {
	frame := createHeartbeatFrame()
	conn := newMockConn(nil)

	err := sendFrame(conn, frame)
	if err != nil {
		t.Fatalf("sendFrame failed: %v", err)
	}

	sentData := conn.writeBuf.Bytes()
	if !bytes.Equal(sentData, frame) {
		t.Errorf("Expected sent data %v, got %v", frame, sentData)
	}
}

// TestReadFrame_IncompleteHeader tests error handling for incomplete frame header
func TestReadFrame_IncompleteHeader(t *testing.T) {
	// Only 5 bytes instead of 7-byte header
	frameData := []byte{0x01, 0x00, 0x00, 0x00, 0x00}
	conn := newMockConn(frameData)

	_, err := readFrame(conn)
	if err == nil {
		t.Fatal("Expected error for incomplete header, got nil")
	}
}

// TestReadFrame_IncompletePayload tests error handling for incomplete payload
func TestReadFrame_IncompletePayload(t *testing.T) {
	// Header claims payload size of 10, but only provide 5
	frameData := make([]byte, 0)
	frameData = append(frameData, byte(TYPE_METHOD))            // frame type
	frameData = append(frameData, 0x00, 0x01)                   // channel 1
	frameData = append(frameData, 0x00, 0x00, 0x00, 0x0A)       // payload size 10
	frameData = append(frameData, 0x00, 0x00, 0x00, 0x00, 0x00) // only 5 bytes

	conn := newMockConn(frameData)

	_, err := readFrame(conn)
	if err == nil {
		t.Fatal("Expected error for incomplete payload, got nil")
	}
}

// TestFrameRoundTrip tests sending and receiving a complete frame
func TestFrameRoundTrip(t *testing.T) {
	// Create a complete METHOD frame
	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.BigEndian, uint16(CHANNEL)) // Error ignored as bytes.Buffer.Write never fails
	_ = binary.Write(&payload, binary.BigEndian, uint16(CHANNEL_OPEN)) // Error ignored as bytes.Buffer.Write never fails
	payload.Write([]byte{0x00}) // empty short string

	channelNum := uint16(1)
	frame := formatMethodFrame(channelNum, TypeClass(CHANNEL), TypeMethod(CHANNEL_OPEN), payload.Bytes())

	// Send the frame
	sendConn := newMockConn(nil)
	err := sendFrame(sendConn, frame)
	if err != nil {
		t.Fatalf("sendFrame failed: %v", err)
	}

	// Read the frame back
	sentData := sendConn.writeBuf.Bytes()
	readConn := newMockConn(sentData)

	receivedFrame, err := readFrame(readConn)
	if err != nil {
		t.Fatalf("readFrame failed: %v", err)
	}

	// The received frame should be the original frame without the frame-end byte
	expectedFrame := frame[:len(frame)-1]
	if !bytes.Equal(receivedFrame, expectedFrame) {
		t.Errorf("Frame round-trip mismatch.\nExpected: %v\nGot: %v", expectedFrame, receivedFrame)
	}

	// Parse the received frame to ensure it's valid
	result, err := parseFrame(receivedFrame)
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

	if state.MethodFrame.Channel != channelNum {
		t.Errorf("Expected channel %d, got %d", channelNum, state.MethodFrame.Channel)
	}

	if state.MethodFrame.ClassID != uint16(CHANNEL) {
		t.Errorf("Expected class ID %d, got %d", uint16(CHANNEL), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(CHANNEL_OPEN) {
		t.Errorf("Expected method ID %d, got %d", uint16(CHANNEL_OPEN), state.MethodFrame.MethodID)
	}
}

// TestMultipleFramesSequence tests reading multiple frames in sequence
func TestMultipleFramesSequence(t *testing.T) {
	// Create two heartbeat frames in sequence
	frame1 := createHeartbeatFrame()
	frame2 := createHeartbeatFrame()

	// Concatenate both frames
	var data bytes.Buffer
	data.Write(frame1)
	data.Write(frame2)

	conn := newMockConn(data.Bytes())

	// Read first frame
	receivedFrame1, err := readFrame(conn)
	if err != nil {
		t.Fatalf("readFrame (1st) failed: %v", err)
	}

	expectedFrame := frame1[:len(frame1)-1] // without frame-end
	if !bytes.Equal(receivedFrame1, expectedFrame) {
		t.Errorf("First frame mismatch")
	}

	// Read second frame
	receivedFrame2, err := readFrame(conn)
	if err != nil {
		t.Fatalf("readFrame (2nd) failed: %v", err)
	}

	if !bytes.Equal(receivedFrame2, expectedFrame) {
		t.Errorf("Second frame mismatch")
	}
}

// TestFrameMaxSizeCompliance tests frame size limits according to AMQP 0-9-1
func TestFrameMaxSizeCompliance(t *testing.T) {
	tests := []struct {
		name        string
		payloadSize uint32
		shouldPass  bool
	}{
		{"Small frame (100 bytes)", 100, true},
		{"Medium frame (1KB)", 1024, true},
		{"Large frame (128KB)", 131072, true},
		{"Standard max (128KB)", 131072, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a frame with specified payload size
			payload := make([]byte, tt.payloadSize)
			for i := range payload {
				payload[i] = byte(i % 256)
			}

			frameData := make([]byte, 0, 7+tt.payloadSize+1)
			frameData = append(frameData, byte(TYPE_BODY)) // frame type
			frameData = append(frameData, 0x00, 0x01)      // channel 1

			// Payload size
			var sizeBuf [4]byte
			binary.BigEndian.PutUint32(sizeBuf[:], tt.payloadSize)
			frameData = append(frameData, sizeBuf[:]...)

			frameData = append(frameData, payload...) // payload
			frameData = append(frameData, FRAME_END)  // frame-end

			conn := newMockConn(frameData)

			receivedFrame, err := readFrame(conn)
			if tt.shouldPass {
				if err != nil {
					t.Fatalf("readFrame failed: %v", err)
				}

				// Verify payload size in received frame
				receivedPayloadSize := binary.BigEndian.Uint32(receivedFrame[3:7])
				if receivedPayloadSize != tt.payloadSize {
					t.Errorf("Expected payload size %d, got %d", tt.payloadSize, receivedPayloadSize)
				}
			}
		})
	}
}
