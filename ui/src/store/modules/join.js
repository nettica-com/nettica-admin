import ApiService from "../../services/api.service";

const state = {
  error: null,
  result: "",
}

const getters = {
  error(state) {
    return state.error;
  },
  join(state) {
    return state.join;
  },
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  activate({ state, commit }, id){
    ApiService.post("/accounts/"+ id + "/activate")
      .then(resp => {
        commit('result', resp)
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
  join(state, join){
    state.join = join
  },
  create(state, join){
    state.join.push(join)
  },
  
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
