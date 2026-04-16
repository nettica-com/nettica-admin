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
      <v-card-title>Authenticated</v-card-title>
      <v-row>
        <v-col cols="1" sm="1"></v-col>
        <v-col cols="10" class="px-6"></v-col>
      </v-row>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="#004000" @click="accept">{{ buttonText }}</v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'

const notification = ref({ show: false, color: '', text: '' })
const buttonText = ref('OK')

function accept() {
  if (buttonText.value === 'Close') {
    window.close()
    return
  }
  if (buttonText.value === 'OK') {
    buttonText.value = 'Close'
  }
  const urlParams = new URLSearchParams(window.location.search)
  const newUrl = new URL('com.nettica.agent://callback/agent')
  for (const [key, value] of urlParams.entries()) {
    newUrl.searchParams.append(key, value)
  }
  window.location.replace(newUrl.toString())
  setTimeout(() => window.close(), 500)
}
</script>
