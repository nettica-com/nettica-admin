<template>
  <v-container>
    <v-snackbar v-model="notification.show" location="bottom center" :color="notification.color">
      <v-row>
        <v-col cols="9" class="text-center">{{ notification.text }}</v-col>
        <v-col cols="3">
          <v-btn variant="text" @click="notification.show = false">close</v-btn>
        </v-col>
      </v-row>
    </v-snackbar>
    <v-card>
      <v-card-title>Consent</v-card-title>
      <v-row>
        <v-col cols="1" sm="1"></v-col>
        <v-col cols="10" class="px-6">
          <v-text-field label="Server Name" :model-value="route.query.referer" readonly />
          <p>
            Do you want to accept this connection? The server will have the same privileges as your
            account. You can revoke this connection at any time by logging out of the server.
          </p>
        </v-col>
      </v-row>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="#004000" @click="accept">Accept</v-btn>
        <v-btn color="#400000" @click="reject">Reject</v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import TokenService from '@/services/token.service'

const route = useRoute()
const notification = ref({ show: false, color: '', text: '' })

function accept() {
  const referer = TokenService.getReferer()
  TokenService.destroyReferer()
  const code = TokenService.getCode()
  TokenService.destroyCode()
  const state = TokenService.getState()
  TokenService.destroyState()
  const client_id = TokenService.getClientId()
  TokenService.destroyClientId()

  if (TokenService.isValidRedirect(referer)) {
    const url =
      referer + '/?client_id=' + client_id + '&code=' + code + '&state=' + state + '&server=' + window.location.origin
    window.location.replace(url)
  }
}

function reject() {
  TokenService.destroyReferer()
  TokenService.destroyCode()
  TokenService.destroyState()
  TokenService.destroyClientId()
  window.location.replace('/')
}
</script>
