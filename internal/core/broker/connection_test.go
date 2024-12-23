package broker

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/andrelcunha/ottermq/pkg/connection/shared"
)

func TestEncodeDecodeTable(t *testing.T) {
	originalTable := map[string]interface{}{
		"product":     "OtterMQ",
		"version":     "0.6.0-alpha",
		"platform":    "Linux",
		"information": "https://github.com/andrelcunha/ottermq",
	}

	encodedTable := shared.EncodeTable(originalTable)
	decodedTable, err := shared.DecodeTable(encodedTable)
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
			result := shared.FormatHeader(tt.frameType, tt.channel, tt.payloadSize)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("formatHeader(%d, %d, %d) = %x; want %x",
					tt.frameType, tt.channel, tt.payloadSize, result, tt.expected)
			}
		})
	}

	return
}

// TestCreateConnectionStartFrame tests the createConnectionStartOkFrame function
func TestCreateConnectionStartFrame(t *testing.T) {
	var expected []byte
	header := []byte{
		1, 0, 0, 0, 0, 0, 48, // Frame header
	}
	payload := []byte{
		0, 10, 0, 10, // class, method
		0, 9, // version-major, version-minor
		0, 0, 0, 20, // Size of the server properties table
		// Server properties
		7, 'p', 'r', 'o', 'd', 'u', 'c', 't', 'S', 0, 0, 0, 7, 'O', 't', 't', 'e', 'r', 'M', 'Q',
		0, 0, 0, 5, // Size of mechanisms
		'P', 'L', 'A', 'I', 'N', // Mechanisms
		0, 0, 0, 5, // Size of locales
		'e', 'n', '_', 'U', 'S', // Mechanisms
	}
	frame_end := []byte{
		0xCE, // frame-end
	}
	fmt.Printf("len(header): %d, len(payload): %d, len(frame_end): %d \n", len(header), len(payload), len(frame_end))
	expected = append(expected, header...)
	expected = append(expected, payload...)
	expected = append(expected, frame_end...)

	frame := createConnectionStartFrame()

	if !bytes.Equal(frame, expected) {
		t.Errorf("createConnectionStartFrame() = %x; want %x", frame, expected)
	}
}
