import axios from 'axios'
import TokenService from '../services/token.service'

let baseUrl = '/api/v1.0'
if (import.meta.env.MODE === 'development') {
  baseUrl = import.meta.env.VITE_API_BASE_URL || 'https://dev.nettica.com/api/v1.0'
}

const instance = axios.create({ baseURL: baseUrl })

TokenService.saveServer(baseUrl)

instance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      TokenService.destroyToken()
      TokenService.destroyClientId()
      window.location = '/'
    } else {
      return Promise.reject(error)
    }
  },
)

export default instance
