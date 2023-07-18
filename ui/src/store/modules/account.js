import ApiService from "../../services/api.service";

const state = {
  error: null,
  accounts: [],
  users: [],
  members: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  accounts(state) {
    return state.accounts;
  },
  users(state) {
    return state.users;
  },
  members(state) {
    return state.members;
  },
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  readAll({ commit, dispatch }, id){
    ApiService.get(`/accounts/${id}`)
      .then(resp => {
        commit('accounts', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readUsers({ commit, dispatch }, id){
    ApiService.get(`/accounts/${id}`)
      .then(resp => {
        commit('users', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readMembers({ commit, dispatch }, id){
    ApiService.get(`/accounts/${id}`)
      .then(resp => {
        commit('members', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit }, account){
    ApiService.post(`/accounts`, account)
      .then(resp => {
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, account){
    ApiService.patch(`/accounts/${account.id}`,account)
      .then(resp => {
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, account){
    ApiService.delete(`/accounts/${account.id}`)
      .then(() => {
        commit('delete', account)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  email({ commit }, account){
    ApiService.get(`/accounts/${account.id}/invite`)
      .then(() => {
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
  accounts(state, accounts){
    state.accounts = accounts
  },
  users(state, users) {
    state.users = users
  },
  members(state, members) {
    state.members = members
  },
  create(state, account){
    state.accounts.push(account)
  },
  create(state, user) {
    state.users.push(user)
  },
  update(state, account){
    let index = state.accounts.findIndex(x => x.id === account.id);
    if (index !== -1) {
      state.accounts.splice(index, 1);
      state.accounts.push(account);
    } else {
      state.error = "update account failed, not in list"
    }
  },
  delete(state, account){
    let index = state.accounts.findIndex(x => x.id === account.id);
    if (index !== -1) {
      state.accounts.splice(index, 1);
    } else {
      state.error = "delete account failed, not in list"
    }
  },
  update(state, user){
    let index = state.users.findIndex(x => x.id === user.id);
    if (index !== -1) {
      state.users.splice(index, 1);
      state.users.push(user);
    } else {
      state.error = "update account (user) failed, not in list"
    }
  },
  delete(state, user){
    let index = state.users.findIndex(x => x.id === user.id);
    if (index !== -1) {
      state.users.splice(index, 1);
    } else {
      state.error = "delete user failed, not in list"
    }
  },
  update(state, member){
    let index = state.users.findIndex(x => x.id === member.id);
    if (index !== -1) {
      state.members.splice(index, 1);
      state.members.push(member);
    } else {
      state.error = "update account (member) failed, not in list"
    }
  },
  delete(state, member){
    let index = state.users.findIndex(x => x.id === member.id);
    if (index !== -1) {
      state.members.splice(index, 1);
    } else {
      state.error = "delete user failed, not in list"
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
