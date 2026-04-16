import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useNetStore = defineStore('net', {
  state: () => ({
    error: null,
    nets: [],
  }),

  getters: {
    getNetConfig: (state) => (id) => {
      const item = state.nets.find((n) => n.id === id)
      return item ? item.config : null
    },
  },

  actions: {
    async readAll() {
      try {
        this.nets = await ApiService.get('/net')
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async create(net) {
      try {
        const resp = await ApiService.post('/net', net)
        this.nets.push(resp)
        this.error = `Network ${net.netName} created`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async update(net) {
      try {
        const resp = await ApiService.patch(`/net/${net.id}`, net)
        const index = this.nets.findIndex((x) => x.id === resp.id)
        if (index !== -1) {
          this.nets.splice(index, 1)
          this.nets.push(resp)
        } else {
          this.error = 'update net failed, not in list'
        }
        this.error = `Network ${net.netName} updated`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    update_net(net) {
      const index = this.nets.findIndex((x) => x.id === net.id)
      if (index !== -1) {
        this.nets.splice(index, 1)
        this.nets.push(net)
      }
    },

    async delete(net) {
      try {
        await ApiService.delete(`/net/${net.id}`)
        const index = this.nets.findIndex((x) => x.id === net.id)
        if (index !== -1) this.nets.splice(index, 1)
        this.error = `Network ${net.netName} deleted`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },
  },
})
