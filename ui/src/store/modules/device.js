import ApiService from "../../services/api.service";

const state = {
  error: null,
  devices: [],
  deviceQrcodes: [],
  deviceConfigs: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  devices(state) {
    return state.devices;
  },
  getdeviceQrcode: (state) => (id) => {
    let item = state.deviceQrcodes.find(item => item.id === id)
    // initial load fails, must wait promise and stuff...
    return item ? item.qrcode : "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=="
  },
  getdeviceConfig: (state) => (id) => {
    let item = state.deviceConfigs.find(item => item.id === id)
    return item ? item.config : null
  }
}

const actions = {
  error({ commit }, error) {
    commit('error', error)
  },

  readAll({ commit, dispatch }) {
    ApiService.get("/device")
      .then(resp => {
        for (var i = 0; i < resp.length; i++) {
          var device = resp[i]
          if (device.lastSeen == null) {
            continue
          }
          var last = new Date(device.lastSeen)
          var diff = Math.abs(Date.now() - last)
          console.log("Host: " + device.name + " lastSeen: " + device.lastSeen + " ms: " + diff)
          if (diff > 60000) {
            device.status = "Offline"
            if (device.platform == "Windows" || device.platform == "Native" || device.platform == "ios" || device.os == "android" || device.platform == "MacOS") {
              device.status = "Native"
            }
          } else {
            device.status = "Online"
          }
        }
        commit('devices', resp)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  create({ commit, dispatch }, device) {
    ApiService.post("/device", device)
      .then(resp => {
        commit('create', resp)
        commit('error', `Device ${device.name} created`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  update({ commit, dispatch }, device) {
    ApiService.patch(`/device/${device.id}`, device)
      .then(resp => {
        commit('update', resp)
        commit('error', `Device ${device.name} updated`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  update_vpn({ commit, dispatch }, vpn) {
    console.log("update_vpn", vpn)
    let device = state.devices.find(x => x.id === vpn.deviceid);
    if (device !== null) {
      if (device.vpns == null) {
        device.vpns = []
        device.vpns.push(vpn)
      } else {
        let vpnindex = device.vpns.findIndex(x => x.id === vpn.id);
        if (vpnindex !== -1) {
          device.vpns.splice(vpnindex, 1);
          device.vpns.push(vpn);
        }
      }
      commit('update', device)
    } else {
      commit('error', "update vpn failed, not in list")
    }

  },


  delete({ commit }, device) {
    ApiService.delete(`/device/${device.id}`)
      .then(() => {
        commit('delete', device)
        commit('error', `Device ${device.name} deleted`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  email({ commit }, device) {
    ApiService.get(`/device/${device.id}/email`)
      .then(() => {
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcode({ state, commit }, device) {
    ApiService.getWithConfig(`/device/${device.id}/config?qrcode=true`, { responseType: 'arraybuffer' })
      .then(resp => {
        let image = Buffer.from(resp, 'binary').toString('base64')
        commit('deviceQrcodes', { device, image })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readConfig({ state, commit }, device) {
    ApiService.getWithConfig(`/device/${device.id}/config?qrcode=false`, { responseType: 'arraybuffer' })
      .then(resp => {
        commit('deviceConfigs', { device: device, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcodes({ state, dispatch }) {
    state.devices.forEach(device => {
      dispatch('readQrcode', device)
    })
  },

  readConfigs({ state, dispatch }) {
    state.devices.forEach(device => {
      dispatch('readConfig', device)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  devices(state, devices) {
    state.devices = devices
  },
  create(state, device) {
    state.devices.push(device)
  },
  update(state, device) {
    let index = state.devices.findIndex(x => x.id === device.id);
    if (index !== -1) {
      device.vpns = state.devices[index].vpns
      state.devices.splice(index, 1);
      state.devices.push(device);
    } else {
      state.error = "update device failed, not in list"
    }
  },
  delete(state, device) {
    let index = state.devices.findIndex(x => x.id === device.id);
    if (index !== -1) {
      state.devices.splice(index, 1);
    } else {
      state.error = "delete device failed, not in list"
    }
  },
  deviceQrcodes(state, { device, image }) {
    let index = state.deviceQrcodes.findIndex(x => x.id === device.id);
    if (index !== -1) {
      state.deviceQrcodes.splice(index, 1);
    }
    state.deviceQrcodes.push({
      id: device.id,
      qrcode: image
    })
  },
  deviceConfigs(state, { device, config }) {
    let index = state.deviceConfigs.findIndex(x => x.id === device.id);
    if (index !== -1) {
      state.deviceConfigs.splice(index, 1);
    }
    state.deviceConfigs.push({
      id: device.id,
      config: config
    })
  },
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
