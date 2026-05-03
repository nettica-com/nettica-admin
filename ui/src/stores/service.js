import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useServiceStore = defineStore('service', {
  state: () => ({
    error: null,
    services: [],
  }),

  actions: {
    async read() {
      try {
        const resp = await ApiService.get('/service')
        const incoming = new Map(resp.map(s => [s.id, s]))
        for (let i = this.services.length - 1; i >= 0; i--) {
          const id = this.services[i].id
          if (incoming.has(id)) {
            this.services[i] = incoming.get(id)
            incoming.delete(id)
          } else {
            this.services.splice(i, 1)
          }
        }
        for (const s of incoming.values()) this.services.push(s)
        this.services = [...this.services]
      } catch (err) {
        this.error = err
      }
    },

    async create(service) {
      try {
        const resp = await ApiService.post('/service', service)
        this.services.push(resp)
        this.services = [...this.services]
      } catch (err) {
        this.error = err
      }
    },

    async update(service) {
      try {
        const resp = await ApiService.patch(`/service/${service.id}`, service)
        const index = this.services.findIndex((x) => x.id === resp.id)
        if (index !== -1) {
          this.services.splice(index, 1)
          this.services.push(resp)
        } else {
          this.error = 'update service failed, not in list'
        }
        this.services = [...this.services]
      } catch (err) {
        this.error = err
      }
    },

    async delete(service) {
      try {
        await ApiService.delete(`/service/${service.id}`)
        const index = this.services.findIndex((x) => x.id === service.id)
        if (index !== -1) this.services.splice(index, 1)
        this.services = [...this.services]
      } catch (err) {
        this.error = err
      }
    },
  },
})
