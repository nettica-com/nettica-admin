import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'
import TokenService from '@/services/token.service'

export const useWildnetStore = defineStore('wildnet', {
  state: () => ({
    error: null,
    nets: [],
    server: '',
    serverError: null,
  }),

  getters: {
    getNetConfig: (state) => (id) => {
      const item = state.nets.find((n) => n.id === id)
      return item ? item.config : null
    },
  },

  actions: {
    wildServer(server) {
      this.server = server
      TokenService.saveWildServer(server)
      ApiService.setWildServer(server)
    },

    wildServerError(error) {
      this.serverError = error
    },

    async readAll() {
      ApiService.setWildServer()
      ApiService.setWildHeader()
      try {
        this.nets = await ApiService.get('/net')
      } catch (error) {
        TokenService.destroyWildToken()
        TokenService.destroyWildServer()
        if (error.response) this.error = error.response.data.error
      }
      ApiService.setServer()
      ApiService.setHeader()
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
