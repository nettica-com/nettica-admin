<template>
  <v-navigation-drawer app permanent>
    <v-list nav density="compact">
      <v-list-item
        v-show="showServices"
        prepend-icon="mdi-weather-cloudy"
        title="Services"
        value="services"
        to="/services"
      />
      <v-list-item prepend-icon="mdi-network-outline" title="Networks" value="networks" to="/networks" />
      <v-list-item prepend-icon="mdi-devices" title="Devices" value="devices" to="/devices" />
      <v-list-item prepend-icon="mdi-account-group" title="Accounts" value="accounts" to="/accounts" />
    </v-list>
  </v-navigation-drawer>
</template>

<script setup>
import env from '../../env.json'
const showServices = env.showServicesTab
</script>
