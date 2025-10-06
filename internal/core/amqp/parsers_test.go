package amqp

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestParseHeaderFrame tests parsing of HEADER frames
func TestParseHeaderFrame(t *testing.T) {
	// Create a valid HEADER frame payload
	var payload bytes.Buffer

	// ClassID (short) - BASIC class (60)
	binary.Write(&payload, binary.BigEndian, uint16(BASIC))

	// Weight (short) - must be 0
	binary.Write(&payload, binary.BigEndian, uint16(0))

	// Body size (long long)
	binary.Write(&payload, binary.BigEndian, uint64(100))

	// Property flags (short) - bit 15 set for contentType
	binary.Write(&payload, binary.BigEndian, uint16(0x8000))

	// Content type (short string)
	payload.Write(EncodeShortStr("text/plain"))

	channel := uint16(1)
	payloadSize := uint32(payload.Len())

	state, err := parseHeaderFrame(channel, payloadSize, payload.Bytes())
	if err != nil {
		t.Fatalf("parseHeaderFrame failed: %v", err)
	}

	if state.HeaderFrame == nil {
		t.Fatal("HeaderFrame is nil")
	}

	if state.HeaderFrame.ClassID != uint16(BASIC) {
		t.Errorf("Expected class ID %d, got %d", uint16(BASIC), state.HeaderFrame.ClassID)
	}

	if state.HeaderFrame.BodySize != 100 {
		t.Errorf("Expected body size 100, got %d", state.HeaderFrame.BodySize)
	}

	if state.HeaderFrame.Properties == nil {
		t.Fatal("Properties is nil")
	}

	if state.HeaderFrame.Properties.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", state.HeaderFrame.Properties.ContentType)
	}
}

// TestParseHeaderFrame_InvalidWeight tests that weight must be 0
func TestParseHeaderFrame_InvalidWeight(t *testing.T) {
	var payload bytes.Buffer

	// ClassID (short)
	binary.Write(&payload, binary.BigEndian, uint16(BASIC))

	// Weight (short) - invalid non-zero value
	binary.Write(&payload, binary.BigEndian, uint16(5))

	// Body size (long long)
	binary.Write(&payload, binary.BigEndian, uint64(100))

	channel := uint16(1)
	payloadSize := uint32(payload.Len())

	_, err := parseHeaderFrame(channel, payloadSize, payload.Bytes())
	if err == nil {
		t.Fatal("Expected error for non-zero weight, got nil")
	}

	expectedErr := "weight must be 0"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestParseBodyFrame tests parsing of BODY frames
func TestParseBodyFrame(t *testing.T) {
	bodyContent := []byte("Test message body")
	channel := uint16(2)
	payloadSize := uint32(len(bodyContent))

	state, err := parseBodyFrame(channel, payloadSize, bodyContent)
	if err != nil {
		t.Fatalf("parseBodyFrame failed: %v", err)
	}

	if !bytes.Equal(state.Body, bodyContent) {
		t.Errorf("Expected body %q, got %q", bodyContent, state.Body)
	}
}

// TestParseBodyFrame_TooShort tests error handling for body frames with short payload
func TestParseBodyFrame_TooShort(t *testing.T) {
	bodyContent := []byte("Short")
	channel := uint16(1)
	payloadSize := uint32(100) // Claims 100 bytes but only 5 provided

	_, err := parseBodyFrame(channel, payloadSize, bodyContent)
	if err == nil {
		t.Fatal("Expected error for payload too short, got nil")
	}
}

// TestParseMethodFrame_ConnectionClass tests parsing CONNECTION class methods
func TestParseMethodFrame_ConnectionClass(t *testing.T) {
	// Create a payload for Connection.Start-Ok
	var methodPayload bytes.Buffer

	// Client properties (table)
	clientProps := map[string]any{
		"product": "OtterMQ Test",
		"version": "1.0.0",
	}
	encodedProps := EncodeTable(clientProps)
	methodPayload.Write(EncodeLongStr(encodedProps))

	// Mechanism (short string)
	methodPayload.Write(EncodeShortStr("PLAIN"))

	// Response (long string)
	methodPayload.Write(EncodeLongStr([]byte("\x00test\x00password")))

	// Locale (short string)
	methodPayload.Write(EncodeShortStr("en_US"))

	// Create full payload with class and method IDs
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, uint16(CONNECTION))
	binary.Write(&payload, binary.BigEndian, uint16(CONNECTION_START_OK))
	payload.Write(methodPayload.Bytes())

	channel := uint16(0)
	state, err := parseMethodFrame(channel, payload.Bytes())
	if err != nil {
		t.Fatalf("parseMethodFrame failed: %v", err)
	}

	if state.MethodFrame == nil {
		t.Fatal("MethodFrame is nil")
	}

	if state.MethodFrame.Channel != channel {
		t.Errorf("Expected channel %d, got %d", channel, state.MethodFrame.Channel)
	}

	if state.MethodFrame.ClassID != uint16(CONNECTION) {
		t.Errorf("Expected class ID %d, got %d", uint16(CONNECTION), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(CONNECTION_START_OK) {
		t.Errorf("Expected method ID %d, got %d", uint16(CONNECTION_START_OK), state.MethodFrame.MethodID)
	}

	startOk, ok := state.MethodFrame.Content.(*ConnectionStartOk)
	if !ok {
		t.Fatalf("Expected *ConnectionStartOk, got %T", state.MethodFrame.Content)
	}

	if startOk.Mechanism != "PLAIN" {
		t.Errorf("Expected mechanism 'PLAIN', got '%s'", startOk.Mechanism)
	}

	if startOk.Locale != "en_US" {
		t.Errorf("Expected locale 'en_US', got '%s'", startOk.Locale)
	}
}

// TestParseMethodFrame_ChannelClass tests parsing CHANNEL class methods
func TestParseMethodFrame_ChannelClass(t *testing.T) {
	// Create a payload for Channel.Open
	var methodPayload bytes.Buffer

	// Reserved field (short string, empty)
	methodPayload.Write(EncodeShortStr(""))

	// Create full payload with class and method IDs
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, uint16(CHANNEL))
	binary.Write(&payload, binary.BigEndian, uint16(CHANNEL_OPEN))
	payload.Write(methodPayload.Bytes())

	channel := uint16(1)
	state, err := parseMethodFrame(channel, payload.Bytes())
	if err != nil {
		t.Fatalf("parseMethodFrame failed: %v", err)
	}

	if state.MethodFrame == nil {
		t.Fatal("MethodFrame is nil")
	}

	if state.MethodFrame.ClassID != uint16(CHANNEL) {
		t.Errorf("Expected class ID %d, got %d", uint16(CHANNEL), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(CHANNEL_OPEN) {
		t.Errorf("Expected method ID %d, got %d", uint16(CHANNEL_OPEN), state.MethodFrame.MethodID)
	}
}

// TestParseMethodFrame_QueueClass tests parsing QUEUE class methods
func TestParseMethodFrame_QueueClass(t *testing.T) {
	// Create a payload for Queue.Declare
	var methodPayload bytes.Buffer

	// Reserved (short)
	binary.Write(&methodPayload, binary.BigEndian, uint16(0))

	// Queue name (short string)
	methodPayload.Write(EncodeShortStr("test-queue"))

	// Flags (bits packed in a byte)
	// passive (bit 0), durable (bit 1), exclusive (bit 2), auto-delete (bit 3), no-wait (bit 4)
	flags := byte(0x02) // durable = true
	methodPayload.WriteByte(flags)

	// Arguments (table)
	args := map[string]any{}
	encodedArgs := EncodeTable(args)
	methodPayload.Write(EncodeLongStr(encodedArgs))

	// Create full payload with class and method IDs
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, uint16(QUEUE))
	binary.Write(&payload, binary.BigEndian, uint16(QUEUE_DECLARE))
	payload.Write(methodPayload.Bytes())

	channel := uint16(1)
	state, err := parseMethodFrame(channel, payload.Bytes())
	if err != nil {
		t.Fatalf("parseMethodFrame failed: %v", err)
	}

	if state.MethodFrame == nil {
		t.Fatal("MethodFrame is nil")
	}

	if state.MethodFrame.ClassID != uint16(QUEUE) {
		t.Errorf("Expected class ID %d, got %d", uint16(QUEUE), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(QUEUE_DECLARE) {
		t.Errorf("Expected method ID %d, got %d", uint16(QUEUE_DECLARE), state.MethodFrame.MethodID)
	}
}

// TestParseMethodFrame_ExchangeClass tests parsing EXCHANGE class methods
func TestParseMethodFrame_ExchangeClass(t *testing.T) {
	// Create a payload for Exchange.Declare
	var methodPayload bytes.Buffer

	// Reserved (short)
	binary.Write(&methodPayload, binary.BigEndian, uint16(0))

	// Exchange name (short string)
	methodPayload.Write(EncodeShortStr("test-exchange"))

	// Type (short string)
	methodPayload.Write(EncodeShortStr("direct"))

	// Flags (bits packed in a byte)
	// passive, durable, auto-delete, internal, no-wait
	flags := byte(0x02) // durable = true
	methodPayload.WriteByte(flags)

	// Arguments (table)
	args := map[string]any{}
	encodedArgs := EncodeTable(args)
	methodPayload.Write(EncodeLongStr(encodedArgs))

	// Create full payload with class and method IDs
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, uint16(EXCHANGE))
	binary.Write(&payload, binary.BigEndian, uint16(EXCHANGE_DECLARE))
	payload.Write(methodPayload.Bytes())

	channel := uint16(1)
	state, err := parseMethodFrame(channel, payload.Bytes())
	if err != nil {
		t.Fatalf("parseMethodFrame failed: %v", err)
	}

	if state.MethodFrame == nil {
		t.Fatal("MethodFrame is nil")
	}

	if state.MethodFrame.ClassID != uint16(EXCHANGE) {
		t.Errorf("Expected class ID %d, got %d", uint16(EXCHANGE), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(EXCHANGE_DECLARE) {
		t.Errorf("Expected method ID %d, got %d", uint16(EXCHANGE_DECLARE), state.MethodFrame.MethodID)
	}
}

// TestParseMethodFrame_BasicClass tests parsing BASIC class methods
func TestParseMethodFrame_BasicClass(t *testing.T) {
	// Create a payload for Basic.Publish
	var methodPayload bytes.Buffer

	// Reserved (short)
	binary.Write(&methodPayload, binary.BigEndian, uint16(0))

	// Exchange name (short string)
	methodPayload.Write(EncodeShortStr(""))

	// Routing key (short string)
	methodPayload.Write(EncodeShortStr("test-queue"))

	// Flags (bits: mandatory, immediate)
	flags := byte(0x00)
	methodPayload.WriteByte(flags)

	// Create full payload with class and method IDs
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, uint16(BASIC))
	binary.Write(&payload, binary.BigEndian, uint16(BASIC_PUBLISH))
	payload.Write(methodPayload.Bytes())

	channel := uint16(1)
	state, err := parseMethodFrame(channel, payload.Bytes())
	if err != nil {
		t.Fatalf("parseMethodFrame failed: %v", err)
	}

	if state.MethodFrame == nil {
		t.Fatal("MethodFrame is nil")
	}

	if state.MethodFrame.ClassID != uint16(BASIC) {
		t.Errorf("Expected class ID %d, got %d", uint16(BASIC), state.MethodFrame.ClassID)
	}

	if state.MethodFrame.MethodID != uint16(BASIC_PUBLISH) {
		t.Errorf("Expected method ID %d, got %d", uint16(BASIC_PUBLISH), state.MethodFrame.MethodID)
	}
}

// TestDecodeBasicHeaderFlags tests decoding of basic header property flags
func TestDecodeBasicHeaderFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint16
		expected []string
	}{
		{
			name:     "No flags set",
			flags:    0x0000,
			expected: []string{},
		},
		{
			name:     "Content type flag (bit 15)",
			flags:    0x8000,
			expected: []string{"contentType"},
		},
		{
			name:     "Content encoding flag (bit 14)",
			flags:    0x4000,
			expected: []string{"contentEncoding"},
		},
		{
			name:     "Multiple flags",
			flags:    0xC000, // bits 15 and 14
			expected: []string{"contentType", "contentEncoding"},
		},
		{
			name:     "Delivery mode flag (bit 12)",
			flags:    0x1000,
			expected: []string{"deliveryMode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeBasicHeaderFlags(tt.flags)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d flags, got %d", len(tt.expected), len(result))
				return
			}

			// Check each expected flag is present
			for _, expectedFlag := range tt.expected {
				found := false
				for _, resultFlag := range result {
					if resultFlag == expectedFlag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected flag '%s' not found in result", expectedFlag)
				}
			}
		})
	}
}

// TestCreateContentPropertiesTable tests content properties parsing
func TestCreateContentPropertiesTable(t *testing.T) {
	// Test with content type
	var buf bytes.Buffer
	buf.Write(EncodeShortStr("application/json"))

	flags := []string{"contentType"}
	props, err := createContentPropertiesTable(flags, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("createContentPropertiesTable failed: %v", err)
	}

	if props.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", props.ContentType)
	}
}

// TestCreateContentPropertiesTable_DeliveryMode tests delivery mode validation
func TestCreateContentPropertiesTable_DeliveryMode(t *testing.T) {
	tests := []struct {
		name         string
		deliveryMode uint8
		shouldError  bool
	}{
		{"Valid non-persistent", 1, false},
		{"Valid persistent", 2, false},
		{"Invalid delivery mode 0", 0, true},
		{"Invalid delivery mode 3", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteByte(tt.deliveryMode)

			flags := []string{"deliveryMode"}
			props, err := createContentPropertiesTable(flags, bytes.NewReader(buf.Bytes()))

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error for invalid delivery mode, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if props.DeliveryMode != tt.deliveryMode {
					t.Errorf("Expected delivery mode %d, got %d", tt.deliveryMode, props.DeliveryMode)
				}
			}
		})
	}
}

// TestCreateContentPropertiesTable_Priority tests priority validation
func TestCreateContentPropertiesTable_Priority(t *testing.T) {
	tests := []struct {
		name        string
		priority    uint8
		shouldError bool
	}{
		{"Valid priority 0", 0, false},
		{"Valid priority 5", 5, false},
		{"Valid priority 9", 9, false},
		{"Invalid priority 10", 10, true},
		{"Invalid priority 255", 255, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteByte(tt.priority)

			flags := []string{"priority"}
			props, err := createContentPropertiesTable(flags, bytes.NewReader(buf.Bytes()))

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error for invalid priority, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if props.Priority != tt.priority {
					t.Errorf("Expected priority %d, got %d", tt.priority, props.Priority)
				}
			}
		})
	}
}
