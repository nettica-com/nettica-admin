import axios from '@/plugins/axios'
import TokenService from './token.service'

const ApiService = {

  setHeader() {
    axios.defaults.headers['Authorization'] = `Bearer ${TokenService.getToken()}`
  },

  setWildHeader() {
    axios.defaults.headers['Authorization'] = `Bearer ${TokenService.getWildToken()}`
  },

  setServer() {
    axios.defaults.baseURL = TokenService.getServer()
  },

  setWildServer() {
    axios.defaults.baseURL = TokenService.getWildServer() + '/api/v1.0'
  },

  get(resource) {
    return axios.get(resource)
      .then((response) => response.data)
      .catch((error) => { throw error })
  },

  post(resource, params) {
    return axios.post(resource, params)
      .then((response) => response.data)
      .catch((error) => { throw error })
  },

  put(resource, params) {
    return axios.put(resource, params)
      .then((response) => response.data)
      .catch((error) => { throw error })
  },

  patch(resource, params) {
    return axios.patch(resource, params)
      .then((response) => response.data)
      .catch((error) => { throw error })
  },

  delete(resource) {
    return axios.delete(resource)
      .then((response) => response.data)
      .catch((error) => { throw error })
  },

  getWithConfig(resource, config) {
    return axios.get(resource, config)
      .then((response) => response.data)
      .catch((error) => { throw error })
  },
}

export default ApiService
