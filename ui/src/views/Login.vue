<template>
  <v-main style="padding-top:74px;">
    <v-container>
      <v-card>
                <v-card-title class="headline">Login</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                                <v-text-field
                                        v-model="username"
                                        label="Username"
                                        :rules="[ v => !!v || 'username is required', ]"
                                        required
                                />
                                <v-text-field
                                        v-model="password"
                                        type="password"
                                        :append-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                                        :type="showPrivate ? 'text' : 'password'"
                                        @click:append="showPrivate = !showPrivate"
                                        label="Password"
                                        :rules="[ v => !!v || 'password is required', ]"
                                        required
                                />

                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer/>
                    <v-btn
                            :disabled="!valid"
                            color="success"
                            @click="login()"
                    >
                        Login
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
    </v-container>

  </v-main>
</template>

<script>
  import {mapActions, mapGetters} from "vuex";
  import TokenService from "../services/token.service";

  export default {
    name: 'Login',


    data: () => ({
      valid: true,
      username: '',
      password: '',
      showPrivate: false,
    }),

    computed:{
      ...mapGetters({
        isAuthenticated: 'auth/isAuthenticated',
        authStatus: 'auth/authStatus',
        authRedirectUrl: 'auth/authRedirectUrl',
        authError: 'auth/error',
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
      //this.basic_auth()
      //this.authStatus = 'basic'
      //this.authRedirectUrl = "/"
      //this.requiresAuth = false

/*      if (this.isAuthenticated == false) {
        if (this.$route.query.code && this.$route.query.state) {
            try {
              console.log("oauth2_exchange")
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
      */
    },

    watch: {

      isAuthenticated(newValue, oldValue) {
        console.log(`login: Updating isAuthenticated from ${oldValue} to ${newValue}`);
        if (newValue === true) {
           this.$router.push('/')
        }
      },

      authStatus(newValue, oldValue) {
        console.log(`login: Updating authStatus from ${oldValue} to ${newValue}`);
        if (newValue === 'redirect') {
          this.authStatus = 'basic_auth'
//          window.location.replace(this.authRedirectUrl)
        }
      },
    },

    methods: {
      ...mapActions('auth', {
        oauth2_exchange: 'oauth2_exchange',
        oauth2_url: 'oauth2_url',
        basic_auth: 'basic_auth',
        basic_login: 'login',
      }),

      login() {

	var clientId = TokenService.getClientId()
	if (this.$route.query.client_id) {
            clientId = this.$route.query.client_id;
        }
        // base64 encode the username and password
        let auth = btoa(this.username + ':' + this.password);
        this.basic_login( {
          clientId: clientId,
          code: auth,
          state: this.$route.query.state,
          redirect_uri: this.$route.query.redirect_uri } )

        //if (this.$route.query.redirect_uri == "com.nettica.agent://callback/agent") {
        //  let response = "com.nettica.agent://callback/agent" + "?code=" + auth + "&state=basic_auth"
        //  console.log("redirecting to %s", response)
        //  window.location.replace(response)
        //}

      },

    }
  };
</script>
