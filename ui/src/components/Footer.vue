<template>
  <v-footer app>
    <v-row justify="start" no-gutters>
      <v-col>
        <small>Copyright &copy; {{ new Date().getFullYear() }} nettica.com</small>
        <v-spacer></v-spacer>
      </v-col>
      <v-col align="right">
        <small>Version: {{ version }}</small>
      </v-col>
    </v-row>
  </v-footer>
</template>

<script setup>
import mypackage from '../../package.json'
const version = mypackage.version
</script>
