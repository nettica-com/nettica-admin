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
      <v-card-title>Join Nettica Network</v-card-title>
      <v-row>
        <v-col cols="1" sm="1"></v-col>
        <v-col cols="10" class="px-6">
          <v-text-field label="Account Name" :model-value="joinStore.account.accountName" readonly />
          <v-text-field label="Account ID" :model-value="joinStore.account.parent" readonly />
          <v-text-field label="Name" :model-value="joinStore.account.name" readonly />
          <v-text-field label="Role" :model-value="joinStore.account.role" readonly />
          <v-text-field label="Email" :model-value="joinStore.account.email" readonly />
          <v-text-field label="Invitation From" :model-value="joinStore.account.createdBy" readonly />
          <v-text-field label="Network" :model-value="joinStore.account.netName" readonly />
          <v-text-field label="Status" :model-value="joinStore.account.status" readonly />
          <v-btn color="#000040" @click="login">Login</v-btn>
        </v-col>
      </v-row>
      <v-card-actions></v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useJoinStore } from '@/stores/join'

const route = useRoute()
const joinStore = useJoinStore()
const notification = ref({ show: false, color: '', text: '' })

onMounted(() => {
  joinStore.activate(route.query.id)
  notification.value = {
    show: true,
    text: route.query.id + ' joined',
    color: 'success',
    timeout: 5000,
  }
})

function login() {
  location.replace('/')
}
</script>
