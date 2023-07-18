<template>
    <v-container>
        <v-app-bar app>
            <a href="https://nettica.com"><img class="mr-3" src="/logo.png" height="50" /></a>
            <v-toolbar-title to="/">
                {{ title }}</v-toolbar-title>

            <v-spacer />
            <v-toolbar-items>
                <v-btn to="/services" v-show="showServices">
                    Services
                    <v-icon right dark>mdi-weather-cloudy</v-icon>
                </v-btn>
                <v-btn to="/networks" right>
                    Networks
                    <v-icon right dark>mdi-network-outline</v-icon>
                </v-btn>
                <v-btn to="/hosts">
                    Hosts
                    <v-icon right dark>mdi-server</v-icon>
                </v-btn>
                <v-btn to="/accounts">
                    Account
                    <v-icon right dark>mdi-account-group</v-icon>
                </v-btn>
            </v-toolbar-items>

            <v-menu
                    left
                    bottom
            >
                <template v-slot:activator="{ on }">
                    <v-btn icon v-on="on">
                        <v-avatar size="36">
                            <img :src="user.picture"/>
                        </v-avatar>
                    </v-btn>
                </template>
                <v-card
                        class="mx-auto"
                        max-width="344"
                        outlined
                >
                    <v-list-item three-line v-show="isAuthenticated">
                        <v-list-item-content>
                            <div class="overline mb-4">connected as</div>
                            <v-list-item-title class="headline mb-1">{{user.name}}
                            <v-avatar size="64">
                                <img alt="user.name" :src="user.picture"/>
                            </v-avatar>
                            </v-list-item-title>
                            <v-list-item-subtitle>Email: {{user.email}}</v-list-item-subtitle>
                            <v-list-item-subtitle>Issuer: {{user.issuer}}</v-list-item-subtitle>
                            <v-list-item-subtitle>Issued at: {{ user.issuedAt | formatDate }}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-card-actions>
                        <v-btn small
                                v-on:click="mylogout"
                        >
                            logout
                            <v-icon small right dark>mdi-logout</v-icon>
                        </v-btn>
                    </v-card-actions>
                </v-card>
            </v-menu>

        </v-app-bar>
    </v-container>
</template>

<script>
  import {mapActions, mapGetters} from "vuex";
  import env from "../../env"

  export default {
    name: 'Header',
      data: () => ({
            showServices: env.showServicesTab,
            title: env.title,
        }),

    computed:{
      ...mapGetters({
        user: 'auth/user',
        isAuthenticated: 'auth/isAuthenticated',
      }),
    },

    methods: {
      ...mapActions('auth', {
        logout: 'logout',
      }),
      mylogout() {
        this.logout();
        window.location.href = env.logoutUrl;
      }
    },
  }
</script>
