package amqp

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestParseBasicConsumeFrame(t *testing.T) {
	tests := []struct {
		name           string
		payload        []byte
		expectedQueue  string
		expectedTag    string
		expectedNoLocal bool
		expectedNoAck  bool
		expectedExclusive bool
		expectedNoWait bool
		expectedArgs   map[string]any
		shouldError    bool
		errorMsg       string
	}{
		{
			name: "Valid basic consume with empty queue and tag",
			payload: buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "",
				ConsumerTag: "",
				NoLocal:     false,
				NoAck:       false,
				Exclusive:   false,
				NoWait:      false,
				Arguments:   nil,
			}),
			expectedQueue:     "",
			expectedTag:       "",
			expectedNoLocal:   false,
			expectedNoAck:     false,
			expectedExclusive: false,
			expectedNoWait:    false,
			expectedArgs:      nil,
			shouldError:       false,
		},
		{
			name: "Valid basic consume with queue and consumer tag",
			payload: buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "test-queue",
				ConsumerTag: "consumer-1",
				NoLocal:     false,
				NoAck:       true,
				Exclusive:   false,
				NoWait:      false,
				Arguments:   nil,
			}),
			expectedQueue:     "test-queue",
			expectedTag:       "consumer-1",
			expectedNoLocal:   false,
			expectedNoAck:     true,
			expectedExclusive: false,
			expectedNoWait:    false,
			expectedArgs:      nil,
			shouldError:       false,
		},
		{
			name: "Valid basic consume with all flags set",
			payload: buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "exclusive-queue",
				ConsumerTag: "exclusive-consumer",
				NoLocal:     true,
				NoAck:       true,
				Exclusive:   true,
				NoWait:      true,
				Arguments:   nil,
			}),
			expectedQueue:     "exclusive-queue",
			expectedTag:       "exclusive-consumer",
			expectedNoLocal:   true,
			expectedNoAck:     true,
			expectedExclusive: true,
			expectedNoWait:    true,
			expectedArgs:      nil,
			shouldError:       false,
		},
		{
			name: "Valid basic consume with arguments",
			payload: buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "queue-with-args",
				ConsumerTag: "consumer-with-args",
				NoLocal:     false,
				NoAck:       false,
				Exclusive:   false,
				NoWait:      false,
				Arguments: map[string]any{
					"x-priority": int32(10),
					"x-message-ttl": int32(60000),
				},
			}),
			expectedQueue:     "queue-with-args",
			expectedTag:       "consumer-with-args",
			expectedNoLocal:   false,
			expectedNoAck:     false,
			expectedExclusive: false,
			expectedNoWait:    false,
			expectedArgs: map[string]any{
				"x-priority": int32(10),
				"x-message-ttl": int32(60000),
			},
			shouldError: false,
		},
		{
			name: "Valid basic consume with complex arguments",
			payload: buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "complex-queue",
				ConsumerTag: "complex-consumer",
				NoLocal:     false,
				NoAck:       false,
				Exclusive:   false,
				NoWait:      false,
				Arguments: map[string]any{
					"x-prefetch-count": int32(100),
					"x-consumer-timeout": int32(30000),
					"x-cancel-on-ha-failover": true,
				},
			}),
			expectedQueue:     "complex-queue",
			expectedTag:       "complex-consumer",
			expectedNoLocal:   false,
			expectedNoAck:     false,
			expectedExclusive: false,
			expectedNoWait:    false,
			expectedArgs: map[string]any{
				"x-prefetch-count": int32(100),
				"x-consumer-timeout": int32(30000),
				"x-cancel-on-ha-failover": true,
			},
			shouldError: false,
		},
		{
			name:        "Payload too short",
			payload:     []byte{0x00, 0x00, 0x01, 'q'}, // Only 4 bytes
			shouldError: true,
			errorMsg:    "payload too short",
		},
		{
			name:        "Empty payload",
			payload:     []byte{},
			shouldError: true,
			errorMsg:    "payload too short",
		},
		{
			name: "Minimum valid payload without arguments",
			payload: buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "",
				ConsumerTag: "",
				NoLocal:     false,
				NoAck:       false,
				Exclusive:   false,
				NoWait:      false,
				Arguments:   nil,
			}),
			expectedQueue:     "",
			expectedTag:       "",
			expectedNoLocal:   false,
			expectedNoAck:     false,
			expectedExclusive: false,
			expectedNoWait:    false,
			expectedArgs:      nil,
			shouldError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBasicConsumeFrame(tt.payload)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			content, ok := result.Content.(*BasicConsumeContent)
			if !ok {
				t.Errorf("Expected BasicConsumeContent, got %T", result.Content)
				return
			}

			// Validate all fields
			if content.Queue != tt.expectedQueue {
				t.Errorf("Expected queue '%s', got '%s'", tt.expectedQueue, content.Queue)
			}
			if content.ConsumerTag != tt.expectedTag {
				t.Errorf("Expected consumer tag '%s', got '%s'", tt.expectedTag, content.ConsumerTag)
			}
			if content.NoLocal != tt.expectedNoLocal {
				t.Errorf("Expected NoLocal %t, got %t", tt.expectedNoLocal, content.NoLocal)
			}
			if content.NoAck != tt.expectedNoAck {
				t.Errorf("Expected NoAck %t, got %t", tt.expectedNoAck, content.NoAck)
			}
			if content.Exclusive != tt.expectedExclusive {
				t.Errorf("Expected Exclusive %t, got %t", tt.expectedExclusive, content.Exclusive)
			}
			if content.NoWait != tt.expectedNoWait {
				t.Errorf("Expected NoWait %t, got %t", tt.expectedNoWait, content.NoWait)
			}

			// Validate arguments
			if tt.expectedArgs == nil {
				if content.Arguments != nil {
					t.Errorf("Expected nil arguments, got %v", content.Arguments)
				}
			} else {
				if content.Arguments == nil {
					t.Error("Expected arguments but got nil")
					return
				}
				for key, expectedValue := range tt.expectedArgs {
					actualValue, exists := content.Arguments[key]
					if !exists {
						t.Errorf("Expected argument '%s' not found", key)
						continue
					}
					if actualValue != expectedValue {
						t.Errorf("Expected argument '%s' value %v, got %v", key, expectedValue, actualValue)
					}
				}
				// Check for unexpected arguments
				for key := range content.Arguments {
					if _, exists := tt.expectedArgs[key]; !exists {
						t.Errorf("Unexpected argument '%s' found", key)
					}
				}
			}
		})
	}
}

// Helper struct for building test payloads
type BasicConsumeParams struct {
	Queue       string
	ConsumerTag string
	NoLocal     bool
	NoAck       bool
	Exclusive   bool
	NoWait      bool
	Arguments   map[string]any
}

// Helper function to build BASIC_CONSUME payload
func buildBasicConsumePayload(t *testing.T, params BasicConsumeParams) []byte {
	var buf bytes.Buffer

	// reserved1 (short int)
	if err := binary.Write(&buf, binary.BigEndian, uint16(0)); err != nil {
		t.Fatalf("Failed to write reserved1: %v", err)
	}

	// queue (shortstr)
	if err := buf.WriteByte(byte(len(params.Queue))); err != nil {
		t.Fatalf("Failed to write queue length: %v", err)
	}
	if _, err := buf.WriteString(params.Queue); err != nil {
		t.Fatalf("Failed to write queue: %v", err)
	}

	// consumer-tag (shortstr)
	if err := buf.WriteByte(byte(len(params.ConsumerTag))); err != nil {
		t.Fatalf("Failed to write consumer tag length: %v", err)
	}
	if _, err := buf.WriteString(params.ConsumerTag); err != nil {
		t.Fatalf("Failed to write consumer tag: %v", err)
	}

	// flags (octet)
	var flags uint8
	if params.NoLocal {
		flags |= 0x01 // bit 0
	}
	if params.NoAck {
		flags |= 0x02 // bit 1  
	}
	if params.Exclusive {
		flags |= 0x04 // bit 2
	}
	if params.NoWait {
		flags |= 0x08 // bit 3
	}
	if err := buf.WriteByte(flags); err != nil {
		t.Fatalf("Failed to write flags: %v", err)
	}

	// arguments (table encoded as longstr)
	if params.Arguments != nil {
		tableData := EncodeTable(params.Arguments)
		
		// Write table length (4 bytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(tableData))); err != nil {
			t.Fatalf("Failed to write arguments length: %v", err)
		}
		
		// Write table data
		if _, err := buf.Write(tableData); err != nil {
			t.Fatalf("Failed to write arguments data: %v", err)
		}
	}
	// If Arguments is nil, don't write anything for the table

	return buf.Bytes()
}

func TestParseBasicConsumeFrame_EdgeCases(t *testing.T) {
	t.Run("Maximum length queue name", func(t *testing.T) {
		longQueue := string(make([]byte, 255)) // Maximum shortstr length
		for i := range longQueue {
			longQueue = longQueue[:i] + "q" + longQueue[i+1:]
		}
		
		payload := buildBasicConsumePayload(t, BasicConsumeParams{
			Queue:       longQueue,
			ConsumerTag: "consumer",
			Arguments:   nil,
		})

		result, err := parseBasicConsumeFrame(payload)
		if err != nil {
			t.Errorf("Unexpected error with max length queue: %v", err)
			return
		}

		content := result.Content.(*BasicConsumeContent)
		if content.Queue != longQueue {
			t.Error("Queue name was not preserved correctly")
		}
	})

	t.Run("Maximum length consumer tag", func(t *testing.T) {
		longTag := string(make([]byte, 255)) // Maximum shortstr length
		for i := range longTag {
			longTag = longTag[:i] + "c" + longTag[i+1:]
		}

		payload := buildBasicConsumePayload(t, BasicConsumeParams{
			Queue:       "test",
			ConsumerTag: longTag,
			Arguments:   nil,
		})

		result, err := parseBasicConsumeFrame(payload)
		if err != nil {
			t.Errorf("Unexpected error with max length consumer tag: %v", err)
			return
		}

		content := result.Content.(*BasicConsumeContent)
		if content.ConsumerTag != longTag {
			t.Error("Consumer tag was not preserved correctly")
		}
	})

	t.Run("All flags combinations", func(t *testing.T) {
		// Test all 16 possible combinations of the 4 flags
		for i := 0; i < 16; i++ {
			noLocal := (i & 8) != 0
			noAck := (i & 4) != 0
			exclusive := (i & 2) != 0
			noWait := (i & 1) != 0

			payload := buildBasicConsumePayload(t, BasicConsumeParams{
				Queue:       "test",
				ConsumerTag: "consumer",
				NoLocal:     noLocal,
				NoAck:       noAck,
				Exclusive:   exclusive,
				NoWait:      noWait,
				Arguments:   nil,
			})

			result, err := parseBasicConsumeFrame(payload)
			if err != nil {
				t.Errorf("Error with flags combination %d: %v", i, err)
				continue
			}

			content := result.Content.(*BasicConsumeContent)
			if content.NoLocal != noLocal ||
				content.NoAck != noAck ||
				content.Exclusive != exclusive ||
				content.NoWait != noWait {
				t.Errorf("Flags not parsed correctly for combination %d", i)
			}
		}
	})
}

func TestParseBasicConsumeFrame_ArgumentsEdgeCases(t *testing.T) {
	t.Run("Empty arguments table", func(t *testing.T) {
		payload := buildBasicConsumePayload(t, BasicConsumeParams{
			Queue:       "test",
			ConsumerTag: "consumer",
			Arguments:   map[string]any{}, // Empty but not nil
		})

		result, err := parseBasicConsumeFrame(payload)
		if err != nil {
			t.Errorf("Unexpected error with empty arguments: %v", err)
			return
		}

		content := result.Content.(*BasicConsumeContent)
		if content.Arguments == nil {
			t.Error("Expected empty map, got nil")
		}
	})

	t.Run("Arguments with various types", func(t *testing.T) {
		args := map[string]any{
			"string-arg": "test-value",
			"int-arg":    int32(42),
			"bool-arg":   true,
		}

		payload := buildBasicConsumePayload(t, BasicConsumeParams{
			Queue:       "test",
			ConsumerTag: "consumer",
			Arguments:   args,
		})

		result, err := parseBasicConsumeFrame(payload)
		if err != nil {
			t.Errorf("Unexpected error with mixed arguments: %v", err)
			return
		}

		content := result.Content.(*BasicConsumeContent)
		for key, expectedValue := range args {
			if actualValue, exists := content.Arguments[key]; !exists {
				t.Errorf("Argument '%s' not found", key)
			} else if actualValue != expectedValue {
				t.Errorf("Argument '%s': expected %v, got %v", key, expectedValue, actualValue)
			}
		}
	})
}