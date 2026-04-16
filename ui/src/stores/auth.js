import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'
import TokenService from '@/services/token.service'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    error: null,
    user: null,
    authStatus: '',
    clientId: '',
    State: '',
    code: '',
    referer: '',
    authRedirectUrl: '',
    requiresAuth: true,
  }),

  getters: {
    isAuthenticated: (state) => state.user !== null,
  },

  actions: {
    async fetchUser() {
      try {
        const resp = await ApiService.get('/auth/user')
        this.user = resp
      } catch (err) {
        this.error = err
        this.logout()
      }
    },

    async oauth2_url() {
      if (TokenService.getToken()) {
        ApiService.setHeader()
        await this.fetchUser()
        return
      }
      try {
        const resp = await ApiService.get('/auth/oauth2_url')
        if (resp.clientId) {
          this.clientId = resp.clientId
          TokenService.saveClientId(resp.clientId)
        }
        if (resp.codeUrl === '/login') {
          console.log('server report oauth2 is disabled, basic auth')
          this.authStatus = 'redirect'
          this.authRedirectUrl = resp.codeUrl
        } else if (resp.codeUrl === '_magic_string_fake_auth_no_redirect_') {
          console.log('server report oauth2 is disabled, fake exchange')
          this.authStatus = 'disabled'
          await this.oauth2_exchange({ code: '', state: resp.state })
        } else {
          this.authStatus = 'redirect'
          this.authRedirectUrl = resp.codeUrl
        }
      } catch (err) {
        this.authStatus = 'error'
        this.error = err
        this.logout()
      }
    },

    basic_auth() {
      this.authStatus = 'basic'
    },

    async login(data) {
      try {
        const resp = await ApiService.post('/auth/login', data)
        console.log('login', resp)
        this.code = resp.code
        TokenService.saveCode(resp.code)
        this.State = resp.state
        TokenService.saveState(resp.state)
        this.authRedirectUrl = resp.redirect_uri
        this.authStatus = 'redirect'
      } catch (err) {
        console.log('login error', err)
        this.authStatus = 'error'
        this.error = err.response?.data?.error
        this.logout()
      }
    },

    async oauth2_exchange(data) {
      console.log('oauth2_exchange', data)
      if (data.clientId === undefined) {
        data.clientId = TokenService.getClientId()
      }
      if (data.server) {
        TokenService.saveWildServer(data.server)
        ApiService.setWildServer()
      }
      try {
        const resp = await ApiService.post('/auth/oauth2_exchange', data)
        if (data.server) {
          TokenService.saveWildToken(resp)
          ApiService.setServer()
        } else {
          TokenService.saveToken(resp)
          ApiService.setHeader()
          TokenService.destroyClientId()
          await this.fetchUser()
        }
        this.authStatus = 'success'
      } catch (err) {
        console.log('oauth2_exchange error', err)
        if (data.server) {
          TokenService.destroyWildToken()
          TokenService.destroyWildServer()
          ApiService.setServer()
        } else {
          TokenService.destroyToken()
        }
        this.authStatus = 'error'
        this.error = err
        this.logout()
      }
    },

    async logout() {
      try {
        await ApiService.get('/auth/logout')
      } catch (_) {
        // ignore
      } finally {
        this.user = null
        TokenService.destroyToken()
        TokenService.destroyClientId()
        TokenService.destroyWildToken()
        TokenService.destroyWildServer()
        TokenService.destroyServer()
        TokenService.destroyReferer()
        this.authStatus = ''
      }
    },
  },
})
