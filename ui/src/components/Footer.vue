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

<script>
import { mapActions, mapGetters } from "vuex";
import mypackage from "../../package.json";

export default {
  name: 'Footer',

  data: () => ({

    version: mypackage.version,

  }),

}
</script>
