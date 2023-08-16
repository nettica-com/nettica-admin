import Vue from 'vue'
import Vuex from 'vuex'
import auth from "./modules/auth";
import host from "./modules/host";
import server from "./modules/server";
import device from "./modules/device";
import net from "./modules/net";
import vpn from "./modules/vpn"
import user from "./modules/users"
import account from "./modules/account"
import join from "./modules/join"
import subscription from "./modules/subscription"
import service from "./modules/service"

Vue.use(Vuex)

export default new Vuex.Store({
  state: {},
  getters: {},
  mutations: {},
  actions: {},
  modules: {
    account,
    auth,
    host,
    device,
    net,
    vpn,
    user,
    join,
    subscription,
    service,
    server
  }
})
