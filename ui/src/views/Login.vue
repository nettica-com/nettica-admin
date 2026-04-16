<template>
  <v-main style="padding-top:74px;">
    <v-container>
      <v-card>
        <v-card-title class="text-h5">Login</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-form ref="formRef" v-model="valid">
                <v-text-field
                  v-model="username"
                  label="Username"
                  :rules="[v => !!v || 'username is required']"
                  required
                  @keyup.enter="login"
                />
                <v-text-field
                  v-model="password"
                  :append-inner-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                  :type="showPrivate ? 'text' : 'password'"
                  label="Password"
                  :rules="[v => !!v || 'password is required']"
                  required
                  @click:append-inner="showPrivate = !showPrivate"
                  @keyup.enter="login"
                />
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn :disabled="!valid" color="success" @click="login">
            Login
            <v-icon end>mdi-check-outline</v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-container>
  </v-main>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import TokenService from '@/services/token.service'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { isAuthenticated, authStatus } = storeToRefs(authStore)

const valid = ref(true)
const username = ref('')
const password = ref('')
const showPrivate = ref(false)

watch(isAuthenticated, (newValue, oldValue) => {
  console.log(`login: Updating isAuthenticated from ${oldValue} to ${newValue}`)
  if (newValue === true) {
    router.push('/')
  }
})

watch(authStatus, (newValue, oldValue) => {
  console.log(`login: Updating authStatus from ${oldValue} to ${newValue}`)
  if (newValue === 'redirect') {
    authStore.authStatus = 'basic_auth'
  }
})

function login() {
  let clientId = TokenService.getClientId()
  if (route.query.client_id) {
    clientId = route.query.client_id
  }
  const auth = btoa(username.value + ':' + password.value)
  authStore.login({
    clientId,
    code: auth,
    state: route.query.state,
    redirect_uri: route.query.redirect_uri,
  })
}
</script>
