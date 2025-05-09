package broker

import (
	. "github.com/andrelcunha/ottermq/pkg/common"
)

func ListExchanges(b *Broker) []ExchangeDTO {
	exchanges := make([]ExchangeDTO, 0, b.GetTotalExchanges())
	b.mu.Lock()
	defer b.mu.Unlock()
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, exchange := range b.VHosts[vhost.Name].Exchanges {
			exchanges = append(exchanges, ExchangeDTO{
				VHostName: vhost.Name,
				VHostId:   vhost.Id,
				Name:      exchange.Name,
				Type:      string(exchange.Typ),
			})
		}
	}
	return exchanges
}

func (b *Broker) GetTotalExchanges() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := 0
	for vhostName := range b.VHosts {
		vhost := b.VHosts[vhostName]
		for _, exchange := range b.VHosts[vhost.Name].Exchanges {
			if exchange.Name != "" {
				total++
			}
		}
	}
	return total
}
