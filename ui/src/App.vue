<template>
  <v-app id="inspire">
    <v-layout>
    <Notification v-bind:notification="notification"/>
      <Header v-show="isAuthenticated" />
        <v-main style="padding: 0px">
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
        vpnError: 'vpn/error',
        deviceError: 'device/error',
        netError: 'net/error',
        accountError: 'account/error',
        serverError: 'server/error',
        serviceError: 'service/error',
        subscriptionError: 'subscription/error',
      })
    },

    created () {
      this.$vuetify.theme.dark = true;
    },

    mounted() {
      if (this.requiresAuth || location.pathname == "/") {
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
            if (!location.pathname.startsWith("/join")) {
              //alert("this.$route.path = " + this.$route.path + "location.pathname=" + location.pathname)
              this.oauth2_url()
            }
          }
        }
      }
    },

    watch: {
      authError(newValue, oldValue) {
        console.log(newValue)
        this.notify(newValue);
      },

      clientError(newValue, oldValue) {
        console.log(newValue)
        this.notify(newValue);
      },

      netError(newValue, oldValue) {
        this.notify(newValue);
        this.errorNet(null)
      },

      accountError(newValue) {
        this.notify(newValue);
        this.errorAccount(null)
      },

      vpnError(newValue, oldValue) {
        this.notify(newValue);
        this.errorVpn(null)
      },

      deviceError(newValue, oldValue) {
        this.notify(newValue);
        this.errorDevice(null)
      },

      serviceError(newValue, oldValue) {
        this.notify(newValue);
        this.errorService(null)
      },

      subscriptionError(newValue, oldValue) {
        this.notify(newValue);
        this.errorSubscription(null)
      },

      serverError(newValue, oldValue) {
        this.notify(newValue);
        this.errorServer(null)
      },
      requiresAuth(newValue, oldValue) {
        console.log(`Updating requiresAuth from ${oldValue} to ${newValue}`);
      },

      isAuthenticated(newValue, oldValue) {
        console.log(`Updating isAuthenticated from ${oldValue} to ${newValue}`);
        if (newValue === true  && this.requiresAuth === true || location.pathname == "/") {
          //alert("isAuthenticated = " + newValue + " requiresAuth = " + this.requiresAuth)
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
      ...mapActions('account', {
        errorAccount: 'error',
      }),
      ...mapActions('device', {
        errorDevice: 'error',
      }),
      ...mapActions('vpn', {
        errorVpn: 'error',
      }),
      ...mapActions('net', {
        errorNet: 'error',
      }),
      ...mapActions('client', {
        errorClient: 'error',
      }),
      ...mapActions('server', {
        errorServer: 'error',
      }),
      ...mapActions('service', {
        errorService: 'error',
      }),
      ...mapActions('subscription', {
        errorSubscription: 'error',
      }),

      notify(msg) {
        if (msg == null) {
          return;
        }
        this.notification.show = true;
        this.notification.text = msg;
        this.notification.timeout = 10;
      }
    }
  };
</script>
