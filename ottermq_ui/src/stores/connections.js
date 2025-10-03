import { defineStore } from "pinia"
import api from "src/services/api"

export const useConnectionsStore = defineStore("connections", {
  state: () => ({
    items: [],
    loading: false,
    error: null,
  }),
  actions: {
    async fetch() {
      this.loading = true;
      this.error = null
      try {
        const {data} = await api.get("/connections")
        this.items = Array.isArray(data?.connections) ? data.connections : []
        // Sort by vhost -> state (running first) -> user_name -> name (groups by vhost, prioritizes active connections)
        this.items.sort((a, b) => {
          // First: vhost (case-insensitive asc)
          const vhostCompare = (a.vhost || '').toLowerCase().localeCompare((b.vhost || '').toLowerCase())
          if (vhostCompare !== 0) return vhostCompare
          
          // Second: state (running before disconnected/other)
          const statePriority = (state) => state === 'running' ? 0 : 1  // 0 for running, 1 for others
          const stateCompare = statePriority(a.state) - statePriority(b.state)
          if (stateCompare !== 0) return stateCompare
          
          // Third: user_name (case-insensitive asc)
          const userNameCompare = (a.user_name || '').toLowerCase().localeCompare((b.user_name || '').toLowerCase())
          if (userNameCompare !== 0) return userNameCompare
          
          // Fourth: name (IP, case-insensitive asc)
          return (a.name || '').toLowerCase().localeCompare((b.name || '').toLowerCase())
        })
      } catch (e) {
        this.error = e?.response?.data?.error || e.message
        this.items = []
      } finally { this.loading = false }
    }
  }
})
