package connection

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
			result := formatHeader(tt.frameType, tt.channel, tt.payloadSize)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("formatHeader(%d, %d, %d) = %x; want %x",
					tt.frameType, tt.channel, tt.payloadSize, result, tt.expected)
			}
		})
	}

	return
}

// TestCreateConnectionStartOkFrame tests the createConnectionStartOkFrame function
func TestCreateConnectionStartOkFrame(t *testing.T) {
	var expected []byte
	header := []byte{
		1, 0, 0, 0, 0, 0, 125, // Frame header
	}
	payload := []byte{
		0, 9, // version-major, version-minor
		7, 'p', 'r', 'o', 'd', 'u', 'c', 't', 'S', 0, 0, 0, 7, 'O', 't', 't', 'e', 'r', 'M', 'Q', // Seerver properties
		7, 'v', 'e', 'r', 's', 'i', 'o', 'n', 'S', 0, 0, 0, 5, '0', '.', '1', '.', '0',
		8, 'p', 'l', 'a', 't', 'f', 'o', 'r', 'm', 'S', 0, 0, 0, 5, 'L', 'i', 'n', 'u', 'x',
		11, 'i', 'n', 'f', 'o', 'r', 'm', 'a', 't', 'i', 'o', 'n', 'S', 0, 0, 0, 38,
		'h', 't', 't', 'p', 's', ':', '/', '/', 'g', 'i', 't', 'h', 'u', 'b', '.', 'c', 'o', 'm', '/',
		'a', 'n', 'd', 'r', 'e', 'l', 'c', 'u', 'n', 'h', 'a', '/', 'o', 't', 't', 'e', 'r', 'm', 'q', 'P', 'L', 'A', 'I', 'N', 0, 'e', 'n', '_', 'U', 'S', // Mechanisms and Locales
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
		t.Errorf("createConnectionStartOkFrame() = %x; want %x", frame, expected)
	}
}
