<template>
  <v-snackbar v-model="notification.show" :center="true" :bottom="true" :color="notification.color">
      <v-row>
          <v-col cols="9" class="text-center">
              {{ notification.text }}
          </v-col>
          <v-col cols="3">
              <v-btn text @click="notification.show = false">close</v-btn>
          </v-col>
      </v-row>
  </v-snackbar>
</template>
<script>
export default {
  name: 'Notification',
  props: ['notification'],
  data: () => ({
  }),
};
</script>
