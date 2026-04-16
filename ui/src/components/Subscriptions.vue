<template>
  <v-container>
    <v-card>
      <v-card-title>Subscriptions</v-card-title>
      <v-data-table
        :headers="headers"
        :items="subscriptionStore.subscriptions"
        :items-per-page="25"
        no-data-text="No subscriptions found."
      />
    </v-card>
  </v-container>
</template>

<script setup>
import { onMounted } from 'vue'
import { useSubscriptionStore } from '@/stores/subscription'

const subscriptionStore = useSubscriptionStore()

const headers = [
  { title: 'ID', key: 'id' },
  { title: 'Name', key: 'name' },
  { title: 'Status', key: 'status' },
]

onMounted(() => {
  subscriptionStore.read()
})
</script>
