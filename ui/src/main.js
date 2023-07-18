import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import vuetify from './plugins/vuetify';
import './plugins/moment';
import './plugins/cidr'
import './plugins/axios'
import D3Network from 'vue-d3-network'
  components: {
    D3Network
  }

// Don't warn about using the dev version of Vue in development.
Vue.config.productionTip = process.env.NODE_ENV === 'production'
Vue.use(D3Network)
Vue.component('d3-network', D3Network);

new Vue({
  router,
  store,
  vuetify,
  D3Network,
  render: function (h) { return h(App) }
}).$mount('#app')
