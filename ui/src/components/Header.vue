<template>
    <div>
        <v-app-bar app clipped-left class="hidden-xs-only">
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
                <v-btn to="/devices">
                    Devices
                    <v-icon right dark>mdi-devices</v-icon>
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
        <v-app-bar app clipped-left class="hidden-sm-and-up">
            <a href="https://nettica.com"><img class="mr-3" src="/logo.png" height="50" /></a>
            <v-toolbar-title to="/">
                {{ title }}</v-toolbar-title>

            <v-spacer />
            <v-btn icon @click="myShowMenu()">
                  <v-icon>mdi-menu</v-icon>
            </v-btn>
        </v-app-bar>
        <v-navigation-drawer app clipped right v-model="showMenu" class="hidden-sm-and-up">
                <v-list nav dense>
                <v-list-item prepend-icon="mdi-weather-cloudy" title="Services" value="services" to="/services" v-show="showServices">
                    <v-list-item-icon><v-icon>mdi-weather-cloudy</v-icon></v-list-item-icon>
                    <v-list-item-title>Services</v-list-item-title>
                </v-list-item>
                <v-list-item prepend-icon="mdi-network-outline" title="Networks" value="networks" to="/networks">
                    <v-list-item-icon><v-icon>mdi-network-outline</v-icon></v-list-item-icon>
                    <v-list-item-title>Networks</v-list-item-title>    
                </v-list-item>
                <v-list-item prepend-icon="mdi-devices" title="Devices" value="devices" to="/devices">
                    <v-list-item-icon><v-icon>mdi-devices</v-icon></v-list-item-icon>
                    <v-list-item-title>Devices</v-list-item-title>
                </v-list-item>
                <v-list-item prepend-icon="mdi-account-group" title="Accounts" value="/accounts" to="/accounts">
                    <v-list-item-icon><v-icon>mdi-account-group</v-icon></v-list-item-icon>
                    <v-list-item-title>Accounts</v-list-item-title>                    
                </v-list-item>
                </v-list>
        </v-navigation-drawer>

    </div>

</template>

<script>
  import {mapActions, mapGetters} from "vuex";
  import env from "../../env"

  export default {
    name: 'Header',
      data: () => ({
            title: env.title,
            showMenu: false,
            showServices: env.showServicesTab,
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
      },
      myShowMenu() {
        this.showMenu = !this.showMenu;
        console.log("showMenu = " + this.showMenu);
      },

    },
  }
</script>

<!--
<div id="app">
    <v-app id="inspire">
      <v-layout row justify-center>
        <v-toolbar app dark color="blue-grey darken-1" class="hidden-xs-and-down">
          <v-toolbar-title>Desktop Menu</v-toolbar-title>
          <v-spacer></v-spacer>
          <v-toolbar-items>
           <v-btn
             v-for="item in nav"
             :key="item.icon"
             to="#"
             :title="item.title"
             flat
           >{{ item.text }}</v-btn>
          </v-toolbar-items>
        </v-toolbar>
        
        <v-toolbar app dark color="blue-grey darken-3" class="hidden-sm-and-up">
          <v-toolbar-title>Mobile Menu</v-toolbar-title>
          <v-spacer></v-spacer>
  
          <v-dialog v-model="dialog" fullscreen hide-overlay transition="dialog-bottom-transition">
            <v-toolbar-side-icon dark slot="activator"></v-toolbar-side-icon>
            <v-card>
              <v-toolbar flat color="blue-grey darken-2">
                <v-toolbar-title>Mobile Menu</v-toolbar-title>
                <v-spacer></v-spacer>
                <v-btn icon @click.native="dialog = false">
                  <v-icon>close</v-icon>
                </v-btn>
              </v-toolbar>
  
              <v-list>
                <v-list-tile
                  v-for="(item, index) in nav"
                  :key="index"
                  to="#"
                >
                  <v-list-tile-action>
                    <v-icon v-if="item.icon">{{item.icon}}</v-icon>
                  </v-list-tile-action>
                  <v-list-tile-content>
                    <v-list-tile-title :title="item.title">{{ item.text }}</v-list-tile-title>
                  </v-list-tile-content>
                </v-list-tile>
              </v-list>
            </v-card>
          </v-dialog>
  
        </v-toolbar>
        
        <v-container fluid>
          <v-slide-y-transition mode="out-in">
            <v-layout column align-center>
              <h1 class="display-1">Vuetify Desktop / Mobile navbar example</h1>
              <p>
                A quick demo of how to combine a desktop navigation and a 
                mobile overlay (dialog) navigation menu.
              </p>
              <p>
                Resize the window to see the navbar change to mobile version.
              </p>
              <p>
                My deep gratitude towards the VueJS and Vuetify team!
              </p>
            </v-layout>
          </v-slide-y-transition>
        </v-container>
      </v-layout>
    </v-app>
  </div>

-->