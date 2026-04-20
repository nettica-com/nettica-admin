<template>
  <v-container style="padding-top:0px">
    <v-row>
      <v-col cols="12">
        <div>
          <v-btn class="mb-3 mt-0" @click="Refresh">
            <v-icon>mdi-refresh</v-icon>
            Refresh
          </v-btn>
        </div>
        <v-card>
          <v-card-title>
            <v-row>
              <v-col cols="3">Accounts</v-col>
              <v-col cols="4">
                <v-text-field v-model="search" append-inner-icon="mdi-magnify" label="Search" hide-details></v-text-field>
              </v-col>
              <v-col cols="5" class="text-right">
                <v-btn color="#004000" @click="startInvite">
                  Invite <v-icon end>mdi-account-group</v-icon>
                </v-btn>
                &nbsp;
                <v-btn color="#000040" @click="dialogAPI = true">
                  API <v-icon end>mdi-key</v-icon>
                </v-btn>
              </v-col>
            </v-row>
          </v-card-title>
        </v-card>
        <v-card>
          <v-row>
            <v-col cols="6">
              <v-treeview
                ref="treeRef"
                v-if="showTree"
                :items="items"
                :search="search"
                :filter-fn="filter"
                item-title="name"
                item-value="id"
                v-model:activated="active"
                v-model:opened="open"
                activatable
                open-all
                hoverable
              >
                <template #prepend="{ item }">
                  <template v-if="item">
                  <span v-if="item.symbol" class="material-symbols-outlined">{{ item.symbol }}</span>
                  <v-avatar v-else-if="item.isAccount && (item.account?.accountPicture || item.account?.accountPict)" size="32">
                    <img :src="accountImageSrc(item.account)" width="32" height="32" />
                  </v-avatar>
                  <v-avatar v-else-if="item.member?.userPicture || item.member?.picture" size="32">
                    <img :src="memberImageSrc(item.member)" width="32" height="32" />
                  </v-avatar>
                  <v-icon v-else>{{ item.icon }}</v-icon>
                  </template>
                </template>
                <template #title="{ item }">
                  <table>
                    <tbody>
                    <tr><td>{{ item?.name }}</td></tr>
                    <tr v-if="item?.isMember"><td class="gray" style="font-size: small;">{{ item?.email }}</td></tr>
                    </tbody>
                  </table>
                </template>
              </v-treeview>
            </v-col>
            <v-divider vertical></v-divider>
            <v-col cols="6" class="text-center">
              <div v-if="!selected" class="text-h6 font-weight-light" style="align-self: center;"></div>
              <v-card v-else-if="selected.hasLimits" :key="selected.pid" class="px-3 mx-auto" flat>
                <v-card-text>
                  <v-avatar v-if="selected.account?.accountPicture || selected.account?.accountPict" size="50">
                    <img :src="accountImageSrc(selected.account)" width="50" height="50" />
                  </v-avatar>
                  <v-icon v-else size="50">mdi-account-group</v-icon>
                  <h3 class="text-h5 mb-2">{{ selected.account.accountName }}</h3>
                  <h5 class="text-h6 mb-2">{{ selected.account.parent }}</h5>
                </v-card-text>
                <v-divider></v-divider>
                <table width="100%" style="text-align: left;">
                  <thead><tr><th>Limit</th><th>Current</th><th>Max</th></tr></thead>
                  <tbody>
                    <tr><td>Members</td><td>{{ selected.limits.members }}</td><td>{{ selected.limits.maxmembers }}</td></tr>
                    <tr><td>Devices</td><td>{{ selected.limits.devices }}</td><td>{{ selected.limits.maxdevices }}</td></tr>
                    <tr><td>Networks</td><td>{{ selected.limits.networks }}</td><td>{{ selected.limits.maxnetworks }}</td></tr>
                    <tr><td>Services</td><td>{{ selected.limits.services }}</td><td>{{ selected.limits.maxservices }}</td></tr>
                  </tbody>
                </table>
              </v-card>
              <v-card v-else-if="selected.isMember" :key="selected.id" class="px-3 mx-auto" flat>
                <v-form autocomplete="off">
                  <v-card-text>
                    <v-avatar v-if="selected.member.userPicture || selected.member.picture" size="50">
                      <img :src="memberImageSrc(selected.member)" class="mx-auto d-block" width="50" height="50" />
                    </v-avatar>
                    <v-icon v-else size="50">mdi-account</v-icon>
                    <h3 class="text-h5 mb-2">{{ selected.name }}</h3>
                    <h5 class="text-h6 mb-2">{{ selected.role }}</h5>
                  </v-card-text>
                  <v-divider></v-divider>
                  <v-row class="px-3">
                    <v-col>
                      <v-form ref="formRef" v-model="valid">
                        <v-text-field v-model="selected.member.parent" label="Account ID" readonly />
                        <v-text-field v-model="selected.member.accountName" label="Account Name" :rules="[v => !!v || 'Account name is required']" required />
                        <v-text-field v-model="selected.member.name" label="Name" :rules="[v => !!v || 'Name is required']" required />
                        <v-text-field v-model="selected.member.email" label="Email Address" :rules="[v => !!v || 'Email address is required']" readonly>
                          <template #append-inner>
                            <v-btn icon @click="resendEmail(selected.member)">
                              <v-icon>mdi-email-outline</v-icon>
                            </v-btn>
                          </template>
                        </v-text-field>
                        <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onFileSelected" />
                        <v-text-field v-model="selected.member.picture" label="Picture"
                          append-inner-icon="mdi-upload" @click:append-inner="fileInput.click()" />
                        <v-select :items="networks" v-model="selected.netName" label="To this net" :readonly="selected.isReadOnly" />
                        <v-select :items="roles" v-model="selected.role" label="Role" :readonly="selected.isReadOnly" />
                        <v-select :items="statuses" v-model="selected.status" label="Status" />
                        <p class="text-caption">
                          Created by {{ selected.member.createdBy }} at {{ formatDate(selected.member.created) }}<br />
                          Last update by {{ selected.member.updatedBy }} at {{ formatDate(selected.member.updated) }}
                        </p>
                      </v-form>
                    </v-col>
                  </v-row>
                  <v-card-actions>
                    <v-btn color="#004000" @click="updateMember(selected)">
                      Save <v-icon end>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-spacer></v-spacer>
                    <v-btn color="#400000" @click="remove(selected.member)">
                      Delete <v-icon end>mdi-delete-outline</v-icon>
                    </v-btn>
                  </v-card-actions>
                </v-form>
              </v-card>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-if="account" v-model="dialogCreate" max-width="550" persistent @keydown.esc="dialogCreate = false">
      <v-card>
        <v-card-title class="text-h5">Invite new member</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-form ref="formRef" v-model="valid">
                <v-select
                  return-object v-model="acntList.selected" :items="acntList.items"
                  item-title="text" item-value="value" label="To this account"
                  :rules="[v => !!v || 'Account is required']" persistent-hint required
                />
                <v-select
                  return-object v-model="netList.selected" :items="netList.items"
                  item-title="text" item-value="value" label="To this net"
                  :rules="[v => !!v || 'Net is required']" persistent-hint required
                />
                <v-text-field v-model="account.name" label="Name" :rules="[v => !!v || 'Name is required']" required />
                <v-text-field v-model="toAddress" label="Enter the email address of user you'd like to invite" :rules="[v => !!v || 'Email address is required']" required />
                <v-switch v-model="sendEmail" color="success" inset label="Send Email" />
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn :disabled="!valid" color="#004000" @click="create(toAddress, netList.selected)">
            Submit <v-icon end>mdi-check-outline</v-icon>
          </v-btn>
          <v-btn color="#000040" @click="dialogCreate = false">
            Cancel <v-icon end>mdi-close-circle-outline</v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-if="accountStore.accounts" v-model="dialogAPI" max-width="650" persistent @keydown.esc="dialogAPI = false">
      <v-card>
        <v-card-title class="text-h5">API Keys</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-data-table :headers="kheaders" :items="accountStore.accounts" :items-per-page="5" class="elevation-1">
                <template #item.action="{ item }">
                  <v-row>
                    <v-icon class="pr-1 pl-1" @click="regenerateKey(item)" title="Regenerate API Key">mdi-refresh</v-icon>
                    <v-icon class="pr-1 pl-1" @click="copyKey(item.apiKey)" title="Copy API Key to clipboard">mdi-content-copy</v-icon>
                  </v-row>
                </template>
              </v-data-table>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-btn color="#000040" @click="dialogAPI = false">OK</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<style scoped>
.gray { color: gray; }
</style>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAccountStore } from '@/stores/account'
import { useNetStore } from '@/stores/net'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/formatDate'
import ApiService from '@/services/api.service'

const accountStore = useAccountStore()
const netStore = useNetStore()
const authStore = useAuthStore()
const { user: authuser } = storeToRefs(authStore)
const { accounts, members, limits } = storeToRefs(accountStore)
const { nets } = storeToRefs(netStore)

const showTree = ref(false)
const fileInput = ref(null)
const items = ref([])
const open = ref([])
const active = ref([])
const acntList = ref({})
const dialogCreate = ref(false)
const dialogAPI = ref(false)
const netList = ref({ items: [], selected: null })
const toAddress = ref('')
const sendEmail = ref(true)
const networks = ref(['All Networks'])
const roles = ['Owner', 'Admin', 'User']
const statuses = ['Active', 'Pending', 'Suspended', 'Hidden']
const account = ref(null)
const valid = ref(false)
const search = ref('')
const kheaders = [
  { title: 'Account', key: 'accountName' },
  { title: 'API Key', key: 'apiKey' },
  { title: 'Actions', key: 'action', sortable: false },
]

const selected = computed(() => {
  if (!active.value.length) return undefined
  const id = active.value[0]
  const findValue = (arr, val) => {
    for (const obj of arr) {
      if (obj.id === val) return obj
      if (obj.children) {
        const result = findValue(obj.children, val)
        if (result) return result
      }
    }
  }
  // console.log('selected item = ', findValue(items.value, id))
  return findValue(items.value, id)
})

onMounted(() => {
  netStore.readAll()
  if (authuser.value) accountStore.readAll(authuser.value.email)
})

watch(accounts, async (newAccounts) => {
  const memberFetches = newAccounts.map(a => accountStore.readMembers(a.parent))
  for (const a of newAccounts) {
    if (a.role === 'Owner' || a.role === 'Admin') {
      accountStore.readLimits(a.parent)
    }
  }
  await Promise.all(memberFetches)
  buildTree()
})

watch(members, () => buildTree(), { deep: true })
watch(limits, () => buildTree(), { deep: true })

watch(nets, (newNets) => {
  netList.value.items = [{ text: 'All Networks', value: '' }]
  networks.value = ['All Networks']
  for (let i = 0; i < newNets.length; i++) {
    netList.value.items[i + 1] = { text: newNets[i].netName, value: newNets[i].id }
    networks.value[i + 1] = newNets[i].netName
  }
})

function onFileSelected(event) {
  const file = event.target.files[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    const img = new Image()
    img.onload = () => {
      const size = Math.min(img.width, img.height)
      const sx = (img.width - size) / 2
      const sy = (img.height - size) / 2
      const canvas = document.createElement('canvas')
      canvas.width = 80
      canvas.height = 80
      canvas.getContext('2d').drawImage(img, sx, sy, size, size, 0, 0, 80, 80)
      selected.value.member.userPicture = canvas.toDataURL('image/png').split(',')[1]
      event.target.value = ''
    }
    img.src = e.target.result
  }
  reader.readAsDataURL(file)
}

function memberImageSrc(member) {
  if (member?.userPicture) return `data:image/png;base64,${member.userPicture}`
  return member?.picture || ''
}

function accountImageSrc(account) {
  if (account?.accountPict) return `data:image/png;base64,${account.accountPict}`
  return account?.accountPicture || ''
}

function Refresh() {
  if (authuser.value) accountStore.readAll(authuser.value.email)
  netStore.readAll()
}

function Refreshing() {
  Refresh()
  for (let i = 0; i < 5; i++) {
    setTimeout(() => Refresh(), (i + 1) * 2000)
  }
}

function filter(item) {
  if (item.name.toLowerCase().includes(search.value.toLowerCase())) return true
  if (item.isMember && item.netName.toLowerCase().includes(search.value.toLowerCase())) return true
  if (item.isMember && item.member.email.toLowerCase().includes(search.value.toLowerCase())) return true
  if (item.isAccount && item.account.email?.toLowerCase().includes(search.value.toLowerCase())) return true
  return false
}

function buildTree() {
  const newItems = []
  let x = 0
  for (const acnt of accounts.value) {
    const found = newItems.some((it) => it.idx === acnt.parent)
    if (!found) {
      const lims = accountStore.getLimits(acnt.parent)
      const mems = accountStore.getMembers(acnt.parent)
      const ownerRecord = mems?.find(m => m.id === m.parent) || acnt
      newItems[x++] = {
        id: 'p-' + acnt.parent,
        pid: 'p-' + acnt.parent,
        idx: acnt.parent,
        name: acnt.accountName,
        account: ownerRecord,
        status: acnt.status,
        icon: 'mdi-account-group',
        isAccount: true,
        isMember: false,
        hasLimits: lims != null,
        limits: lims,
        children: [],
      }
    }
  }

  for (const acnt of accounts.value) {
    for (const it of newItems) {
      if (acnt.parent === it.idx) {
        let name = acnt.name
        if (acnt.netName) name += ' (' + acnt.netName + ')'
        const netName = acnt.netName || 'All Networks'
        it.children.push({
          id: 'c-' + acnt.id,
          idx: acnt.id,
          name,
          email: acnt.email,
          netName,
          isReadOnly: acnt.id !== acnt.parent,
          member: acnt,
          status: acnt.status,
          role: acnt.role,
          icon: 'mdi-account',
          isAccount: false,
          isMember: true,
        })
      }
    }
  }

  for (let i = 0; i < newItems.length; i++) {
    if (!accounts.value[i]) continue
    const mems = accountStore.getMembers(accounts.value[i].parent)
    if (!mems) continue
    for (const m of mems) {
      if (m.parent !== newItems[i].idx) continue
      if (newItems[i].children.some((c) => c.idx === m.id)) continue
      let name = m.name
      if (m.netName) name += ' (' + m.netName + ')'
      const netName = m.netName || 'All Networks'
      newItems[i].children.push({
        id: 'm-' + m.id + '-' + m.netName,
        idx: m.id,
        name,
        member: m,
        status: m.status,
        netName,
        isReadOnly: false,
        role: m.role,
        email: m.email,
        icon: 'mdi-account',
        isAccount: false,
        isMember: true,
      })
    }
  }

  const sortByName = (a, b) => a.name.toUpperCase().localeCompare(b.name.toUpperCase())
  newItems.sort(sortByName)
  newItems.forEach((it) => it.children.sort(sortByName))

  items.value = newItems
  showTree.value = true
}

function startInvite() {
  account.value = { name: '', from: authuser.value?.email, email: '' }
  netList.value = { selected: { text: '', value: '' }, items: [{ text: 'All Networks', value: '' }] }
  for (let i = 0; i < nets.value.length; i++) {
    netList.value.items[i + 1] = { text: nets.value[i].netName, value: nets.value[i].id }
  }
  netList.value.selected = netList.value.items[0]
  acntList.value = { selected: { text: '', value: '' }, items: [] }
  let x = 0, sel = 0
  for (const a of accounts.value) {
    if (a.role === 'Owner' || a.role === 'Admin') {
      acntList.value.items[x] = { text: a.accountName + ' - ' + a.parent, value: a.parent }
      if (acntList.value.items[x].value === a.id) sel = x
      x++
    }
  }
  if (x === 0) {
    accountStore.error = 'You must be an admin or owner to invite a member'
    return
  }
  acntList.value.selected = acntList.value.items[sel]
  dialogCreate.value = true
}

async function create(toAddr, net) {
  account.value.email = toAddr
  account.value.netId = net.value
  account.value.netName = net.text === 'All Networks' ? '' : net.text
  if (!account.value.netName) account.value.netId = ''
  account.value.from = authuser.value?.email
  account.value.role = 'User'
  account.value.status = sendEmail.value ? 'Pending' : 'Active'

  for (const a of accounts.value) {
    if (acntList.value.selected.value === a.parent) {
      account.value.parent = a.parent
      account.value.accountName = a.accountName
      break
    }
  }

  try {
    const a = await ApiService.post('/accounts', account.value)
    if (sendEmail.value) accountStore.email(a)
  } catch (e) {
    accountStore.error = e
  }

  dialogCreate.value = false
  Refreshing()
}

function resendEmail(acnt) {
  accountStore.email(acnt)
}

function remove(item) {
  if (item.role === 'Owner') {
    alert('You cannot delete owners')
  } else if (confirm(`Do you really want to delete ${item.name} ?`)) {
    accountStore.delete(item)
  }
  if (authuser.value) accountStore.readAll(authuser.value.email)
}

function updateMember(item) {
  item.member.netName = item.netName
  if (item.member.netName === 'All Networks') {
    item.member.netName = ''
    item.member.netId = ''
  }
  item.member.role = item.role
  item.member.status = item.status

  for (const n of nets.value) {
    if (n.netName === item.member.netName) {
      item.member.netId = n.id
      break
    }
  }

  const name = item.member.netName ? item.member.name + ' (' + item.member.netName + ')' : item.member.name
  item.name = name
  accountStore.update(item.member)
}

function regenerateKey(item) {
  if (confirm(`Do you really want to regenerate the API key for ${item.accountName} ?`)) {
    item.apiKey = ''
    accountStore.update(item)
    Refresh()
    if (authuser.value) accountStore.readAll(authuser.value.email)
    accountStore.error = 'API key regenerated'
  }
}

function copyKey(text) {
  navigator.clipboard.writeText(text).then(() => {
    accountStore.error = 'API key copied to clipboard'
  })
}
</script>
