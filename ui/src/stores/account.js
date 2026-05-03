import { defineStore } from 'pinia'
import ApiService from '@/services/api.service'

export const useAccountStore = defineStore('account', {
  state: () => ({
    error: null,
    account: null,
    accounts: [],
    users: [],
    members: [],
    limits: [],
  }),

  getters: {
    getMembers: (state) => (id) => {
      const item = state.members.find((m) => m.id === id)
      return item ? item.members : null
    },
    getLimits: (state) => (id) => {
      const item = state.limits.find((l) => l.id === id)
      return item ? item.limits : null
    },
  },

  actions: {
    async readAll(id) {
      try {
        const resp = await ApiService.get(`/accounts/${id}`)
        const incoming = new Map(resp.map(a => [a.id, a]))
        for (let i = this.accounts.length - 1; i >= 0; i--) {
          const id = this.accounts[i].id
          if (incoming.has(id)) {
            this.accounts[i] = incoming.get(id)
            incoming.delete(id)
          } else {
            this.accounts.splice(i, 1)
          }
        }
        for (const a of incoming.values()) this.accounts.push(a)
        this.accounts = [...this.accounts]
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async readUsers(id) {
      try {
        this.users = await ApiService.get(`/accounts/${id}/users`)
      } catch (error) {
        this.error = error
      }
    },

    async readMembers(id) {
      try {
        const resp = await ApiService.get(`/accounts/${id}/users`, { responseType: 'arraybuffer' })
        // console.log('readMembers: ', resp)
        const index = this.members.findIndex((x) => x.id === id)
        if (index !== -1) this.members.splice(index, 1)
        this.members.push({ id, members: resp })
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async readLimits(id) {
      try {
        const resp = await ApiService.get(`/accounts/${id}/limits`)
        // console.log('readLimits: ', resp)
        const index = this.limits.findIndex((x) => x.id === id)
        if (index !== -1) this.limits.splice(index, 1)
        this.limits.push({ id, limits: resp })
      } catch (_) {
        // expected when no limits
      }
    },

    async create(account) {
      try {
        const resp = await ApiService.post('/accounts/', account)
        this.account = resp
        this.error = `Account for ${account.email} created`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async update(account) {
      // console.log('action update account: ', account)
      try {
        const resp = await ApiService.patch(`/accounts/${account.id}`, account)
        // update in members list
        for (const memberGroup of this.members) {
          const idx = memberGroup.members.findIndex((x) => x.id === resp.id)
          if (idx !== -1) {
            memberGroup.members.splice(idx, 1)
            memberGroup.members.push(resp)
          }
        }
        // update root account if applicable
        if (resp.parent === resp.id) {
          const idx = this.accounts.findIndex((x) => x.id === resp.id)
          if (idx !== -1) {
            this.accounts.splice(idx, 1)
            this.accounts.push(resp)
          }
        }
        this.accounts = [...this.accounts]
        this.error = `${account.email} updated`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async delete(account) {
      try {
        await ApiService.delete(`/accounts/${account.id}`)
        const index = this.accounts.findIndex((x) => x.id === account.id)
        if (index !== -1) this.accounts.splice(index, 1)
        this.accounts = [...this.accounts]
        this.error = `${account.email} deleted`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },

    async email(account) {
      try {
        await ApiService.get(`/accounts/${account.id}/invite`)
        this.error = `Email to ${account.email} sent`
      } catch (error) {
        if (error.response) this.error = error.response.data.error
      }
    },
  },
})
