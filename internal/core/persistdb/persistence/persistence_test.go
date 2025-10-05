package persistence

import (
	"os"
	"reflect"
	"testing"
)

func TestSafeVHostName(t *testing.T) {
	cases := map[string]string{
		"/":          "%2F",
		"vhost":      "vhost",
		"vhost/test": "vhost%2Ftest",
	}
	for input, expected := range cases {
		got := safeVHostName(input)
		if got != expected {
			t.Errorf("safeVHostName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestSaveLoadExchange(t *testing.T) {
	db := NewDefaultPersistence()
	vhostName := "/"
	props := ExchangePropertiesDb{
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Arguments:  nil,
	}
	exchange := &PersistedExchange{
		Name:       "test-exchange",
		Type:       "direct",
		Properties: props,
		Bindings:   []PersistedBinding{{QueueName: "q1", RoutingKey: "rk", Arguments: nil}},
	}
	err := db.SaveExchange(vhostName, exchange)
	if err != nil {
		t.Fatalf("SaveExchange failed: %v", err)
	}
	loaded, err := db.LoadExchange(vhostName, exchange.Name)
	if err != nil {
		t.Fatalf("LoadExchange failed: %v", err)
	}
	if !reflect.DeepEqual(exchange, loaded) {
		t.Errorf("Loaded exchange does not match saved.\nSaved: %+v\nLoaded: %+v", exchange, loaded)
	}
	// Cleanup
	safeName := safeVHostName(vhostName)
	os.RemoveAll("data/vhosts/" + safeName)
}

func TestSaveLoadQueue(t *testing.T) {
	db := NewDefaultPersistence()
	vhostName := "vhost/test"
	props := QueuePropertiesDb{
		Passive:    false,
		Durable:    true,
		Exclusive:  false,
		AutoDelete: false,
		NoWait:     false,
		Arguments:  nil,
	}
	msg := PersistedMessage{ID: "m1", Body: []byte("hello"), Properties: MessageProperties{DeliveryMode: 2}}
	queue := &PersistedQueue{
		Name:       "test-queue",
		Properties: props,
		Messages:   []PersistedMessage{msg},
	}
	err := db.SaveQueue(vhostName, queue)
	if err != nil {
		t.Fatalf("SaveQueue failed: %v", err)
	}
	loaded, err := db.LoadQueue(vhostName, queue.Name)
	if err != nil {
		t.Fatalf("LoadQueue failed: %v", err)
	}
	if !reflect.DeepEqual(queue, loaded) {
		t.Errorf("Loaded queue does not match saved.\nSaved: %+v\nLoaded: %+v", queue, loaded)
	}
	// Cleanup
	safeName := safeVHostName(vhostName)
	os.RemoveAll("data/vhosts/" + safeName)
}
