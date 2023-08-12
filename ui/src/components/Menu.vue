<template>
        <v-navigation-drawer app clipped permanent>
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
</template>

<script>
  import env from "../../env"

  export default {
    name: 'Menu',
      data: () => ({
            showServices: env.showServicesTab,
        }),
    }

</script>