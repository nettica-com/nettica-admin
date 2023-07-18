import ApiService from "../../services/api.service";

const state = {
  error: null,
  services: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  services(state) {
    return state.services;
  },
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  read({ commit, dispatch }){
    ApiService.get("/service")
      .then(resp => {
        commit('services', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, service){
    ApiService.post("/service", service)
      .then(resp => {
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, service){
    ApiService.patch(`/service/${service.id}`, service)
      .then(resp => {
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, service){
    ApiService.delete(`/service/${service.id}`)
      .then(() => {
        commit('delete', service)
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
  services(state, services){
    state.services = services
  },
  create(state, service){
    state.services.push(service)
  },
  update(state, service){
    let index = state.services.findIndex(x => x.id === service.id);
    if (index !== -1) {
      state.services.splice(index, 1);
      state.services.push(service);
    } else {
      state.error = "update service failed, not in list"
    }
  },
  delete(state, service){
    let index = state.services.findIndex(x => x.id === service.id);
    if (index !== -1) {
      state.services.splice(index, 1);
    } else {
      state.error = "delete service failed, not in list"
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
