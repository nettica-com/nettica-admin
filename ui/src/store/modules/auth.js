import ApiService from "../../services/api.service";
import TokenService from "../../services/token.service";

const state = {
  error: null,
  user: null,
  authStatus: '',
  clientId: '',
  State: '',
  code: '',
  referer: '',
  authRedirectUrl: '',
  requiresAuth: true,
};

const getters = {
  error(state) {
    return state.error;
  },
  user(state) {
    return state.user;
  },
  isAuthenticated(state) {
    return state.user !== null;
  },
  requiresAuth(state) {
    return state.requiresAuth;
  },
  authRedirectUrl(state) {
    return state.authRedirectUrl
  },
  authStatus(state) {
    return state.authStatus
  },
  clientId(state) {
    return state.clientId
  },
  State(state) {
    return state.State
  },
  code(state) {
    return state.code
  },
  referer(state) {
    return state.referer
  },
};

const actions = {
  user({ commit }) {
    ApiService.get("/auth/user")
      .then(resp => {
        commit('user', resp)
      })
      .catch(err => {
        commit('error', err);
        commit('logout')
      });
  },

  oauth2_url({ commit, dispatch }) {
    if (TokenService.getToken()) {
      ApiService.setHeader();
      dispatch('user');
      return
    }
    ApiService.get("/auth/oauth2_url")
      .then(resp => {
  	    if (resp.clientId) {
          commit('clientId', resp.clientId)
        }
        if (resp.codeUrl === '/login') {
          console.log("server report oauth2 is disabled, basic auth")
          commit('authStatus', 'redirect')
          commit('authRedirectUrl', resp.codeUrl)
        } else if (resp.codeUrl === '_magic_string_fake_auth_no_redirect_') {
          console.log("server report oauth2 is disabled, fake exchange")
          commit('authStatus', 'disabled')
          dispatch('oauth2_exchange', { code: "", state: resp.state })
        } else {
          commit('authStatus', 'redirect')
          commit('authRedirectUrl', resp.codeUrl)
        }
      })
      .catch(err => {
        commit('authStatus', 'error')
        commit('error', err);
        commit('logout')
      })
  },

  basic_auth({ commit }) {
    commit('authStatus', 'basic')

  },

  login({ commit }, data) {
    ApiService.post("/auth/login", data)
      .then(resp => {
        // The result of this call is a redirect
        // with code and state in the query string
        console.log( "login", resp);
        commit('code', resp.code);
        commit('state', resp.state);
      	commit('authRedirectUrl', resp.redirect_uri );
        commit('authStatus', 'redirect');

      })
      .catch(err => {
        console.log("login error", err);
        commit('authStatus', 'error')
        commit('error', err.response.data.error);
        commit('logout')
      })
  },

  oauth2_exchange({ commit, dispatch }, data) {
    console.log("oauth2_exchange", data);

    if (data.clientId === undefined) {
      data.clientId = TokenService.getClientId()
    }

    if (data.server) {
      TokenService.saveWildServer(data.server)
      ApiService.setWildServer()
    }

    ApiService.post("/auth/oauth2_exchange", data)
      .then(resp => {
        if (data.server) {
          TokenService.saveWildToken(resp)
          ApiService.setServer()  // reset to server
        } else {
          TokenService.saveToken(resp)
          commit('token', resp)
          dispatch('user');
        }
        commit('authStatus', 'success')
      })
      .catch(err => {
        console.log("oauth2_exchange error", err);
        if (data.server) {
          TokenService.destroyWildToken()
          TokenService.destroyWildServer()
          ApiService.setServer()  // reset to server
        } else {
          TokenService.destroyToken()
        }
        commit('authStatus', 'error')
        commit('error', err);
        commit('logout')
      })
  },

  logout({ commit }) {
    ApiService.get("/auth/logout")
      .then(resp => {
        commit('logout')
      })
      .catch(err => {
        commit('authStatus', '')
        commit('error', err);
        commit('logout')
      })
  },



}

const mutations = {
  error(state, error) {
    state.error = error;
  },
  authStatus(state, authStatus) {
    state.authStatus = authStatus;
  },
  requiresAuth(state, requiresAuth) {
    state.requiresAuth = requiresAuth;
  },
  authRedirectUrl(state, url) {
    state.authRedirectUrl = url;
  },
  token(state, token) {
    TokenService.saveToken(token);
    ApiService.setHeader();
    TokenService.destroyClientId();
  },
  user(state, user) {
    state.user = user;
  },
  logout(state) {
    state.user = null;
    TokenService.destroyToken();
    TokenService.destroyClientId();
    TokenService.destroyWildToken();
    TokenService.destroyWildServer();
    TokenService.destroyServer();
    TokenService.destroyReferer();
  },
  clientId(state, clientId) {
    state.clientId = clientId;
    TokenService.saveClientId(clientId);
  },
  State(state, s) {
    state.State = s;
    TokenService.saveState(s);
  },
  code(state, code) {
    state.code = code;
    TokenService.saveCode(code);
  },
  referer(state, referer) {
    state.referer = referer;
    TokenService.saveReferer(referer);
  },

};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
