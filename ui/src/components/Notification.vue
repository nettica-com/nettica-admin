<template>
  <v-snackbar v-model="notification.show" location="bottom center" :color="notification.color">
    <v-row>
      <v-col cols="9" class="text-center">
        {{ notification.text }}
      </v-col>
      <v-col cols="3">
        <v-btn variant="text" @click="notification.show = false">close</v-btn>
      </v-col>
    </v-row>
  </v-snackbar>
</template>

<script setup>
defineProps({
  notification: { type: Object, required: true },
})
</script>
