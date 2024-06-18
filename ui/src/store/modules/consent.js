import ApiService from "../../services/api.service";

const state = {
  error: null,
  account: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  account(state) {
    return state.account;
  },
}

const actions = {
  error({ commit }, error) {
    commit('error', error)
  },

  activate({ state, commit }, id) {
    ApiService.post("/accounts/" + id + "/activate")
      .then(resp => {
        commit('account', resp)
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
  account(state, account) {
    state.account = account;
  },

}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
