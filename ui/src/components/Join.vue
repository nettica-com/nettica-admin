<template>
    <v-container>
        <v-card>
            <v-card-title>
                Join
            </v-card-title>
        </v-card>
        <v-card>
            <p>Joining {{ $route.query.id }}</p>
            <p>Click <a href="/hosts">here</a> to manage your meshes</p>
        </v-card>
    </v-container>
</template>
<script>
  import { mapActions, mapGetters } from 'vuex'

  export default {
    name: 'Join',

    data: () => ({
      result : "",
      panel: 1,
      id: "",
      valid: false,
      search: '',
    }),

    computed:{
      ...mapGetters({
        accounts: 'account/accounts',

      }),
    },

    mounted () {
      this.id = this.$route.query.id
      this.result = this.activate(this.$route.query.id)
    },

    methods: {
      ...mapActions('join', {
        activate: 'activate',
      }),

    }
  };
</script>
