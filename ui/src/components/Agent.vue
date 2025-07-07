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
        Authenticated
      </v-card-title>
      <v-row>
        <v-col cols="1" sm="1"></v-col>
        <v-col cols="10" class="px-6">
        </v-col>
      </v-row>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="#004000" v-on:click="accept">{{ buttonText }}</v-btn>
        <v-spacer></v-spacer>
      </v-card-actions>
    </v-card>
  </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'Agent',

  data: () => ({
    notification: {},
    id: "",
    valid: false,
    search: '',
    buttonText: 'OK',
  }),

  computed: {
    ...mapGetters({
      account: 'agent/account',

    }),
  },


  mounted() {

/*    const urlParams = new URLSearchParams(window.location.search);
    const newUrl = new URL('com.nettica.agent://callback/agent');
    for (const [key, value] of urlParams.entries()) {
        newUrl.searchParams.append(key, value);
    }
    const url = newUrl.toString();
    // If opened by JavaScript (popup), redirect and close
    if (window.opener || window.name) {
        window.location.replace(url);
        setTimeout(function() { window.close(); }, 500);
    } else {
        window.location.replace(url);
    }
*/
  },

  methods: {
    ...mapActions('agent', {
      activate: 'activate',
    }),

    accept() {

        if (this.buttonText === "Close") {
          window.close();
          return;
        }

        if (this.buttonText === "OK") {
          this.buttonText = "Close";
        }  
        
        const urlParams = new URLSearchParams(window.location.search);
        const newUrl = new URL('com.nettica.agent://callback/agent');
        for (const [key, value] of urlParams.entries()) {
          newUrl.searchParams.append(key, value);
        }
        const url = newUrl.toString();

        // If opened by JavaScript (popup), redirect and close
          window.location.replace(url);
          setTimeout(function() { window.close(); }, 500);
    },

  },

};
</script>
