package broker

import (
	. "github.com/andrelcunha/ottermq/pkg/common"
)

func ListQueues(b *Broker) []QueueDTO {
	queues := make([]QueueDTO, 0, b.GetTotalQueues())
	b.mu.Lock()
	defer b.mu.Unlock()
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, queue := range b.VHosts[vhost.Name].Queues {
			queues = append(queues, QueueDTO{
				VHostName: vhost.Name,
				VHostId:   vhost.Id,
				Name:      queue.Name,
				Messages:  queue.Len(),
			})
		}
	}
	return queues
}

func (b *Broker) GetTotalQueues() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := 0
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, queue := range b.VHosts[vhost.Name].Queues {
			if queue.Name != "" {
				total++
			}
		}
	}
	return total
}
