import { defineStore } from "pinia"
import api from "src/services/api"

export const useConnectionsStore = defineStore("connections", {
  state: () => ({
    items: [],
    loading: false,
    error: null,
  }),
  actions: {
    async fetchConnections() {
      this.loading = true ; this.error = null
      try {
        const {data} = await api.get("/connections")
        this.items = data.connections || []
      } catch (error) { this.error = error } 
      finally { this.loading = false }
    }
  }
})
