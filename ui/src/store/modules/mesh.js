import ApiService from "../../services/api.service";

const state = {
  error: null,
  meshes: [],
}

const getters = {
  error(state) {
    return state.error;
  },
  meshes(state) {
    return state.meshes;
  },
  getMeshConfig: (state) => (id) => {
    let item = state.meshes.find(item => item.id === id)
    return item ? item.config : null
  }
}

const actions = {
  error({ commit }, error){
    commit('error', error)
  },

  readAll({ commit, dispatch }){
    ApiService.get("/mesh")
      .then(resp => {
        commit('meshes', resp)
//        dispatch('readMeshConfigs')
      })
      .catch(err => {
        commit('error', err)
      })
  },

  create({ commit, dispatch }, mesh){
    ApiService.post("/mesh", mesh)
      .then(resp => {
//        dispatch('readMeshConfig', resp)
        commit('create', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit, dispatch }, mesh){
    ApiService.patch(`/mesh/${mesh.id}`, mesh)
      .then(resp => {
//        dispatch('readMeshConfig', resp)
        commit('update', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  delete({ commit }, mesh){
    ApiService.delete(`/mesh/${mesh.id}`)
      .then(() => {
        commit('delete', mesh)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readMeshConfig({ state, commit }, mesh){
    ApiService.getWithConfig(`/mesh/${mesh.id}`, {responseType: 'arraybuffer'})
      .then(resp => {
//        commit('meshes', { mesh: mesh, config: resp })
      })
      .catch(err => {
        commit('error', err)
      })
  },

  readMeshConfigs({ state, dispatch }){
    state.meshes.forEach(mesh => {
      dispatch('readMeshConfig', mesh)
    })
  },
}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  meshes(state, meshes){
    state.meshes = meshes
  },
  create(state, mesh){
    state.meshes.push(mesh)
  },
  update(state, mesh){
    let index = state.meshes.findIndex(x => x.id === mesh.id);
    if (index !== -1) {
      state.meshes.splice(index, 1);
      state.meshes.push(mesh);
    } else {
      state.error = "update mesh failed, not in list"
    }
  },
  delete(state, mesh){
    let index = state.meshes.findIndex(x => x.id === mesh.id);
    if (index !== -1) {
      state.meshes.splice(index, 1);
    } else {
      state.error = "delete mesh failed, not in list"
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
