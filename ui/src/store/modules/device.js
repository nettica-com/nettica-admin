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
//    return "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
  },
  getdeviceConfig: (state) => (id) => {
    let item = state.deviceConfigs.find(item => item.id === id)
    return item ? item.config : null
  }
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  readAll({ commit, dispatch }){
    ApiService.get("/device")
      .then(resp => {
        for (var i = 0; i < resp.length; i++) {
          var device = resp[i]
          var last = new Date(device.lastSeen)
          var diff = Math.abs(Date.now() - last)
          console.log( "Host: " + device.name + " lastSeen: " + device.lastSeen + " ms: "  + diff)
          if (diff > 30000) {
              device.status = "Offline"
              if (device.platform == "Native" || device.platform == "iOS" || device.platform == "Android" || device.platform == "MacOS") {
                device.status = "Native"
              }
          } else {
              device.status = "Online"
          }
          commit('devices', resp)
        }
  
//        dispatch('readQrcodes')
//        dispatch('readConfigs')
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, device){
    ApiService.post("/device", device)
      .then(resp => {
//        dispatch('readQrcode', resp)
        dispatch('readConfig', resp)
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, device){
    ApiService.patch(`/device/${device.id}`, device)
      .then(resp => {
//        dispatch('readQrcode', resp)
//        dispatch('readConfig', device.id)
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, device){
    ApiService.delete(`/device/${device.id}`)
      .then(() => {
        commit('delete', device)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  email({ commit }, device){
    ApiService.get(`/device/${device.id}/email`)
      .then(() => {
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcode({ state, commit }, device){
    ApiService.getWithConfig(`/device/${device.id}/config?qrcode=true`, {responseType: 'arraybuffer'})
      .then(resp => {
        let image = Buffer.from(resp, 'binary').toString('base64')
        commit('deviceQrcodes', { device, image })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readConfig({ state, commit }, device){
    ApiService.getWithConfig(`/device/${device.id}/config?qrcode=false`, {responseType: 'arraybuffer'})
      .then(resp => {
        commit('deviceConfigs', { device: device, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcodes({ state, dispatch }){
    state.devices.forEach(device => {
      dispatch('readQrcode', device)
    })
  },

  readConfigs({ state, dispatch }){
    state.devices.forEach(device => {
      dispatch('readConfig', device)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  devices(state, devices){
    state.devices = devices
  },
  create(state, device){
    state.devices.push(device)
  },
  update(state, device){
    let index = state.devices.findIndex(x => x.id === device.id);
    if (index !== -1) {
      state.devices.splice(index, 1);
      state.devices.push(device);
    } else {
      state.error = "update device failed, not in list"
    }
  },
  delete(state, device){
    let index = state.devices.findIndex(x => x.id === device.id);
    if (index !== -1) {
      state.devices.splice(index, 1);
    } else {
      state.error = "delete device failed, not in list"
    }
  },
  deviceQrcodes(state, { device, image }){
    let index = state.deviceQrcodes.findIndex(x => x.id === device.id);
    if (index !== -1) {
      state.deviceQrcodes.splice(index, 1);
    }
    state.deviceQrcodes.push({
      id: device.id,
      qrcode: image
    })
  },
  deviceConfigs(state, { device, config }){
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
