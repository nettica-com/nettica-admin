<template>
  <v-container>
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
    <v-card>
      <v-card-title>
        Join
      </v-card-title>
    </v-card>
    <v-card>
      <p>Joining {{ $route.query.id }}</p>
      <p>Click <a href="/hosts">here</a> to manage your nets</p>
    </v-card>
  </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'Join',

  data: () => ({
    notification: {},
    result: "",
    panel: 1,
    id: "",
    valid: false,
    search: '',
  }),

  computed: {
    ...mapGetters({
      accounts: 'account/accounts',

    }),
  },

  mounted() {
    this.id = this.$route.query.id
    this.result = this.activate(this.$route.query.id)
    this.notification = {
      show: true,
      text: this.$route.query.id + " joined",
      color: "success",
      timeout: 5000,
    }
  },

  methods: {
    ...mapActions('join', {
      activate: 'activate',
    }),

  }
};
</script>
