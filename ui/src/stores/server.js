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
        const resp = await ApiService.get('/server')
        const incoming = new Map(resp.map(s => [s.id, s]))
        for (let i = this.servers.length - 1; i >= 0; i--) {
          const id = this.servers[i].id
          if (incoming.has(id)) {
            this.servers[i] = incoming.get(id)
            incoming.delete(id)
          } else {
            this.servers.splice(i, 1)
          }
        }
        for (const s of incoming.values()) this.servers.push(s)
        this.servers = [...this.servers]
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
