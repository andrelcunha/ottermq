package broker

import (
	"os"
	"path/filepath"

	"github.com/andrelcunha/ottermq/internal/core/broker/vhost"
)

// RecoverBrokerState loads vhosts, exchanges, queues, bindings, and messages from disk
func RecoverBrokerState(b *Broker) error {
	vhostsDir := "data/vhosts"
	entries, err := os.ReadDir(vhostsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		vhostName := entry.Name()
		v := vhost.NewVhost(vhostName, b.config.QueueBufferSize, b.persist)
		// Recover exchanges
		exchangesDir := filepath.Join(vhostsDir, vhostName, "exchanges")
		exFiles, _ := os.ReadDir(exchangesDir)
		for _, exFile := range exFiles {
			exName := exFile.Name()
			exName = exName[:len(exName)-len(".json")]
			persistedEx, err := b.persist.LoadExchange(vhostName, exName)
			if err == nil {
				// Create exchange in vhost using persistedEx
				v.RecoverExchange(persistedEx)
			}
		}
		// Recover queues
		queuesDir := filepath.Join(vhostsDir, vhostName, "queues")
		qFiles, _ := os.ReadDir(queuesDir)
		for _, qFile := range qFiles {
			qName := qFile.Name()
			qName = qName[:len(qName)-len(".json")]
			persistedQ, err := b.persist.LoadQueue(vhostName, qName)
			if err == nil {
				// Create queue in vhost using persistedQ
				v.RecoverQueue(persistedQ)
			}
		}
		b.VHosts[vhostName] = v
	}
	return nil
}
