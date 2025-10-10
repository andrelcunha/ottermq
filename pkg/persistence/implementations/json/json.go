package json

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/andrelcunha/ottermq/pkg/persistence"
)

type JsonPersistence struct {
	dataDir string
}

func NewJsonPersistence(config *persistence.Config) (*JsonPersistence, error) {
	jp := &JsonPersistence{
		dataDir: config.DataDir,
	}
	return jp, jp.Initialize()
}

func (jp *JsonPersistence) Initialize() error {
	// Create the base data directory if it doesn't exist
	if err := os.MkdirAll(jp.dataDir, 0755); err != nil {
		return err
	}
	return nil
}

func (jp *JsonPersistence) Close() error {
	// JSON implementation doesn't need to clean up
	return nil
}

// safeVHostName encodes vhost names for safe filesystem usage
func safeVHostName(name string) string {
	return url.PathEscape(name)
}

// SaveExchangeMetadata persists exchange metadata to a JSON file
func (jp *JsonPersistence) SaveExchangeMetadata(vhost, name, exchangeType string, props persistence.ExchangeProperties) error {
	safeName := safeVHostName(vhost)
	dir := filepath.Join(jp.dataDir, "vhosts", safeName, "exchanges")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	exchangeData := JsonExchangeData{
		Name:       name,
		Type:       exchangeType,
		Properties: props,
	}
	file := filepath.Join(dir, name+".json")
	data, err := json.MarshalIndent(exchangeData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// LoadExchangeMetadata loads exchange metadata from a JSON file (type and properties)
func (jp *JsonPersistence) LoadExchangeMetadata(vhost, name string) (exchangeType string, props persistence.ExchangeProperties, err error) {
	safeName := safeVHostName(vhost)
	file := filepath.Join(jp.dataDir, "vhosts", safeName, "exchanges", name+".json")

	data, err := os.ReadFile(file)
	if err != nil {
		return "", persistence.ExchangeProperties{}, err
	}

	var exchangeData JsonExchangeData

	if err := json.Unmarshal(data, &exchangeData); err != nil {
		return "", persistence.ExchangeProperties{}, err
	}
	return exchangeData.Type, exchangeData.Properties, nil
}

// DeleteExchange removes a single exchange JSON file
func (jp *JsonPersistence) DeleteExchangeMetadata(vhost, name string) error {
	safeName := safeVHostName(vhost)
	file := filepath.Join(jp.dataDir, "vhosts", safeName, "exchanges", name+".json")
	if err := os.Remove(file); err != nil {
		return err
	}
	return nil
}

// SaveQueueMetadata persists a single queue's metadata to a JSON file
func (jp *JsonPersistence) SaveQueueMetadata(vhost, name string, props persistence.QueueProperties) error {
	safeName := safeVHostName(vhost)
	dir := filepath.Join(jp.dataDir, "vhosts", safeName, "queues")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	existingQueue := &JsonQueueData{}

	// Update metadata while preserving messages
	queueData := JsonQueueData{
		Name:       name,
		Properties: props,
		Messages:   []JsonMessageData{}, // Preserve existing messages
	}

	// For JSON implementation, we need to preserve existing messages if they exist
	existingQueue, err := jp.loadQueueFile(vhost, name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if existingQueue != nil && len(existingQueue.Messages) > 0 {
		queueData.Messages = existingQueue.Messages
	}

	file := filepath.Join(dir, name+".json")
	data, err := json.MarshalIndent(queueData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// LoadQueueMetadata loads a single queue's metadata from a JSON file (props)
func (jp *JsonPersistence) LoadQueueMetadata(vhost, name string) (props persistence.QueueProperties, err error) {
	queueData, err := jp.loadQueueFile(vhost, name)
	if err != nil {
		return persistence.QueueProperties{}, err
	}
	return queueData.Properties, nil
}

func (jp *JsonPersistence) SaveMessage(vhost, queue, msgId string, msgBody []byte, msgProps persistence.MessageProperties) error {
	qData, err := jp.loadQueueFile(vhost, queue)
	if err != nil {
		return fmt.Errorf("queue not found: %v", err)
	}

	newMessage := JsonMessageData{
		ID:         msgId,
		Body:       msgBody,
		Properties: msgProps,
	}

	qData.Messages = append(qData.Messages, newMessage)

	return jp.SaveQueueMetadata(vhost, queue, qData.Properties)

}

func (jp *JsonPersistence) LoadMessages(vhostName, queueName string) ([]struct {
	ID         string                        `json:"id"`
	Body       []byte                        `json:"body"`
	Properties persistence.MessageProperties `json:"properties"`
}, error) {
	qData, err := jp.loadQueueFile(vhostName, queueName)
	if err != nil {
		return nil, fmt.Errorf("queue not found: %v", err)
	}

	var messages []struct {
		ID         string                        `json:"id"`
		Body       []byte                        `json:"body"`
		Properties persistence.MessageProperties `json:"properties"`
	}
	for _, msg := range qData.Messages {
		messages = append(messages, struct {
			ID         string                        `json:"id"`
			Body       []byte                        `json:"body"`
			Properties persistence.MessageProperties `json:"properties"`
		}{
			ID:         msg.ID,
			Body:       msg.Body,
			Properties: msg.Properties,
		})
	}
	return messages, nil
}

// LoadQueue loads a single queue from a JSON file
func (jp *JsonPersistence) loadQueueFile(vhost, name string) (*JsonQueueData, error) {
	safeName := safeVHostName(vhost)
	file := filepath.Join(jp.dataDir, "vhosts", safeName, "queues", name+".json")

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var queueData JsonQueueData
	if err := json.Unmarshal(data, &queueData); err != nil {
		return nil, err
	}
	return &queueData, nil
}

/*
// Message operations (basic implementation for JSON)
func (jp *JsonPersistence) PublishMessage(vhost, queue, messageID string, body []byte, props persistence.MessageProperties) error {
    // Load existing queue data
    queueData, err := jp.loadQueueFile(vhost, queue)
    if err != nil {
        return fmt.Errorf("queue not found: %v", err)
    }

    // Add new message
    newMessage := JsonMessageData{
        ID:         messageID,
        Body:       body,
        Properties: props,
    }

    queueData.Messages = append(queueData.Messages, newMessage)

    // Save updated queue
    return jp.saveQueueFile(vhost, queue, queueData)
}

func (jp *JsonPersistence) ConsumeMessage(vhost, queue string) (messageID string, body []byte, props persistence.MessageProperties, error) {
    // Load existing queue data
    queueData, err := jp.loadQueueFile(vhost, queue)
    if err != nil {
        return "", nil, persistence.MessageProperties{}, fmt.Errorf("queue not found: %v", err)
    }

    // Check if queue has messages
    if len(queueData.Messages) == 0 {
        return "", nil, persistence.MessageProperties{}, fmt.Errorf("queue is empty")
    }

    // Get first message (FIFO)
    message := queueData.Messages[0]
    queueData.Messages = queueData.Messages[1:] // Remove from queue

    // Save updated queue
    if err := jp.saveQueueFile(vhost, queue, queueData); err != nil {
        return "", nil, persistence.MessageProperties{}, err
    }

    return message.ID, message.Body, message.Properties, nil
}

func (jp *JsonPersistence) AckMessage(vhost, queue, messageID string) error {
    // For JSON implementation, messages are already removed when consumed
    // This is a no-op for the current implementation
    // In a future version, we might track unacked messages separately
    return nil
}
*/

// DeleteQueue removes a single queue JSON file
func (jp *JsonPersistence) DeleteQueueMetadata(vhostName, queueName string) error {
	safeName := safeVHostName(vhostName)
	file := filepath.Join(jp.dataDir, "vhosts", safeName, "queues", queueName+".json")
	if err := os.Remove(file); err != nil {
		return err
	}
	return nil
}
