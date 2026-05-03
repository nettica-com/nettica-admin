import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useDeviceStore = defineStore('device', {
  state: () => ({
    error: null,
    devices: [],
    deviceQrcodes: [],
    deviceConfigs: [],
  }),

  getters: {
    getdeviceQrcode: (state) => (id) => {
      const item = state.deviceQrcodes.find((d) => d.id === id)
      return item
        ? item.qrcode
        : 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=='
    },
    getdeviceConfig: (state) => (id) => {
      const item = state.deviceConfigs.find((d) => d.id === id)
      return item ? item.config : null
    },
  },

  actions: {
    async readAll() {
      try {
        const resp = await ApiService.get('/device')
        for (const device of resp) {
          if (device.lastSeen == null) continue
          const diff = Math.abs(Date.now() - new Date(device.lastSeen))
          // console.log('Host: ' + device.name + ' lastSeen: ' + device.lastSeen + ' ms: ' + diff)
          if (diff > 60000) {
            device.status =
              device.platform === 'Windows' ||
              device.platform === 'Native' ||
              device.os === 'ios' ||
              device.os === 'android' ||
              device.os === 'macos'
                ? device.platform
                : 'Offline'
          } else {
            device.status = 'Online'
          }
        }
        const incoming = new Map(resp.map(d => [d.id, d]))
        for (let i = this.devices.length - 1; i >= 0; i--) {
          const id = this.devices[i].id
          if (incoming.has(id)) {
            this.devices[i] = incoming.get(id)
            incoming.delete(id)
          } else {
            this.devices.splice(i, 1)
          }
        }
        for (const d of incoming.values()) this.devices.push(d)
        this.devices = [...this.devices]
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async create(device) {
      try {
        const resp = await ApiService.post('/device', device)
        this.devices.push(resp)
        this.error = `Device ${device.name} created`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async update(device) {
      try {
        const resp = await ApiService.patch(`/device/${device.id}`, device)
        const index = this.devices.findIndex((x) => x.id === resp.id)
        if (index !== -1) {
          resp.vpns = this.devices[index].vpns
          this.devices.splice(index, 1)
          this.devices.push(resp)
        } else {
          this.error = 'update device failed, not in list'
        }
        this.error = `Device ${device.name} updated`
        this.devices = [...this.devices]
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    update_vpn(vpn) {
      const device = this.devices.find((x) => x.id === vpn.deviceid)
      if (device) {
        if (!device.vpns) {
          device.vpns = [vpn]
        } else {
          const vpnIndex = device.vpns.findIndex((x) => x.id === vpn.id)
          if (vpnIndex !== -1) {
            device.vpns.splice(vpnIndex, 1)
            device.vpns.push(vpn)
          }
        }
        const index = this.devices.findIndex((x) => x.id === device.id)
        if (index !== -1) {
          this.devices.splice(index, 1)
          this.devices.push(device)
        }
        this.devices = [...this.devices]
      } else {
        this.error = 'update vpn failed, not in list'
      }
    },

    async delete(device) {
      try {
        await ApiService.delete(`/device/${device.id}`)
        const index = this.devices.findIndex((x) => x.id === device.id)
        if (index !== -1) this.devices.splice(index, 1)
        this.error = `Device ${device.name} deleted`
        this.devices = [...this.devices]
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async email(device) {
      try {
        await ApiService.get(`/device/${device.id}/email`)
      } catch (err) {
        this.error = err
      }
    },

    async readQrcode(device) {
      try {
        const resp = await ApiService.getWithConfig(
          `/device/${device.id}/config?qrcode=true`,
          { responseType: 'arraybuffer' },
        )
        const image = btoa(String.fromCharCode(...new Uint8Array(resp)))
        const index = this.deviceQrcodes.findIndex((x) => x.id === device.id)
        if (index !== -1) this.deviceQrcodes.splice(index, 1)
        this.deviceQrcodes.push({ id: device.id, qrcode: image })
      } catch (err) {
        this.error = err
      }
    },

    async readConfig(device) {
      try {
        const resp = await ApiService.getWithConfig(
          `/device/${device.id}/config?qrcode=false`,
          { responseType: 'arraybuffer' },
        )
        const index = this.deviceConfigs.findIndex((x) => x.id === device.id)
        if (index !== -1) this.deviceConfigs.splice(index, 1)
        this.deviceConfigs.push({ id: device.id, config: resp })
      } catch (err) {
        this.error = err
      }
    },

    readQrcodes() {
      this.devices.forEach((device) => this.readQrcode(device))
    },

    readConfigs() {
      this.devices.forEach((device) => this.readConfig(device))
    },
  },
})
