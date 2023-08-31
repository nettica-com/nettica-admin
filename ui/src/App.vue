<template>
  <v-app id="inspire">
    <v-layout>
    <Notification v-bind:notification="notification"/>
      <Header v-show="isAuthenticated" />

        <v-main style="padding: 0px 0px 0px 0px;">
          <router-view />
        </v-main>

      <Footer/>
    </v-layout>
  </v-app>
</template>

<script>
  import Notification from './components/Notification'
  import Header from "./components/Header";
  import MainMenu from "./components/Menu";
  import Footer from "./components/Footer";
  import {mapActions, mapGetters} from "vuex";

  export default {
    name: 'App',

    components: {
      Footer,
      Header,
      MainMenu,
      Notification
    },

    data: () => ({
      notification: {
        show: false,
        color: '',
        text: '',
      },
    }),

    computed:{
      ...mapGetters({
        isAuthenticated: 'auth/isAuthenticated',
        authStatus: 'auth/authStatus',
        authRedirectUrl: 'auth/authRedirectUrl',
        requiresAuth: 'auth/requiresAuth',
        authError: 'auth/error',
        clientError: 'host/error',
        netError: 'net/error',
        serverError: 'server/error',
        serviceError: 'service/error',
        subscriptionError: 'subscription/error',
      })
    },

    created () {
      this.$vuetify.theme.dark = true;
    },

    mounted() {
      if (this.requiresAuth || this.$route.path == "/") {
        if (this.isAuthenticated == false) {
          if (this.$route.query.code && this.$route.query.state) {
              try {
                this.oauth2_exchange({
                  code: this.$route.query.code,
                  state: this.$route.query.state
              })
            } catch (e) {
              this.notification = {
                show: true,
                color: 'error',
                text: e.message,
              }
            }
          } else {
            console.log("this.$route.path = %s", this.$route.path);
            if (this.$route.path != "/join") {
              this.oauth2_url()
            }
          }
        }
      }
    },

    watch: {
      authError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      clientError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      netError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      serviceError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      subscriptionError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      serverError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },
      requiresAuth(newValue, oldValue) {
        console.log(`Updating requiresAuth from ${oldValue} to ${newValue}`);
      },

      isAuthenticated(newValue, oldValue) {
        console.log(`Updating isAuthenticated from ${oldValue} to ${newValue}`);
        if (newValue === true  && this.requiresAuth === true) {
          alert("isAuthenticated = " + newValue + " requiresAuth = " + this.requiresAuth)
           this.$router.push('/').catch(err => {
            if (err.name != "NavigationDuplicated") {
              throw err;
              }
          })
        }
      },

      authStatus(newValue, oldValue) {
        console.log(`Updating authStatus from ${oldValue} to ${newValue}`);
        if (newValue === 'redirect') {
          window.location.replace(this.authRedirectUrl)
        }
      },
    },

    methods: {
      ...mapActions('auth', {
        oauth2_exchange: 'oauth2_exchange',
        oauth2_url: 'oauth2_url',
      }),

      notify(color, msg) {
        this.notification.show = true;
        this.notification.color = color;
        this.notification.text = msg;
        this.notification.timeout = 10;
      }
    }
  };
</script>
