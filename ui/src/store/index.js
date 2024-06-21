import Vue from 'vue'
import Vuex from 'vuex'
import auth from "./modules/auth";
import consent from "./modules/consent";
import server from "./modules/server";
import device from "./modules/device";
import net from "./modules/net";
import wildnet from "./modules/wildnet";
import vpn from "./modules/vpn"
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
    consent,
    device,
    net,
    wildnet,
    vpn,
    join,
    subscription,
    service,
    server
  }
})
