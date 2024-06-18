import Vue from "vue";
import TokenService from "./token.service";

const ApiService = {

  setHeader() {
    Vue.axios.defaults.headers['Authorization'] = `Bearer ${TokenService.getToken()}`;
  },

  setWildHeader() {
    Vue.axios.defaults.headers['Authorization'] = `Bearer ${TokenService.getWildToken()}`;
  },

  setServer() {
    Vue.axios.defaults.baseURL = TokenService.getServer() + "/api/v1.0";
  },

  setWildServer() {
    Vue.axios.defaults.baseURL = TokenService.getWildServer() + "/api/v1.0";
  },

  get(resource) {
    return Vue.axios.get(resource)
      .then(response => response.data)
      .catch(error => {
        throw (error)
      });
  },

  post(resource, params) {
    return Vue.axios.post(resource, params)
      .then(response => response.data)
      .catch(error => {
        throw (error)
      });
  },

  put(resource, params) {
    return Vue.axios.put(resource, params)
      .then(response => response.data)
      .catch(error => {
        throw (error)
      });
  },

  patch(resource, params) {
    return Vue.axios.patch(resource, params)
      .then(response => response.data)
      .catch(error => {
        throw (error)
      });
  },

  delete(resource) {
    return Vue.axios.delete(resource)
      .then(response => response.data)
      .catch(error => {
        throw (error)
      });
  },

  getWithConfig(resource, config) {
    return Vue.axios.get(resource, config)
      .then(response => response.data)
      .catch(error => {
        throw (error)
      });
  },
};

export default ApiService;
