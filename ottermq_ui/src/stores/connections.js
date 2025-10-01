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
      } catch (e) {
        this.error = e?.response?.data?.error || e.message
        this.items = []
      } finally { this.loading = false }
    }
  }
})
