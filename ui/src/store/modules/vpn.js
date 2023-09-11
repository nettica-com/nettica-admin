import ApiService from "../../services/api.service";

const state = {
  error: null,
  vpns: [],
  vpnQrcodes: [],
  vpnConfigs: []
}

const getters = {
  error(state) {
    return state.error;
  },
  vpns(state) {
    return state.vpns;
  },
  getvpnQrcode: (state) => (id) => {
    let item = state.vpnQrcodes.find(item => item.id === id)
    // initial load fails, must wait promise and stuff...
    return item ? item.qrcode : "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg=="
    //    return "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
  },
  getVPNConfig: (state) => (id) => {
    let item = state.vpnConfigs.find(item => item.id === id)
    console.log("getVPNConfig: " + id + " item: " + item)
    return item ? item.config : null
  }
}

const actions = {
  error({ commit }, error) {
    commit('error', error)
  },

  readAll({ commit, dispatch }) {
    ApiService.get("/vpn")
      .then(resp => {
        commit('vpns', resp)

        //        dispatch('readQrcodes')
        //        dispatch('readConfigs')
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  create({ commit, dispatch }, vpn) {
    ApiService.post("/vpn", vpn)
      .then(resp => {
        commit('create', resp)
        commit('error', "VPN created")
      })
      .catch((error) => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  update({ commit, dispatch }, vpn) {
    ApiService.patch(`/vpn/${vpn.id}`, vpn)
      .then(resp => {
        commit('update', resp)
        commit('error', "VPN updated")
      })
      .catch((error) => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
        console.log(error)
      })
  },

  delete({ commit }, vpn) {
    ApiService.delete(`/vpn/${vpn.id}`)
      .then(() => {
        commit('delete', vpn)
        commit('error', "VPN deleted")
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  email({ commit }, vpn) {
    ApiService.get(`/vpn/${vpn.id}/email`)
      .then(() => {
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcode({ state, commit }, vpn) {
    ApiService.getWithConfig(`/vpn/${vpn.id}/config?qrcode=true`, { responseType: 'arraybuffer' })
      .then(resp => {
        let image = Buffer.from(resp, 'binary').toString('base64')
        commit('vpnQrcodes', { vpn, image })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readConfig({ state, commit }, vpn) {
    ApiService.getWithConfig(`/vpn/${vpn.id}/config?qrcode=false`, { responseType: 'arraybuffer' })
      .then(resp => {
        commit('vpnConfigs', { vpn: vpn, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcodes({ state, dispatch }) {
    state.vpns.forEach(vpn => {
      dispatch('readQrcode', vpn)
    })
  },

  readConfigs({ state, dispatch }) {
    state.vpns.forEach(vpn => {
      dispatch('readConfig', vpn)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  vpns(state, vpns) {
    state.vpns = vpns
  },
  create(state, vpn) {
    state.vpns.push(vpn)
  },
  update(state, vpn) {
    let index = state.vpns.findIndex(x => x.id === vpn.id);
    if (index !== -1) {
      state.vpns.splice(index, 1);
      state.vpns.push(vpn);
    } else {
      state.error = "update vpn failed, not in list"
    }
  },
  delete(state, vpn) {
    let index = state.vpns.findIndex(x => x.id === vpn.id);
    if (index !== -1) {
      state.vpns.splice(index, 1);
    } else {
      state.error = "delete vpn failed, not in list"
    }
  },
  vpnQrcodes(state, { vpn, image }) {
    let index = state.vpnQrcodes.findIndex(x => x.id === vpn.id);
    if (index !== -1) {
      state.vpnQrcodes.splice(index, 1);
    }
    state.vpnQrcodes.push({
      id: vpn.id,
      qrcode: image
    })
  },
  vpnConfigs(state, { vpn, config }) {
    let index = state.vpnConfigs.findIndex(x => x.id === vpn.id);
    if (index !== -1) {
      state.vpnConfigs.splice(index, 1);
    }
    state.vpnConfigs.push({
      id: vpn.id,
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
