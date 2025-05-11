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
        <v-spacer></v-spacer>
        <v-btn color="#004000" v-on:click="accept">Accept</v-btn>
        <v-btn color="#400000" v-on:click="reject">Reject</v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>
  </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'
import TokenService from "../services/token.service";

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
    //alert(this.$route.query.id + " joined")
  },

  methods: {
    ...mapActions('consent', {
      activate: 'activate',
    }),

    accept() {

      var referer = TokenService.getReferer();
      TokenService.destroyReferer();
      var code = TokenService.getCode();
      TokenService.destroyCode();
      var state = TokenService.getState();
      TokenService.destroyState();
      var client_id = TokenService.getClientId();
      TokenService.destroyClientId();

      if (TokenService.isValidRedirect(referer)) { 
        var url = referer + "/?client_id=" + client_id + "&code=" + code + "&state=" + state + "&server=" + window.location.origin;
        window.location.replace(url);
      }
    },

    reject() {
      TokenService.destroyReferer();
      TokenService.destroyCode();
      TokenService.destroyState();
      TokenService.destroyClientId();
      window.location.replace("/")
    },
  },

};
</script>
