import { defineStore } from 'pinia'
import api from 'src/services/api'

export const useExchangesStore = defineStore('exchanges', {
  state: () => ({
    items: [],
    bindings: [], // [{ routingKey, queue }]
    loading: false,
    error: null,
    selected: null,
  }),
  actions: {
    async fetch() {
      this.loading = true; this.error = null
      try {
        const {data} = await api.get('/exchanges')
        this.items = data.exchanges || []
      } catch (err) { this.error = err } 
      finally { this.loading = false }
    },
    async addExchange(name, type = 'direct') {
      await api.post('/exchanges', {exchange_name: name, exchange_type: type})
      await this.fetch()
    },
    async deleteExchange(name) {
      await api.delete(`/exchanges/${encodeURIComponent(name)}`)
      await this.fetch()
    },
    async fetchBindings(exchange) {
        const {data} = await api.get(`/bindings/${encodeURIComponent(exchange)}`)
        const list = []
        Object.entries(data.bindings ||{}).forEach(([key, value]) => {
          value.forEach(q => list.push({routingKey: key, queue: q}))
        })
        this.bindings = list
    },
    async addBinding(exchange, routingKey, queue) {
      await api.post(`/bindings`, {exchange_name: exchange, routing_key: routingKey, queue_name: queue})
      await this.fetchBindings(exchange)
    },
    async deleteBinding(exchange, routingKey, queue) {
      await api.delete(`/bindings`, {exchange_name: exchange, routing_key: routingKey, queue_name: queue})
      await this.fetchBindings(exchange)
    },
    select(exchange) {
      this.selected = exchange
    }
  }
})
