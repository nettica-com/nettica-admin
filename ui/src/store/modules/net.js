import ApiService from "../../services/api.service";

const state = {
  error: null,
  nets: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  nets(state) {
    return state.nets;
  },
  getNetConfig: (state) => (id) => {
    let item = state.nets.find(item => item.id === id)
    return item ? item.config : null
  }
}

const actions = {
  error({ commit }, error) {
    commit('error', error)
  },

  readAll({ commit, dispatch }) {
    ApiService.get("/net")
      .then(resp => {
        commit('nets', resp)
        //        dispatch('readNetConfigs')
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  create({ commit, dispatch }, net) {
    ApiService.post("/net", net)
      .then(resp => {
        //        dispatch('readNetConfig', resp)
        commit('create', resp)
        commit('error', `Network ${net.netName} created`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  update({ commit, dispatch }, net) {
    ApiService.patch(`/net/${net.id}`, net)
      .then(resp => {
        //        dispatch('readNetConfig', resp)
        commit('update', resp)
        commit('error', `Network ${net.netName} updated`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  delete({ commit }, net) {
    ApiService.delete(`/net/${net.id}`)
      .then(() => {
        commit('delete', net)
        commit('error', `Network ${net.netName} deleted`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  readNetConfig({ state, commit }, net) {
    ApiService.getWithConfig(`/net/${net.id}`, { responseType: 'arraybuffer' })
      .then(resp => {
        //        commit('nets', { net: net, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readNetConfigs({ state, dispatch }) {
    state.nets.forEach(net => {
      dispatch('readNetConfig', net)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  nets(state, nets) {
    state.nets = nets
  },
  create(state, net) {
    state.nets.push(net)
  },
  update(state, net) {
    let index = state.nets.findIndex(x => x.id === net.id);
    if (index !== -1) {
      state.nets.splice(index, 1);
      state.nets.push(net);
    } else {
      state.error = "update net failed, not in list"
    }
  },
  delete(state, net) {
    let index = state.nets.findIndex(x => x.id === net.id);
    if (index !== -1) {
      state.nets.splice(index, 1);
    } else {
      state.error = "delete net failed, not in list"
    }
  },
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
