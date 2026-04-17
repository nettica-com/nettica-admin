import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'
import { useDeviceStore } from './device'

export const useVpnStore = defineStore('vpn', {
  state: () => ({
    error: null,
    vpns: [],
    vpnQrcodes: [],
    vpnConfigs: [],
  }),

  getters: {
    getvpnQrcode: (state) => (id) => {
      const item = state.vpnQrcodes.find((v) => v.id === id)
      return item
        ? item.qrcode
        : 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=='
    },
    getVPNConfig: (state) => (id) => {
      const item = state.vpnConfigs.find((v) => v.id === id)
      // console.log('getVPNConfig: ' + id + ' item: ' + item)
      return item ? item.config : null
    },
  },

  actions: {
    async readAll() {
      try {
        this.vpns = await ApiService.get('/vpn')
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async create(vpn) {
      try {
        const resp = await ApiService.post('/vpn', vpn)
        this.vpns.push(resp)
        this.error = `${vpn.name} created`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async update(vpn) {
      try {
        const resp = await ApiService.patch(`/vpn/${vpn.id}`, vpn)
        const index = this.vpns.findIndex((x) => x.id === resp.id)
        if (index !== -1) {
          this.vpns.splice(index, 1)
          this.vpns.push(resp)
        } else {
          this.error = 'update vpn failed, not in list'
        }
        useDeviceStore().update_vpn(resp)
        this.error = `${vpn.name} updated`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
        // console.log(error)
      }
    },

    async delete(vpn) {
      try {
        await ApiService.delete(`/vpn/${vpn.id}`)
        const index = this.vpns.findIndex((x) => x.id === vpn.id)
        if (index !== -1) this.vpns.splice(index, 1)
        this.error = `${vpn.name} deleted`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async email(vpn) {
      try {
        await ApiService.get(`/vpn/${vpn.id}/email`)
      } catch (err) {
        this.error = err
      }
    },

    async readQrcode(vpn) {
      try {
        const resp = await ApiService.getWithConfig(
          `/vpn/${vpn.id}/config?qrcode=true`,
          { responseType: 'arraybuffer' },
        )
        const image = btoa(String.fromCharCode(...new Uint8Array(resp)))
        const index = this.vpnQrcodes.findIndex((x) => x.id === vpn.id)
        if (index !== -1) this.vpnQrcodes.splice(index, 1)
        this.vpnQrcodes.push({ id: vpn.id, qrcode: image })
      } catch (err) {
        this.error = err
      }
    },

    async readConfig(vpn) {
      try {
        const resp = await ApiService.getWithConfig(
          `/vpn/${vpn.id}/config?qrcode=false`,
          { responseType: 'arraybuffer' },
        )
        const index = this.vpnConfigs.findIndex((x) => x.id === vpn.id)
        if (index !== -1) this.vpnConfigs.splice(index, 1)
        this.vpnConfigs.push({ id: vpn.id, config: resp })
      } catch (err) {
        this.error = err
      }
    },

    readQrcodes() {
      this.vpns.forEach((vpn) => this.readQrcode(vpn))
    },

    readConfigs() {
      this.vpns.forEach((vpn) => this.readConfig(vpn))
    },
  },
})
