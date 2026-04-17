import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useServerStore = defineStore('server', {
  state: () => ({
    error: null,
    servers: [],
    config: '',
    version: '2.0',
  }),

  actions: {
    async read() {
      try {
        this.servers = await ApiService.get('/server')
      } catch (err) {
        this.error = err
      }
    },

    async update(server) {
      try {
        await ApiService.patch(`/server/${server.id}`, server)
      } catch (err) {
        this.error = err
      }
    },

    async version() {
      try {
        const resp = await ApiService.get('/server/version')
        this.version = resp.version
      } catch (err) {
        this.error = err
      }
    },
  },
})
