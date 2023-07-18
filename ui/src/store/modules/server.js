import ApiService from "../../services/api.service";

const state = {
  error: null,
  servers: [],
  config: '',
  version: '1.0',
}

const getters = {
  error(state) {
    return state.error;
  },

  servers(state) {
    return state.servers;
  },

  version(state) {
    return state.version;
  },

  config(state) {
    return state.config;
  },
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  read({ commit, dispatch }){
    ApiService.get("/server")
      .then(resp => {
        commit('servers', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit }, server){
    ApiService.patch(`/server/${server.id}`, server)
      .then(resp => {
        // commit('servers', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },


  version({ commit }){
    ApiService.get("/server/version")
      .then(resp => {
        commit('version', resp.version)
      })
      .catch(err => {
        commit('error', err)
      })
  },

}

const mutations = {
  error(state, error) {
    state.error = error;
  },

  servers(state, servers){
    state.servers = servers
  },

  config(state, config){
    state.config = config
  },

  version(state, version){
    state.version = version
  },
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
