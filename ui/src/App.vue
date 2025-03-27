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
  import TokenService from "./services/token.service";
  import ApiService from "./services/api.service";
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
      if (this.$route && this.$route.query && this.$route.query.redirect_uri) {
        TokenService.saveRedirect(this.$route.query.redirect_uri)
        TokenService.destroyToken() // force a token exchange
      }
      if (this.$route && this.$route.query && this.$route.query.code && this.$route.query.state) {
        
        let redirect = TokenService.getRedirect();
        if (redirect != null && redirect != "") {
          TokenService.destroyRedirect()
          var url = redirect + "?code=" + this.$route.query.code + "&state=" + this.$route.query.state + "&client_id=" + TokenService.getClientId();
            window.location.replace(url);
            return;
        }
      }
      if (this.$route && this.$route.query && this.$route.query.referer) {
        let r = TokenService.getReferer()
        if (r == null) {
          TokenService.saveReferer(this.$route.query.referer)
	  console.log("saved referer ", this.$route.query.referer );
          TokenService.destroyToken() // force a token exchange
	  this.isAuthenticated = false
        }
      } 
      if (this.$route && this.$route.query && this.$route.query.server &&
          this.$route.query.code && this.$route.query.state && this.$route.query.client_id) {
        exchange({
          code: this.$route.query.code,
          state: this.$route.query.state,
          clientId: this.$route.query.client_id,
          server: this.$route.query.server
        }).catch(err => {
          console.log("exchange error", err);
        });

        return;
      }
      if (this.requiresAuth || location.pathname == "/") {
        if (this.isAuthenticated == false) {
          if (this.$route.query.code && this.$route.query.state) {

              TokenService.saveCode(this.$route.query.code)
              TokenService.saveState(this.$route.query.state)

              var referer = TokenService.getReferer()
              var client_id = TokenService.getClientId()
              if (referer) {
                var url = "/consent?referer=" + referer + "&client_id=" + client_id + "&code=" + this.$route.query.code + "&state=" + this.$route.query.state;
                this.$router.push(url).catch(err => {
                  if (err.name != "NavigationDuplicated") {
                    throw err;
                  }
                })
              } else {
                try {
                  this.oauth2_exchange({
                    code: this.$route.query.code,
                    state: this.$route.query.state
                })
                TokenService.destroyReferer();
                TokenService.destroyCode();
                TokenService.destroyState();
                TokenService.destroyClientId();
              } catch (e) {
                this.notification = {
                  show: true,
                  color: 'error',
                  text: e.message,
                }
              }
            }
          } else {
            console.log("this.$route.path = %s", this.$route.path);
            if (!location.pathname.startsWith("/join") &&
                !location.pathname.startsWith("/consent")) {
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
	  console.log('redirecting to ', this.authRedirectUrl );
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

  function exchange(x) {
    return new Promise((resolve, reject) => {
      // this oauth2_exchange is strictly for the wilderness
      // this will not screw up the current server's login state
      
      TokenService.saveWildServer(x.server)
      ApiService.setWildServer()
      var token;
      ApiService.post("/auth/oauth2_exchange", x)
      .then(resp => {
          console.log("wild exchange successful")
          token = resp
          TokenService.saveWildToken(token)
          ApiService.setServer()  // reset to server
          window.location.replace("/"); // <-- This is messy

      })
      .catch(err => {
        console.log("wild exchange error", err);
        TokenService.destroyWildToken()
        TokenService.destroyWildServer()
        reject(err);
      });

      resolve(token);
    });
  }
</script>
