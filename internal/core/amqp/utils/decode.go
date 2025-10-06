package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog/log"
)

// decodeTable decodes an AMQP field table from a byte slice
func DecodeTable(data []byte) (map[string]interface{}, error) {

	table := make(map[string]interface{})
	buf := bytes.NewReader(data)

	for buf.Len() > 0 {
		// Read field name
		fieldNameLength, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		fieldName := make([]byte, fieldNameLength)
		_, err = buf.Read(fieldName)
		if err != nil {
			return nil, err
		}

		// Read field value type
		fieldType, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		// Read field value based on the type
		switch fieldType {
		case 'S': // String
			var strLength uint32
			if err := binary.Read(buf, binary.BigEndian, &strLength); err != nil {
				return nil, err
			}

			strValue := make([]byte, strLength)
			_, err := buf.Read(strValue)
			if err != nil {
				return nil, err
			}

			table[string(fieldName)] = string(strValue)

		case 'I': // Integer (simplified, normally long-int should be used)
			var intValue int32
			if err := binary.Read(buf, binary.BigEndian, &intValue); err != nil {
				return nil, err
			}

			table[string(fieldName)] = intValue

		case 'F': // Field Table
			var strLength uint32
			if err := binary.Read(buf, binary.BigEndian, &strLength); err != nil {
				return nil, err
			}

			strValue := make([]byte, strLength)
			_, err := buf.Read(strValue)
			if err != nil {
				return nil, err
			}

			value, err := DecodeTable(strValue)
			if err != nil {
				return nil, err
			}
			table[string(fieldName)] = value

		case 't':
			value, err := DecodeBoolean(buf)
			if err != nil {
				return nil, err
			}
			table[string(fieldName)] = value

		// Add cases for other types as needed
		case 'V': // Void (null)
			// No payload, just store nil or skip
			table[string(fieldName)] = nil

		default:
			return nil, fmt.Errorf("unknown field type: %c", fieldType)
		}
	}
	return table, nil
}

// DecodeTimestamp reads and decodes a 64-bit POSIX timestamp from a bytes.Reader
func DecodeTimestamp(buf *bytes.Reader) (time.Time, error) {
	var timestamp int64
	err := binary.Read(buf, binary.BigEndian, &timestamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to decode timestamp: %v", err)
	}
	return time.Unix(timestamp, 0), nil
}

func DecodeLongStr(buf *bytes.Reader) (string, error) {
	var strLen uint32
	err := binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return "", err
	}

	strData := make([]byte, strLen)
	_, err = buf.Read(strData)
	if err != nil {
		return "", err
	}

	return string(strData), nil
}

func DecodeShortStr(buf *bytes.Reader) (string, error) {
	var strLen uint8
	err := binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return "", err
	}

	strData := make([]byte, strLen)
	_, err = buf.Read(strData)
	if err != nil {
		return "", err
	}

	return string(strData), nil
}

func DecodeShortInt(buf *bytes.Reader) (uint16, error) {
	var value uint16
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func DecodeLongInt(buf *bytes.Reader) (uint32, error) {
	var value uint32
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func DecodeLongLongInt(buf *bytes.Reader) (uint64, error) {
	var value uint64
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func DecodeBoolean(buf *bytes.Reader) (bool, error) {
	var value uint8
	err := binary.Read(buf, binary.BigEndian, &value)
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

// DecodeFlags decodes an octet into a map of flag names to boolean values.
func DecodeFlags(octet byte, flagNames []string, lsbFirst bool) map[string]bool {
	if len(flagNames) > 8 {
		log.Warn().Msg("More than 8 flag names provided; extra names will be ignored")
	}
	copiedFlagNames := make([]string, min(len(flagNames), 8))
	copy(copiedFlagNames, flagNames)
	flags := make(map[string]bool)
	for i := range copiedFlagNames {
		bit := i
		if !lsbFirst {
			bit = 7 - i
		}
		flags[copiedFlagNames[i]] = (octet & (1 << uint(bit))) != 0
	}
	return flags
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DecodeSecurityPlain(buf *bytes.Reader) (string, error) {
	var strLen uint32
	err := binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return "", err
	}

	if uint32(buf.Len()) < strLen {
		log.Debug().Int("buf_len", buf.Len()).Msg("Reached EOF")
		return "", io.EOF
	}

	// Read each byte and replace 0x00 with a space
	strData := make([]byte, strLen)
	for i := uint32(0); i < strLen; i++ {
		b, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if b == 0x00 {
			strData[i] = ' '
		} else {
			strData[i] = b
		}
	}

	return string(strData), nil
}
