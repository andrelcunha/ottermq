package shared

import (
	"bytes"
	"testing"

	. "github.com/andrelcunha/ottermq/pkg/connection/utils"
)

func TestEncodeDecodeTable(t *testing.T) {
	originalTable := map[string]interface{}{
		"product":     "OtterMQ",
		"version":     "0.6.0-alpha",
		"platform":    "Linux",
		"information": "https://github.com/andrelcunha/ottermq",
	}

	encodedTable := EncodeTable(originalTable)
	decodedTable, err := DecodeTable(encodedTable)
	if err != nil {
		t.Fatalf("Error decoding table: %v", err)
	}

	if !compareTables(originalTable, decodedTable) {
		t.Fatalf("Tables do not match: original %v, decoded %v", originalTable, decodedTable)
	}
}

func compareTables(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		if valueB, ok := b[key]; !ok || valueA != valueB {
			return false
		}
	}

	return true
}

// TestFormatHeader tests the formatHeader function
func TestFormatHeader(t *testing.T) {
	tests := []struct {
		frameType   byte
		channel     uint16
		payloadSize uint32
		expected    []byte
	}{
		{
			frameType:   1,
			channel:     0,
			payloadSize: 83,
			expected:    []byte{1, 0, 0, 0, 0, 0, 83},
		},
		{
			frameType:   2,
			channel:     1,
			payloadSize: 256,
			expected:    []byte{2, 0, 1, 0, 0, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run("Testing formatHeader", func(t *testing.T) {
			result := FormatHeader(tt.frameType, tt.channel, tt.payloadSize)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("formatHeader(%d, %d, %d) = %x; want %x",
					tt.frameType, tt.channel, tt.payloadSize, result, tt.expected)
			}
		})
	}
}
