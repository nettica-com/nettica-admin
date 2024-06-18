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
        Consent
      </v-card-title>
      <v-row>
        <v-col cols="1" sm="1"></v-col>
        <v-col cols="10" class="px-6">
          <v-text-field label="Server Name" v-model="$route.query.referer" readonly></v-text-field>
          <p>Do you want to accept this connection?  The server will have the same
             privileges as your account.  You can revoke this connection at any time by
             logging out of the server.
          </p>
        </v-col>
      </v-row>
      <v-card-actions>
        <v-btn color="success" v-on:click="accept">Accept</v-btn>
        <v-btn color="error" v-on:click="reject">Reject</v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'Consent',

  data: () => ({
    notification: {},
    id: "",
    valid: false,
    search: '',
  }),

  computed: {
    ...mapGetters({
      account: 'consent/account',

    }),
  },


  mounted() {
    this.id = this.$route.query.id
    this.activate(this.$route.query.id)
    this.notification = {
      show: true,
      text: this.$route.query.id + " joined",
      color: "success",
      timeout: 5000,
    }
    //alert(this.$route.query.id + " joined")
  },

  methods: {
    ...mapActions('consent', {
      activate: 'activate',
    }),

    accept() {
      var url = this.$router.query.referer + "/" + this.$router.query + "&server=" + this.window.location.hostname
      location.replace(url)
    },

    reject() {
      location.replace(this.$router.query.referer)
    },
  },

};
</script>
