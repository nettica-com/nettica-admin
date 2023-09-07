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
        Join Nettica Network
      </v-card-title>
      <v-row>
        <v-col cols="1" sm="1"></v-col>
        <v-col cols="10" class="px-6">
          <v-text-field label="Account Name" v-model="account.accountName" readonly></v-text-field>
          <v-text-field label="Account ID" v-model="account.parent" readonly></v-text-field>
          <v-text-field label="Name" v-model="account.name" readonly></v-text-field>
          <v-text-field label="Role" v-model="account.role" readonly></v-text-field>
          <v-text-field label="Email" v-model="account.email" readonly></v-text-field>
          <v-text-field label="Invitation From" v-model="account.createdBy" readonly></v-text-field>
          <v-text-field label="Network" v-model="account.netName" readonly></v-text-field>
          <v-text-field label="Status" v-model="account.status" readonly></v-text-field>
          <v-btn color="primary" v-on:click="login">Login</v-btn>
        </v-col>
      </v-row>
      <v-card-actions>
      </v-card-actions>
    </v-card>
  </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'Join',

  data: () => ({
    notification: {},
    id: "",
    valid: false,
    search: '',
  }),

  computed: {
    ...mapGetters({
      account: 'join/account',

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
    ...mapActions('join', {
      activate: 'activate',
    }),

    login() {
      location.replace("/")
    },
  },

};
</script>
