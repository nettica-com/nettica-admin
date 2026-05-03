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
        const resp = await ApiService.get('/subscriptions')
        const incoming = new Map(resp.map(s => [s.id, s]))
        for (let i = this.subscriptions.length - 1; i >= 0; i--) {
          const id = this.subscriptions[i].id
          if (incoming.has(id)) {
            this.subscriptions[i] = incoming.get(id)
            incoming.delete(id)
          } else {
            this.subscriptions.splice(i, 1)
          }
        }
        for (const s of incoming.values()) this.subscriptions.push(s)
        this.subscriptions = [...this.subscriptions]
      } catch (err) {
        this.error = err
      }
    },

    async create(subscription) {
      try {
        const resp = await ApiService.post('/subscriptions', subscription)
        this.subscriptions.push(resp)
        this.subscriptions = [...this.subscriptions]
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
        this.subscriptions = [...this.subscriptions]
      } catch (err) {
        this.error = err
      }
    },

    async delete(subscription) {
      try {
        await ApiService.delete(`/subscriptions/${subscription.id}`)
        const index = this.subscriptions.findIndex((x) => x.id === subscription.id)
        if (index !== -1) this.subscriptions.splice(index, 1)
        this.subscriptions = [...this.subscriptions]
      } catch (err) {
        this.error = err
      }
    },
  },
})
