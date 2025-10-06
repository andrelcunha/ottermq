package persistence

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
)

// Create an interface for persistence operations
type Persistence interface {
	SaveExchange(vhostName string, exchange *PersistedExchange) error
	LoadExchange(vhostName, exchangeName string) (*PersistedExchange, error)
	SaveQueue(vhostName string, queue *PersistedQueue) error
	LoadQueue(vhostName, queueName string) (*PersistedQueue, error)
}

// create a default implementation for PersistDB
type DefaultPersistence struct {
}

func NewDefaultPersistence() *DefaultPersistence {
	return &DefaultPersistence{}
}

// safeVHostName encodes vhost names for safe filesystem usage
func safeVHostName(name string) string {
	return url.PathEscape(name)
}

// SaveExchange persists a single exchange to a JSON file
func (db *DefaultPersistence) SaveExchange(vhostName string, exchange *PersistedExchange) error {
	safeName := safeVHostName(vhostName)
	dir := filepath.Join("data", "vhosts", safeName, "exchanges")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file := filepath.Join(dir, exchange.Name+".json")
	data, err := json.MarshalIndent(exchange, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// LoadExchange loads a single exchange from a JSON file
func (db *DefaultPersistence) LoadExchange(vhostName, exchangeName string) (*PersistedExchange, error) {
	safeName := safeVHostName(vhostName)
	file := filepath.Join("data", "vhosts", safeName, "exchanges", exchangeName+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var exchange PersistedExchange
	if err := json.Unmarshal(data, &exchange); err != nil {
		return nil, err
	}
	return &exchange, nil
}

// SaveQueue persists a single queue to a JSON file
func (db *DefaultPersistence) SaveQueue(vhostName string, queue *PersistedQueue) error {
	safeName := safeVHostName(vhostName)
	dir := filepath.Join("data", "vhosts", safeName, "queues")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file := filepath.Join(dir, queue.Name+".json")
	data, err := json.MarshalIndent(queue, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// LoadQueue loads a single queue from a JSON file
func (db *DefaultPersistence) LoadQueue(vhostName, queueName string) (*PersistedQueue, error) {
	safeName := safeVHostName(vhostName)
	file := filepath.Join("data", "vhosts", safeName, "queues", queueName+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var queue PersistedQueue
	if err := json.Unmarshal(data, &queue); err != nil {
		return nil, err
	}
	return &queue, nil
}
