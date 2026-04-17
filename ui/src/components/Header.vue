<template>
  <div>
    <v-app-bar app>
      <a href="https://nettica.com">
        <img class="mr-3" src="/nettica-3d-256.png" height="50" width="50" alt="nettica" />
      </a>
      <v-toolbar-title>{{ title }}</v-toolbar-title>

      <v-spacer />

      <!-- Desktop nav -->
      <v-toolbar-items class="d-none d-sm-flex">
        <v-btn :to="'/services'" v-show="showServices" title="services">
          Services
          <v-icon end>mdi-weather-cloudy</v-icon>
        </v-btn>
        <v-btn :to="'/networks'" title="networks">
          Networks
          <span class="material-symbols-outlined">hub</span>
        </v-btn>
        <v-btn :to="'/devices'" title="devices">
          Devices
          <v-icon end>mdi-devices</v-icon>
        </v-btn>
        <v-btn :to="'/accounts'" title="account">
          Account
          <v-icon end>mdi-account-group</v-icon>
        </v-btn>
      </v-toolbar-items>

      <v-menu location="bottom left" class="d-none d-sm-flex">
        <template #activator="{ props: menuProps }">
          <v-btn icon v-bind="menuProps">
            <v-avatar size="36">
              <img :src="picture" height="36" width="36" :alt="name" />
            </v-avatar>
          </v-btn>
        </template>
        <v-card class="mx-auto" max-width="344" variant="outlined">
          <v-list-item v-show="authStore.isAuthenticated">
            <template #prepend>
              <v-avatar size="64">
                <img :alt="name" :src="picture" height="64" width="64" />
              </v-avatar>
            </template>
            <v-list-item-title class="text-h6 mb-1">{{ name }}</v-list-item-title>
            <v-list-item-subtitle>connected as</v-list-item-subtitle>
            <v-list-item-subtitle>Email: {{ email }}</v-list-item-subtitle>
            <v-list-item-subtitle>Issuer: {{ issuer }}</v-list-item-subtitle>
            <v-list-item-subtitle>Issued at: {{ formatDate(issuedAt) }}</v-list-item-subtitle>
          </v-list-item>
          <v-card-actions>
            <v-btn size="small" @click="mylogout">
              logout
              <v-icon end size="small">mdi-logout</v-icon>
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-menu>

      <!-- Mobile hamburger -->
      <v-btn icon class="d-flex d-sm-none" @click="myShowMenu">
        <v-icon>mdi-menu</v-icon>
      </v-btn>
    </v-app-bar>

    <div v-if="friendly" style="width:100%; position:absolute;">
      <div style="height:64px; width:100%;"></div>
      <v-alert type="info" color="#336699">
        Welcome to the Admin! Click on the menu above to add service, create networks, add devices, and invite others to your account.
      </v-alert>
    </div>

    <v-navigation-drawer v-model="showMenu" app location="right" temporary>
      <v-list nav density="compact">
        <v-list-item
          v-show="showServices"
          prepend-icon="mdi-weather-cloudy"
          title="Services"
          value="services"
          to="/services"
        />
        <v-list-item title="Networks" value="networks" to="/networks">
          <template #prepend>
            <span class="material-symbols-outlined" style="display:inline-block; font-size:24px; width:24px; height:24px; line-height:24px; margin-right:32px; opacity:0.87">hub</span>
          </template>
        </v-list-item>
        <v-list-item prepend-icon="mdi-devices" title="Devices" value="devices" to="/devices" />
        <v-list-item prepend-icon="mdi-account-group" title="Accounts" value="accounts" to="/accounts" />
      </v-list>
      <template #append>
        <v-divider />
        <v-list nav density="compact">
          <v-list-item>
            <template #prepend>
              <v-avatar size="36" class="mr-2">
                <img :src="picture" height="36" width="36" :alt="name" />
              </v-avatar>
            </template>
            <v-list-item-title>{{ name }}</v-list-item-title>
            <v-list-item-subtitle>{{ email }}</v-list-item-subtitle>
          </v-list-item>
          <v-list-item prepend-icon="mdi-logout" title="Logout" @click="mylogout" />
        </v-list>
      </template>
    </v-navigation-drawer>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useDisplay } from 'vuetify'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/formatDate'
import env from '../../env.json'

const route = useRoute()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)
const { smAndUp } = useDisplay()

const title = ref(document.location.host)
const showMenu = ref(false)
const showServices = ref(env.showServicesTab)
const name = ref('')
const picture = ref('')
const email = ref('')
const issuer = ref('')
const issuedAt = ref('')
const friendly = ref(false)

onMounted(() => {
  document.title = document.location.host
})

watch(user, (val) => {
  if (!val) return
  name.value = val.name
  picture.value = val.picture
  email.value = val.email
  issuer.value = val.issuer
  issuedAt.value = val.issuedAt
})

watch(smAndUp, (val) => {
  if (val) showMenu.value = false
})

watch(
  () => route.path,
  (path) => {
    friendly.value = path === '/'
    showMenu.value = false
  },
)

function mylogout() {
  authStore.logout()
  window.location.href = '/api/v1.0/auth/logout'
}

function myShowMenu() {
  showMenu.value = !showMenu.value
  console.log('showMenu = ' + showMenu.value)
}
</script>
