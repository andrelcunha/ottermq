package broker

import (
	"os"
	"path/filepath"
	"strings"

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
			exName = strings.TrimSuffix(exName, ".json")
			// persistedEx, err := b.persist.LoadExchange(vhostName, exName)
			exchangeType, props, err := b.persist.LoadExchangeMetadata(vhostName, exName)

			if err == nil {
				// Create exchange in vhost using persistedEx
				v.RecoverExchange(exName, exchangeType, props)
			}
		}
		queuesDir := filepath.Join(vhostsDir, vhostName, "queues")
		qFiles, _ := os.ReadDir(queuesDir)
		for _, qFile := range qFiles {
			qName := qFile.Name()
			qName = strings.TrimSuffix(qName, ".json")
			props, err := b.persist.LoadQueueMetadata(vhostName, qName)
			if err == nil {
				// Create queue in vhost using persistedQ
				v.RecoverQueue(qName, &props)
			}
		}
		b.VHosts[vhostName] = v
	}
	return nil
}
