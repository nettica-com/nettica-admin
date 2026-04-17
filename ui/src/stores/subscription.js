import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useSubscriptionStore = defineStore('subscription', {
  state: () => ({
    error: null,
    subscriptions: [],
  }),

  actions: {
    async read() {
      try {
        this.subscriptions = await ApiService.get('/subscriptions')
      } catch (err) {
        this.error = err
      }
    },

    async create(subscription) {
      try {
        const resp = await ApiService.post('/subscriptions', subscription)
        this.subscriptions.push(resp)
      } catch (err) {
        this.error = err
      }
    },

    async update(subscription) {
      try {
        const resp = await ApiService.patch(`/subscriptions/${subscription.id}`, subscription)
        const index = this.subscriptions.findIndex((x) => x.id === resp.id)
        if (index !== -1) {
          this.subscriptions.splice(index, 1)
          this.subscriptions.push(resp)
        } else {
          this.error = 'update subscription failed, not in list'
        }
      } catch (err) {
        this.error = err
      }
    },

    async delete(subscription) {
      try {
        await ApiService.delete(`/subscriptions/${subscription.id}`)
        const index = this.subscriptions.findIndex((x) => x.id === subscription.id)
        if (index !== -1) this.subscriptions.splice(index, 1)
      } catch (err) {
        this.error = err
      }
    },
  },
})
