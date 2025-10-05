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

		case 'F':
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

func DecodeExchangeDeclareFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"passive", "durable", "autoDelete", "internal", "noWait", "flag6", "flag7", "flag8"}

	for i := range 8 {
		flags[flagNames[i]] = (octet & (1 << uint(i))) != 0
	}

	return flags
}

func DecodeExchangeDeleteFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"ifUnused", "noWait", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}

func DecodeQueueDeclareFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"passive", "durable", "ifUnused", "exclusive", "noWait", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}

func DecodeQueueDeleteFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"ifUnused", "noWait", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
}

func DecodeQueueBindFlags(octet byte) map[string]bool {
	flags := make(map[string]bool)
	flagNames := []string{"noWait", "flag2", "flag3", "flag4", "flag5", "flag6", "flag7", "flag8"}

	for i := 0; i < 8; i++ {
		flags[flagNames[i]] = (octet & (1 << uint(7-i))) != 0
	}

	return flags
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
