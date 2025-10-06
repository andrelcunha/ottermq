package persistence

// DummyPersistence implements persistence.Persistence with no-ops for testing
type DummyPersistence struct{}

func (d *DummyPersistence) SaveExchange(vhost string, ex *PersistedExchange) error {
	return nil
}
func (d *DummyPersistence) LoadExchange(vhost, name string) (*PersistedExchange, error) {
	return nil, nil
}
func (d *DummyPersistence) DeleteExchange(vhost, name string) error         { return nil }
func (d *DummyPersistence) SaveQueue(vhost string, q *PersistedQueue) error { return nil }
func (d *DummyPersistence) LoadQueue(vhost, name string) (*PersistedQueue, error) {
	return nil, nil
}
func (d *DummyPersistence) DeleteQueue(vhost, name string) error { return nil }
