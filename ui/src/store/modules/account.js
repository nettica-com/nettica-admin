import ApiService from "../../services/api.service";

const state = {
  error: null,
  account: null,
  accounts: [],
  users: [],
  members: [],
  limits: [],
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
  limits(state) {
    return state.limits;
  },
  getMembers: (state) => (id) => {
    let item = state.members.find(item => item.id === id)
    return item ? item.members : null
  },
  getLimits: (state) => (id) => {
    let item = state.limits.find(item => item.id === id)
    return item ? item.limits : null
  },

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
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
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
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  readLimits({ commit, dispatch }, id) {
    ApiService.get(`/accounts/${id}/limits`)
      .then(resp => {
        console.log( "readLimits: ", resp)
        commit('limits', { id: id, limits: resp })
      })
      .catch(error => {
        if (error.response) {
          // don't report this error, it's expected if there are no limits
          // commit('error', error.response.data.error)
        }
      })
  },

  create({ commit, dispatch }, account) {
    ApiService.post(`/accounts/`, account)
      .then(resp => {
        commit('account', resp)
        commit('error', `Account for ${account.email} created`)
      }) 
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  update({ commit, dispatch }, account) {
    console.log("action update account: ", account)
    ApiService.patch(`/accounts/${account.id}`, account)
      .then(resp => {
        commit('update_member', resp)
        commit('update_account', resp)
        commit('error', `${account.email} updated`)
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
        commit('error', `${account.email} deleted`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
      })
  },

  email({ commit }, account) {
    ApiService.get(`/accounts/${account.id}/invite`)
      .then(() => {
        commit('error', `Email to ${account.email} sent`)
      })
      .catch(error => {
        if (error.response) {
          commit('error', error.response.data.error)
        }
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
  limits(state, { id , limits }) {
    let index = state.limits.findIndex(x => x.id === id);
    if (index !== -1) {
      state.limits.splice(index, 1);
    }
    state.limits.push({
      id: id,
      limits: limits
    })
  },
  create(state, account) {
    state.accounts.push(account)
  },
  create(state, user) {
    state.users.push(user)
  },
  update_account(state, account) {
    console.log( "update_account: ", account)
    if (account.id !== account.parent) {
      console.log( "not a root account, skipping")
    } else {
      let index = state.accounts.findIndex(x => x.id === account.id);
      if (index !== -1) {
        state.accounts.splice(index, 1);
        state.accounts.push(account);
      } else {
        state.error = "update account failed, not in list"
      }
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
  update_user(state, user) {
    console.log( "update user: ", user)
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
  update_member(state, member) {
    console.log( "update_member: ", member)
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
  update(state, limit) {
    let index = state.limits.findIndex(x => x.id === limit.id);
    if (index !== -1) {
      state.limits.splice(index, 1);
      state.limits.push(limit);
    } else {
      state.error = "update account (limit) failed, not in list"
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
