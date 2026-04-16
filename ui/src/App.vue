<template>
  <v-app>
    <Notification :notification="notification" />
    <Header v-show="authStore.isAuthenticated" />
    <v-main style="padding: 0px">
      <router-view />
    </v-main>
    <Footer />
  </v-app>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTheme } from 'vuetify'
import { storeToRefs } from 'pinia'
import Notification from './components/Notification.vue'
import Header from './components/Header.vue'
import Footer from './components/Footer.vue'
import TokenService from './services/token.service'
import ApiService from './services/api.service'
import { useAuthStore } from './stores/auth'
import { useAccountStore } from './stores/account'
import { useDeviceStore } from './stores/device'
import { useVpnStore } from './stores/vpn'
import { useNetStore } from './stores/net'
import { useServerStore } from './stores/server'
import { useServiceStore } from './stores/service'
import { useSubscriptionStore } from './stores/subscription'

const route = useRoute()
const router = useRouter()
const theme = useTheme()

const authStore = useAuthStore()
const accountStore = useAccountStore()
const deviceStore = useDeviceStore()
const vpnStore = useVpnStore()
const netStore = useNetStore()
const serverStore = useServerStore()
const serviceStore = useServiceStore()
const subscriptionStore = useSubscriptionStore()

const { isAuthenticated, authStatus, authRedirectUrl, requiresAuth } = storeToRefs(authStore)

const notification = ref({ show: false, color: '', text: '', timeout: 10 })

function notify(msg) {
  if (msg == null) return
  notification.value.show = true
  notification.value.text = msg
  notification.value.timeout = 10
}

theme.global.name.value = 'dark'

onMounted(() => {
  if (navigator.userAgent && navigator.userAgent.toLowerCase().includes('dart')) {
    window.close()
    return
  }
  if (route.query.redirect_uri) {
    TokenService.saveRedirect(route.query.redirect_uri)
    TokenService.destroyToken()
  }
  if (route.query.code && route.query.state) {
    const redirect = TokenService.getRedirect()
    if (redirect != null && redirect !== '') {
      TokenService.destroyRedirect()
      const url =
        redirect +
        '?code=' + route.query.code +
        '&state=' + route.query.state +
        '&client_id=' + TokenService.getClientId()
      window.location.replace(url)
      return
    }
  }
  if (route.query.referer) {
    const r = TokenService.getReferer()
    if (r == null) {
      TokenService.saveReferer(route.query.referer)
      console.log('saved referer ', route.query.referer)
      TokenService.destroyToken()
    }
  }
  if (
    route.query.server &&
    route.query.code &&
    route.query.state &&
    route.query.client_id
  ) {
    exchange({
      code: route.query.code,
      state: route.query.state,
      clientId: route.query.client_id,
      server: route.query.server,
    }).catch((err) => {
      console.log('exchange error', err)
    })
    return
  }
  if (authStore.requiresAuth || location.pathname === '/') {
    if (!authStore.isAuthenticated) {
      if (route.query.code && route.query.state) {
        TokenService.saveCode(route.query.code)
        TokenService.saveState(route.query.state)
        const referer = TokenService.getReferer()
        const client_id = TokenService.getClientId()
        if (referer) {
          const url =
            '/consent?referer=' + referer +
            '&client_id=' + client_id +
            '&code=' + route.query.code +
            '&state=' + route.query.state
          router.push(url)
        } else {
          try {
            authStore.oauth2_exchange({
              code: route.query.code,
              state: route.query.state,
            })
            TokenService.destroyReferer()
            TokenService.destroyCode()
            TokenService.destroyState()
            TokenService.destroyClientId()
          } catch (e) {
            notification.value = { show: true, color: 'error', text: e.message, timeout: 10 }
          }
        }
      } else {
        console.log('route.path = %s', route.path)
        if (!location.pathname.startsWith('/join') && !location.pathname.startsWith('/consent')) {
          authStore.oauth2_url()
        }
      }
    }
  }
})

watch(
  () => authStore.error,
  (newValue) => {
    console.log(newValue)
    notify(newValue)
  },
)

watch(
  () => netStore.error,
  (newValue) => {
    notify(newValue)
    netStore.error = null
  },
)

watch(
  () => accountStore.error,
  (newValue) => {
    notify(newValue)
    accountStore.error = null
  },
)

watch(
  () => vpnStore.error,
  (newValue) => {
    notify(newValue)
    vpnStore.error = null
  },
)

watch(
  () => deviceStore.error,
  (newValue) => {
    notify(newValue)
    deviceStore.error = null
  },
)

watch(
  () => serviceStore.error,
  (newValue) => {
    notify(newValue)
    serviceStore.error = null
  },
)

watch(
  () => subscriptionStore.error,
  (newValue) => {
    notify(newValue)
    subscriptionStore.error = null
  },
)

watch(
  () => serverStore.error,
  (newValue) => {
    notify(newValue)
    serverStore.error = null
  },
)

watch(requiresAuth, (newValue, oldValue) => {
  console.log(`Updating requiresAuth from ${oldValue} to ${newValue}`)
})

watch(isAuthenticated, (newValue, oldValue) => {
  console.log(`Updating isAuthenticated from ${oldValue} to ${newValue}`)
  if ((newValue === true && authStore.requiresAuth === true) || location.pathname === '/') {
    router.push('/')
  }
})

watch(authStatus, (newValue, oldValue) => {
  console.log(`Updating authStatus from ${oldValue} to ${newValue}`)
  if (newValue === 'redirect') {
    console.log('redirecting to ', authStore.authRedirectUrl)
    window.location.replace(authStore.authRedirectUrl)
  }
})

function exchange(x) {
  return new Promise((resolve, reject) => {
    TokenService.saveWildServer(x.server)
    ApiService.setWildServer()
    let token
    ApiService.post('/auth/oauth2_exchange', x)
      .then((resp) => {
        console.log('wild exchange successful')
        token = resp
        TokenService.saveWildToken(token)
        ApiService.setServer()
        window.location.replace('/')
        resolve(token)
      })
      .catch((err) => {
        console.log('wild exchange error', err)
        TokenService.destroyWildToken()
        TokenService.destroyWildServer()
        reject(err)
      })
  })
}
</script>
