import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useJoinStore = defineStore('join', {
  state: () => ({
    error: null,
    account: [],
  }),

  actions: {
    async activate(id) {
      try {
        this.account = await ApiService.post('/accounts/' + id + '/activate')
      } catch (err) {
        this.error = err
      }
    },
  },
})
