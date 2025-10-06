package amqp

import (
	"bytes"
	"testing"
)

func TestDecodeFlags_QueueOrder(t *testing.T) {
	// passive, durable, exclusive, autoDelete, noWait
	flagNames := []string{"passive", "durable", "exclusive", "autoDelete", "noWait"}
	// 0b00011110: durable, exclusive, autoDelete, noWait set
	octet := byte(0x1E)
	flags := DecodeFlags(octet, flagNames, true)
	if flags["passive"] {
		t.Errorf("Expected passive=false, got true")
	}
	if !flags["durable"] {
		t.Errorf("Expected durable=true, got false")
	}
	if !flags["exclusive"] {
		t.Errorf("Expected exclusive=true, got false")
	}
	if !flags["autoDelete"] {
		t.Errorf("Expected autoDelete=true, got false")
	}
	if !flags["noWait"] {
		t.Errorf("Expected noWait=true, got false")
	}
}

func TestDecodeFlags_ExchangeOrder(t *testing.T) {
	// passive, durable, autoDelete, internal, noWait
	flagNames := []string{"passive", "durable", "autoDelete", "internal", "noWait"}
	// 0b00011110: durable, autoDelete, internal, noWait set
	octet := byte(0x1E)
	flags := DecodeFlags(octet, flagNames, true)
	if flags["passive"] {
		t.Errorf("Expected passive=false, got true")
	}
	if !flags["durable"] {
		t.Errorf("Expected durable=true, got false")
	}
	if !flags["autoDelete"] {
		t.Errorf("Expected autoDelete=true, got false")
	}
	if !flags["internal"] {
		t.Errorf("Expected internal=true, got false")
	}
	if !flags["noWait"] {
		t.Errorf("Expected noWait=true, got false")
	}
}

func TestDecodeFlags_AllFalse(t *testing.T) {
	flagNames := []string{"passive", "durable", "exclusive", "autoDelete", "noWait"}
	octet := byte(0x00)
	flags := DecodeFlags(octet, flagNames, true)
	for _, name := range flagNames {
		if flags[name] {
			t.Errorf("Expected %s=false, got true", name)
		}
	}
}

func TestDecodeFlags_AllTrue(t *testing.T) {
	flagNames := []string{"passive", "durable", "exclusive", "autoDelete", "noWait"}
	octet := byte(0x1F)
	flags := DecodeFlags(octet, flagNames, true)
	for _, name := range flagNames {
		if !flags[name] {
			t.Errorf("Expected %s=true, got false", name)
		}
	}
}

func TestParseQueueDeclareFrame_Flags(t *testing.T) {
	// passive, durable, exclusive, autoDelete, noWait
	var payload bytes.Buffer
	payload.Write([]byte{0, 0}) // reserved
	payload.WriteByte(4)        // queue name length
	payload.WriteString("test")
	payload.WriteByte(0x1E)           // flags: durable, exclusive, autoDelete, noWait
	payload.Write([]byte{0, 0, 0, 0}) // arguments (empty longstr)

	request, err := parseQueueDeclareFrame(payload.Bytes())
	if err != nil {
		t.Fatalf("parseQueueDeclareFrame failed: %v", err)
	}
	msg, ok := request.Content.(*QueueDeclareMessage)
	if !ok {
		t.Fatalf("Expected *QueueDeclareMessage, got %T", request.Content)
	}
	if msg.Passive {
		t.Errorf("Expected passive=false, got true")
	}
	if !msg.Durable {
		t.Errorf("Expected durable=true, got false")
	}
	if !msg.Exclusive {
		t.Errorf("Expected exclusive=true, got false")
	}
	if !msg.AutoDelete {
		t.Errorf("Expected autoDelete=true, got false")
	}
	if !msg.NoWait {
		t.Errorf("Expected noWait=true, got false")
	}
}

func TestParseExchangeDeclareFrame_Flags(t *testing.T) {
	// passive, durable, autoDelete, internal, noWait
	var payload bytes.Buffer
	payload.Write([]byte{0, 0}) // reserved
	payload.WriteByte(7)        // exchange name length
	payload.WriteString("exchng1")
	payload.WriteByte(6) // type length
	payload.WriteString("direct")
	payload.WriteByte(0x1E)           // flags: durable, autoDelete, internal, noWait
	payload.Write([]byte{0, 0, 0, 0}) // arguments (empty longstr)

	request, err := parseExchangeDeclareFrame(payload.Bytes())
	if err != nil {
		t.Fatalf("parseExchangeDeclareFrame failed: %v", err)
	}
	msg, ok := request.Content.(*ExchangeDeclareMessage)
	if !ok {
		t.Fatalf("Expected *ExchangeDeclareMessage, got %T", request.Content)
	}
	if msg.Passive {
		t.Errorf("Expected passive=false, got true")
	}
	if !msg.Durable {
		t.Errorf("Expected durable=true, got false")
	}
	if !msg.AutoDelete {
		t.Errorf("Expected autoDelete=true, got false")
	}
	if !msg.Internal {
		t.Errorf("Expected internal=true, got false")
	}
	if !msg.NoWait {
		t.Errorf("Expected noWait=true, got false")
	}
}
