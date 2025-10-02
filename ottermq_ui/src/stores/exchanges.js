import { defineStore } from 'pinia'
import api from 'src/services/api'

export const useExchangesStore = defineStore('exchanges', {
  state: () => ({
    items: [],
    bindings: [],
    loading: false,
    error: null,
    selected: null,
  }),
  actions: {
    async fetch() {
      this.loading = true; 
      this.error = null
      try {
        const {data} = await api.get('/exchanges')
        this.items = Array.isArray(data?.exchanges) ? data.exchanges : []
      } catch (err) {
        this.error = err?.response?.data?.error || err.message
        this.items = []
      } finally { 
        this.loading = false 
      }
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
        const map = data?.bindings ?? {}
        const list = []
        Object.entries(map).forEach(([routing_key, queues]) => {
        (queues || []).forEach(q => list.push({routingKey: routing_key, queue: q}))
        })
        this.bindings = list
    },
    async addBinding(exchange, routingKey, queue) {
      await api.post(`/bindings`, {
        exchange_name: exchange, routing_key: routingKey, queue_name: queue
      })
      await this.fetchBindings(exchange)
    },
    async deleteBinding(exchange, routingKey, queue) {
      await api.delete(`/bindings`, { data: { exchange_name: exchange, routing_key: routingKey, queue_name: queue } })
      await this.fetchBindings(exchange)
    },
    async publish(exchange, routingKey, message) {
      await api.post(`/messages`, {
        exchange_name: exchange, routing_key: routingKey, message
      })
    },
    select(exchange) { this.selected = exchange },
  }
})
