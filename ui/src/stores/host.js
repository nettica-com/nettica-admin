import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useHostStore = defineStore('host', {
  state: () => ({
    error: null,
    hosts: [],
    hostQrcodes: [],
    hostConfigs: [],
  }),

  getters: {
    gethostQrcode: (state) => (id) => {
      const item = state.hostQrcodes.find((h) => h.id === id)
      return item
        ? item.qrcode
        : 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=='
    },
    gethostConfig: (state) => (id) => {
      const item = state.hostConfigs.find((h) => h.id === id)
      return item ? item.config : null
    },
  },

  actions: {
    async readAll() {
      try {
        const resp = await ApiService.get('/host')
        for (const host of resp) {
          const diff = Math.abs(Date.now() - new Date(host.lastSeen))
          // console.log('Host: ' + host.name + ' lastSeen: ' + host.lastSeen + ' ms: ' + diff)
          if (diff > 30000) {
            host.status =
              host.platform === 'Native' ||
              host.platform === 'iOS' ||
              host.platform === 'Android' ||
              host.platform === 'MacOS'
                ? 'Native'
                : 'Offline'
          } else {
            host.status = 'Online'
          }
        }
        const incoming = new Map(resp.map(h => [h.id, h]))
        for (let i = this.hosts.length - 1; i >= 0; i--) {
          const id = this.hosts[i].id
          if (incoming.has(id)) {
            this.hosts[i] = incoming.get(id)
            incoming.delete(id)
          } else {
            this.hosts.splice(i, 1)
          }
        }
        for (const h of incoming.values()) this.hosts.push(h)
        this.hosts = [...this.hosts]
      } catch (err) {
        this.error = err
      }
    },

    async create(host) {
      try {
        const resp = await ApiService.post('/host', host)
        await this.readConfig(resp)
        this.hosts.push(resp)
        this.hosts = [...this.hosts]
      } catch (err) {
        this.error = err
      }
    },

    async update(host) {
      try {
        const resp = await ApiService.patch(`/host/${host.id}`, host)
        const index = this.hosts.findIndex((x) => x.id === resp.id)
        if (index !== -1) {
          this.hosts.splice(index, 1)
          this.hosts.push(resp)
        } else {
          this.error = 'update host failed, not in list'
        }
        this.hosts = [...this.hosts]
      } catch (err) {
        this.error = err
      }
    },

    async delete(host) {
      try {
        await ApiService.delete(`/host/${host.id}`)
        const index = this.hosts.findIndex((x) => x.id === host.id)
        if (index !== -1) this.hosts.splice(index, 1)
        this.hosts = [...this.hosts]
      } catch (err) {
        this.error = err
      }
    },

    async email(host) {
      try {
        await ApiService.get(`/host/${host.id}/email`)
      } catch (err) {
        this.error = err
      }
    },

    async readQrcode(host) {
      try {
        const resp = await ApiService.getWithConfig(
          `/host/${host.id}/config?qrcode=true`,
          { responseType: 'arraybuffer' },
        )
        const image = btoa(String.fromCharCode(...new Uint8Array(resp)))
        const index = this.hostQrcodes.findIndex((x) => x.id === host.id)
        if (index !== -1) this.hostQrcodes.splice(index, 1)
        this.hostQrcodes.push({ id: host.id, qrcode: image })
      } catch (err) {
        this.error = err
      }
    },

    async readConfig(host) {
      try {
        const resp = await ApiService.getWithConfig(
          `/host/${host.id}/config?qrcode=false`,
          { responseType: 'arraybuffer' },
        )
        const index = this.hostConfigs.findIndex((x) => x.id === host.id)
        if (index !== -1) this.hostConfigs.splice(index, 1)
        this.hostConfigs.push({ id: host.id, config: resp })
      } catch (err) {
        this.error = err
      }
    },

    readQrcodes() {
      this.hosts.forEach((host) => this.readQrcode(host))
    },

    readConfigs() {
      this.hosts.forEach((host) => this.readConfig(host))
    },
  },
})
