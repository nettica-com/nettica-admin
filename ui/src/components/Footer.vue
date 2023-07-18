<template>
    <v-container>
        <v-footer app>
            <v-row justify="start" no-gutters>
                <v-col cols="12" lg="6" md="12" sm="12">
                    <div :align="$vuetify.breakpoint.smAndDown ? 'center' : 'left'">
                        <small>Copyright &copy; {{ new Date().getFullYear() }}, meshify.app</small>
                    </div>
                </v-col>
            </v-row>
            <v-row justify="end" no-gutters>
                <v-col cols="12" lg="6" md="12" sm="12">
                </v-col>
            </v-row>
        </v-footer>
    </v-container>
</template>

<script>
  import {mapActions, mapGetters} from "vuex";

  export default {
    name: 'Footer',

    data: () => ({

    }),

    computed:{
      ...mapGetters({
        version: 'server/version',
      }),
    },

    mounted() {
      this.versionServer()
    },

    methods: {
      ...mapActions('server', {
        versionServer: 'version',
      }),
    }
  }
</script>
