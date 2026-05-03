import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'
import axios from './plugins/axios'
import { isCidr } from './plugins/cidr'

window.addEventListener('vite:preloadError', () => {
  window.location.reload()
})

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(vuetify)

app.config.globalProperties.$isCidr = isCidr
app.config.globalProperties.$http = axios

app.mount('#app')
