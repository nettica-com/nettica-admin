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
        this.services = await ApiService.get('/service')
      } catch (err) {
        this.error = err
      }
    },

    async create(service) {
      try {
        const resp = await ApiService.post('/service', service)
        this.services.push(resp)
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
      } catch (err) {
        this.error = err
      }
    },

    async delete(service) {
      try {
        await ApiService.delete(`/service/${service.id}`)
        const index = this.services.findIndex((x) => x.id === service.id)
        if (index !== -1) this.services.splice(index, 1)
      } catch (err) {
        this.error = err
      }
    },
  },
})
