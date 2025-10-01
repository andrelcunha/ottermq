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
        const list = data?.connections ?? data?.data?.connections ?? []
        this.items = Array.isArray(list) ? list : []
      } catch (error) {
        this.error = error?.response?.data?.error || error.message || String(error)
        this.items = []
      } finally {
        this.loading = false
      }
    }
  }
})
