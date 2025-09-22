package persistdb

// import (
// 	"testing"

// 	"github.com/andrelcunha/ottermq/config"
// )

// var conf = &config.Config{
// 	Port:              "5672",
// 	Host:              "localhost",
// 	HeartBeatInterval: 5,
// }

// func TestSaveMessage(t *testing.T) {

// b := NewBroker(conf)
// queueName := "testQueue"
// msg := Message{
// 	ID:      "testMessage1",
// 	Content: "Hello, world!",
// }
// err := b.saveMessage(queueName, msg)
// if err != nil {
// 	t.Fatalf("Failed to save message: %v", err)
// }

// // Verify the message file exists
// file := filepath.Join("data", "queues", queueName, "testMessage1.json")
// if _, err := os.Stat(file); os.IsNotExist(err) {
// 	t.Fatalf("Expected message file %s to exist", file)
// }
// }

// func TestLoadMessages(t *testing.T) {
// // Remove the data/queues directory to ensure a clean state
// err := os.RemoveAll(filepath.Join("data", "queues"))
// if err != nil {
// 	t.Fatalf("Failed to remove data/queues directory: %v", err)
// }

// b := NewBroker(conf)
// queueName := "testQueue"
// msg := Message{
// 	ID:      "testMessage1",
// 	Content: "Hello, world!",
// }
// // Save a message to load later
// b.saveMessage(queueName, msg)

// messages, err := b.loadMessages(queueName)
// if err != nil {
// 	t.Fatalf("Failed to load messages: %v", err)
// }
// if len(messages) < 1 {
// 	t.Fatalf("Expected 1 message, but got %d", len(messages))
// }
// if messages[0].Content != "Hello, world!" {
// 	t.Fatalf("Expected message content 'Hello, world!', but got '%s'", messages[0].Content)
// }
// }

// func TestSaveBrokerState(t *testing.T) {
// b := NewBroker(conf)
// b.createQueue("testQueue")
// err := b.saveBrokerState()
// if err != nil {
// 	t.Fatalf("Failed to save broker state: %v", err)
// }

// // Verify the broker state file exists
// file := filepath.Join("data", "broker_state.json")
// if _, err := os.Stat(file); os.IsNotExist(err) {
// 	t.Fatalf("Expected broker state file %s to exist", file)
// }
// }

// func TestLoadBrokerState(t *testing.T) {
// b := NewBroker(conf)
// b.createQueue("testQueue")
// err := b.saveBrokerState()
// if err != nil {
// 	t.Fatalf("Failed to save broker state: %v", err)
// }

// // Create a new broker instance and load the state
// b2 := NewBroker(conf)
// err = b2.loadBrokerState()
// if err != nil {
// 	t.Fatalf("Failed to load broker state: %v", err)
// }

// // Verify the state was loaded correctly
// if _, ok := b2.Queues["testQueue"]; !ok {
// 	t.Fatalf("Expected queue 'testQueue' to exist after loading state")
// }
// }
