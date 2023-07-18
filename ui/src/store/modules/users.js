import ApiService from "../../services/api.service";

const state = {
  error: null,
  users: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  users(state) {
    return state.users;
  },
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  readAll({ commit, dispatch }){
    ApiService.get("/users")
      .then(resp => {
        commit('users', resp)
//        dispatch('readQrcodes')
//        dispatch('readConfigs')
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, host){
    ApiService.post("/users", host)
      .then(resp => {
//        dispatch('readQrcode', resp)
        dispatch('readConfig', resp)
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, user){
    ApiService.patch(`/users/${user.id}`, user)
      .then(resp => {
//        dispatch('readQrcode', resp)
//        dispatch('readConfig', resp)
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, user){
    ApiService.delete(`/users/${user.id}`)
      .then(() => {
        commit('delete', user)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  email({ commit }, toAddress){
    ApiService.get(`/users/${toAddress}/invite`)
      .then(() => {
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcode({ state, commit }, host){
/*    ApiService.getWithConfig(`/host/${host.id}/config?qrcode=true&alan=false`, {responseType: 'arraybuffer'})
      .then(resp => {
        let image = Buffer.from(resp, 'binary').toString('base64')
        commit('hostQrcodes', { host, image })
      })
      .catch(err => {
        commit('error', err)
      })*/
  },

  readConfig({ state, commit }, host){
    ApiService.getWithConfig(`/host/${host.id}/config?qrcode=false`, {responseType: 'arraybuffer'})
      .then(resp => {
        commit('hostConfigs', { host: host, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readQrcodes({ state, dispatch }){
    state.hosts.forEach(host => {
//      dispatch('readQrcode', host)
    })
  },

  readConfigs({ state, dispatch }){
    state.hosts.forEach(host => {
      dispatch('readConfig', host)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  users(state, users){
    state.users = users
  },
  create(state, user){
    state.users.push(user)
  },
  update(state, user){
    let index = state.users.findIndex(x => x.id === user.id);
    if (index !== -1) {
      state.users.splice(index, 1);
      state.users.push(user);
    } else {
      state.error = "update host failed, not in list"
    }
  },
  delete(state, user){
    let index = state.users.findIndex(x => x.id === user.id);
    if (index !== -1) {
      state.hosts.splice(index, 1);
    } else {
      state.error = "delete user failed, not in list"
    }
  },
  hostQrcodes(state, { host, image }){
    let index = state.hostQrcodes.findIndex(x => x.id === host.id);
    if (index !== -1) {
      state.hostQrcodes.splice(index, 1);
    }
    state.hostQrcodes.push({
      id: host.id,
      qrcode: image
    })
  },
  hostConfigs(state, { host, config }){
    let index = state.hostConfigs.findIndex(x => x.id === host.id);
    if (index !== -1) {
      state.hostConfigs.splice(index, 1);
    }
    state.hostConfigs.push({
      id: host.id,
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
