import ApiService from "../../services/api.service";

const state = {
  error: null,
  hosts: [],
  hostQrcodes: [],
  hostConfigs: []
}

const getters = {
  error(state) {
    return state.error;
  },
  hosts(state) {
    return state.hosts;
  },
  gethostQrcode: (state) => (id) => {
    let item = state.hostQrcodes.find(item => item.id === id)
    // initial load fails, must wait promise and stuff...
    return item ? item.qrcode : "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=="
    //    return "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
  },
  gethostConfig: (state) => (id) => {
    let item = state.hostConfigs.find(item => item.id === id)
    return item ? item.config : null
  }
}

const actions = {
  error({ commit }, error) {
    commit('error', error)
  },

  readAll({ commit, dispatch }) {
    ApiService.get("/host")
      .then(resp => {
        for (var i = 0; i < resp.length; i++) {
          var host = resp[i]
          var last = new Date(host.lastSeen)
          var diff = Math.abs(Date.now() - last)
          console.log("Host: " + host.name + " lastSeen: " + host.lastSeen + " ms: " + diff)
          if (diff > 30000) {
            host.status = "Offline"
            if (host.platform == "Native" || host.platform == "iOS" || host.platform == "Android" || host.platform == "MacOS") {
              host.status = "Native"
            }
          } else {
            host.status = "Online"
          }
          commit('hosts', resp)
        }

        //        dispatch('readQrcodes')
        //        dispatch('readConfigs')
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, host) {
    ApiService.post("/host", host)
      .then(resp => {
        //        dispatch('readQrcode', resp)
        dispatch('readConfig', resp)
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, host) {
    ApiService.patch(`/host/${host.id}`, host)
      .then(resp => {
        //        dispatch('readQrcode', resp)
        //        dispatch('readConfig', host.id)
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, host) {
    ApiService.delete(`/host/${host.id}`)
      .then(() => {
        commit('delete', host)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  email({ commit }, host) {
    ApiService.get(`/host/${host.id}/email`)
      .then(() => {
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcode({ state, commit }, host) {
    ApiService.getWithConfig(`/host/${host.id}/config?qrcode=true`, { responseType: 'arraybuffer' })
      .then(resp => {
        let image = Buffer.from(resp, 'binary').toString('base64')
        commit('hostQrcodes', { host, image })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readConfig({ state, commit }, host) {
    ApiService.getWithConfig(`/host/${host.id}/config?qrcode=false`, { responseType: 'arraybuffer' })
      .then(resp => {
        commit('hostConfigs', { host: host, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcodes({ state, dispatch }) {
    state.hosts.forEach(host => {
      dispatch('readQrcode', host)
    })
  },

  readConfigs({ state, dispatch }) {
    state.hosts.forEach(host => {
      dispatch('readConfig', host)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  hosts(state, hosts) {
    state.hosts = hosts
  },
  create(state, host) {
    state.hosts.push(host)
  },
  update(state, host) {
    let index = state.hosts.findIndex(x => x.id === host.id);
    if (index !== -1) {
      state.hosts.splice(index, 1);
      state.hosts.push(host);
    } else {
      state.error = "update host failed, not in list"
    }
  },
  delete(state, host) {
    let index = state.hosts.findIndex(x => x.id === host.id);
    if (index !== -1) {
      state.hosts.splice(index, 1);
    } else {
      state.error = "delete host failed, not in list"
    }
  },
  hostQrcodes(state, { host, image }) {
    let index = state.hostQrcodes.findIndex(x => x.id === host.id);
    if (index !== -1) {
      state.hostQrcodes.splice(index, 1);
    }
    state.hostQrcodes.push({
      id: host.id,
      qrcode: image
    })
  },
  hostConfigs(state, { host, config }) {
    let index = state.hostConfigs.findIndex(x => x.id === host.id);
    if (index !== -1) {
      state.hostConfigs.splice(index, 1);
    }
    state.hostConfigs.push({
      id: host.id,
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
