import ApiService from "../../services/api.service";

const state = {
  error: null,
  account: null,
  accounts: [],
  users: [],
  members: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  account(state) {
    return state.account;
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
  getMembers: (state) => (id) => {
    let item = state.members.find(item => item.id === id)
    return item ? item.members : null
  }

}

const actions = {
  error({ commit }, error) {
    commit('error', error)
  },

  readAll({ commit, dispatch }, id) {
    ApiService.get(`/accounts/${id}`)
      .then(resp => {
        commit('accounts', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readUsers({ commit, dispatch }, id) {
    ApiService.get(`/accounts/${id}/users`)
      .then(resp => {
        commit('users', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readMembers({ commit, dispatch }, id) {
    ApiService.get(`/accounts/${id}/users`, { responseType: 'arraybuffer' })
      .then(resp => {
        console.log( "readMembers: ", resp)
        commit('members', {  id: id, members: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, account) {
    commit('error', null)
    ApiService.post(`/accounts/`, account)
      .then(resp => {
        commit('account', resp)
      }) 
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  update({ commit, dispatch }, account) {
    commit('error', null)
    ApiService.patch(`/accounts/${account.id}`, account)
      .then(resp => {
        commit('update', resp)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  delete({ commit }, account) {
    ApiService.delete(`/accounts/${account.id}`)
      .then(() => {
        commit('delete', account)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  email({ commit }, account) {
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
  account(state, account) {
    state.account = account
  },
  accounts(state, accounts) {
    state.accounts = accounts
  },
  users(state, users) {
    state.users = users
  },
  members(state, { id, members }) {
    let index = state.members.findIndex(x => x.id === id);
    if (index !== -1) {
      state.members.splice(index, 1);
    }
    state.members.push({
      id: id,
      members: members
    })
  },
  create(state, account) {
    state.accounts.push(account)
  },
  create(state, user) {
    state.users.push(user)
  },
  update(state, account) {
    let index = state.accounts.findIndex(x => x.id === account.id);
    if (index !== -1) {
      state.accounts.splice(index, 1);
      state.accounts.push(account);
    } else {
      state.error = "update account failed, not in list"
    }
  },
  delete(state, account) {
    let index = state.accounts.findIndex(x => x.id === account.id);
    if (index !== -1) {
      state.accounts.splice(index, 1);
    } else {
      state.error = "delete account failed, not in list"
    }
  },
  update(state, user) {
    let index = state.users.findIndex(x => x.id === user.id);
    if (index !== -1) {
      state.users.splice(index, 1);
      state.users.push(user);
    } else {
      state.error = "update account (user) failed, not in list"
    }
  },
  delete(state, user) {
    let index = state.users.findIndex(x => x.id === user.id);
    if (index !== -1) {
      state.users.splice(index, 1);
    } else {
      state.error = "delete user failed, not in list"
    }
  },
  update(state, member) {
    var found = false;
    for (let i = 0; i < state.members.length; i++) {
      let index = state.members[i].members.findIndex(x => x.id === member.id);
      if (index !== -1) {
        state.members[i].members.splice(index, 1);
        state.members[i].members.push(member);
        found = true;
      }
    }
    if (!found) {
      state.error = "update account (member) failed, not in list"
    }
  },
  delete(state, member) {
    var found = false;
    for (let i = 0; i < state.members.length; i++) {
      let index = state.members[i].members.findIndex(x => x.id === member.id);
      if (index !== -1) {
        state.members[i].members.splice(index, 1);
        found = true;
      }
    }
    if (!found) {
      state.error = "delete account (member) failed, not in list"
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
