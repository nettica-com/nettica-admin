import ApiService from "../../services/api.service";

const state = {
  error: null,
  subscriptions: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  subscriptions(state) {
    return state.subscriptions;
  }
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  read({ commit, dispatch }){
    ApiService.get("/subscriptions")
      .then(resp => {
        commit('subscriptions', resp)
//        dispatch('readMeshConfigs')
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, mesh){
    ApiService.post("/subscriptions", mesh)
      .then(resp => {
//        dispatch('readMeshConfig', resp)
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, subscription){
    ApiService.patch(`/subscriptions/${subscription.id}`, subscription)
      .then(resp => {
//        dispatch('readMeshConfig', resp)
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, subscription){
    ApiService.delete(`/subscriptions/${subscription.id}`)
      .then(() => {
        commit('delete', subscription)
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
  subscriptions(state, subscriptions){
    state.subscriptions = subscriptions;
  },
  create(state, subscription){
    state.subscriptions.push(subscription)
  },
  update(state, subscription){
    let index = state.subscriptions.findIndex(x => x.id === subscription.id);
    if (index !== -1) {
      state.subscriptions.splice(index, 1);
      state.subscriptions.push(subscription);
    } else {
      state.error = "update subscription failed, not in list"
    }
  },
  delete(state, subscription){
    let index = state.subscriptions.findIndex(x => x.id === subscription.id);
    if (index !== -1) {
      state.subscriptions.splice(index, 1);
    } else {
      state.error = "delete subscription failed, not in list"
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
