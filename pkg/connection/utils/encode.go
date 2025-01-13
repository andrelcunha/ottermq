package utils

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// encodeTable encodes a proper AMQP field table
func EncodeTable(table map[string]interface{}) []byte {
	var buf bytes.Buffer

	for key, value := range table {
		// Field name
		buf.Write(EncodeShortStr(key))

		// Field value type and value
		switch v := value.(type) {
		case string:
			buf.WriteByte('S') // Field value type 'S' (string)
			buf.Write(EncodeLongStr([]byte(v)))

		case int:
			buf.WriteByte('I') // Field value type 'I' (int)
			binary.Write(&buf, binary.BigEndian, int32(v))

		// In the case map[string]interface:
		case map[string]interface{}:
			// Recursively encode the nested map
			buf.WriteByte('F') // Field value type 'F' (field table)
			encodedTable := EncodeTable(v)
			buf.Write(EncodeLongStr(encodedTable))

		case bool:
			buf.WriteByte('t')
			if v {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}

		default:
			buf.WriteByte('U')
		}
	}
	return buf.Bytes()
}

func EncodeLongStr(data []byte) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

func EncodeShortStr(data string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(len(data)))
	buf.WriteString(data)
	return buf.Bytes()
}

func EncodeSecurityPlain(securityStr string) []byte {
	// Concatenate username, null byte, and password
	// securityStr := username + "\x00" + password
	// Replace spaces with null bytes
	encodedStr := strings.ReplaceAll(securityStr, " ", "\x00")
	// Encode length as a uint32 and append the encoded string
	length := uint32(len(encodedStr))
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, length)
	buf.WriteString(encodedStr)
	return buf.Bytes()
}
