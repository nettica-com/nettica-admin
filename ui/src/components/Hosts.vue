<template>
  <v-container style="padding-top:0px">
    <div>
      <v-btn class="mb-3 mt-0" @click="Refresh">
        <v-icon>mdi-refresh</v-icon>
        Refresh
      </v-btn>
    </div>
    <v-card>
      <v-card-title>
        Hosts
        <v-spacer></v-spacer>
        <v-text-field
          v-if="listView"
          v-model="search"
          append-inner-icon="mdi-magnify"
          label="Search"
          hide-details
        ></v-text-field>
        <v-spacer></v-spacer>
        <v-btn color="success" @click="startCreate">
          Add Device Manually
          <v-icon end>mdi-devices</v-icon>
        </v-btn>
      </v-card-title>
      <v-data-table
        v-if="listView"
        :headers="headers"
        :items="hostStore.hosts"
        :search="search"
        :items-per-page="25"
        no-data-text="No Hosts. Click above to create your first host, or use Nettica Agent on the host."
        no-results-text="No results matching your search"
      >
        <template #item.name="{ item }">{{ item.name }}</template>
        <template #item.status="{ item }">
          <v-icon v-if="item.status === 'Online'" color="green">mdi-check-circle</v-icon>
          <v-icon v-else-if="item.status === 'Native'" color="blue">mdi-minus-circle</v-icon>
          <v-icon v-else color="red">mdi-close-circle</v-icon>
          {{ item.status }}
        </template>
        <template #item.address="{ item }">
          <v-chip v-for="(ip, i) in item.address" :key="i" color="#336699" text-color="white">
            <v-icon start>mdi-ip-network</v-icon>
            {{ ip }}
          </v-chip>
        </template>
        <template #item.tags="{ item }">
          <v-chip v-for="(tag, i) in item.tags" :key="i" color="blue-grey">
            <v-icon start>mdi-tag</v-icon>
            {{ tag }}
          </v-chip>
        </template>
        <template #item.created="{ item }">
          <v-row><p>{{ item.createdBy }} at {{ formatDate(item.created) }}</p></v-row>
        </template>
        <template #item.updated="{ item }">
          <v-row><p>At {{ formatDate(item.updated) }} by {{ item.updatedBy }}</p></v-row>
        </template>
        <template #item.action="{ item }">
          <v-row v-if="item.type === 'ServiceHost'">
            <v-icon class="pr-1 pl-1" @click.stop="startUpdateServiceHost(item)" title="Edit Service">
              mdi-square-edit-outline
            </v-icon>
          </v-row>
          <v-row v-else>
            <v-icon class="pr-1 pl-1" @click.stop="startUpdate(item)" title="Edit">mdi-square-edit-outline</v-icon>
            <v-icon class="pr-1 pl-1" @click.stop="startCopy(item)" title="Copy">mdi-content-copy</v-icon>
            <v-icon class="pr-1 pl-1" @click="remove(item)" title="Delete">mdi-trash-can-outline</v-icon>
            <v-switch class="pr-1 pl-1" color="success" v-model="item.enable" @update:model-value="updateEnable(item)" />
          </v-row>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-if="host" v-model="dialogCreate" max-width="550">
      <v-card>
        <v-card-title class="text-h5">Add New Host</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-form ref="formRef" v-model="valid">
                <v-text-field v-model="host.name" label="Host friendly name" :rules="[v => !!v || 'host name is required']" required />
                <v-select
                  return-object
                  v-model="netList.selected"
                  :items="netList.items"
                  item-title="text"
                  item-value="value"
                  label="Join this net"
                  :rules="[v => !!v || 'Net is required']"
                  persistent-hint
                  required
                />
                <v-text-field v-model="host.current.endpoint" label="Public endpoint for clients" />
                <v-text-field v-model="host.current.listenPort" type="number" label="Listen port" />
                <v-combobox v-model="host.tags" chips closable-chips hint="Enter a tag, hit tab, hit enter." label="Tags" multiple />
                <v-switch v-model="host.enable" color="success" inset :label="host.enable ? 'Enable host after creation' : 'Disable host after creation'" />
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn :disabled="!valid" color="success" @click="create(host)">
            Submit <v-icon end>mdi-check-outline</v-icon>
          </v-btn>
          <v-btn color="primary" @click="dialogCreate = false">
            Cancel <v-icon end>mdi-close-circle-outline</v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-if="host" v-model="dialogUpdate" max-width="550">
      <v-card>
        <v-card-title class="text-h5">Edit Host</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-form ref="formRef" v-model="valid">
                <v-text-field v-model="host.name" label="Friendly name" :rules="[v => !!v || 'host name is required']" required />
                <v-select
                  return-object
                  v-model="netList.selected"
                  :items="netList.items"
                  item-title="text"
                  item-value="value"
                  label="Join this net"
                  :rules="[v => !!v || 'Net is required']"
                  persistent-hint
                  required
                />
                <v-text-field v-model="host.current.endpoint" label="Public endpoint for clients" />
                <v-text-field v-model="host.current.listenPort" type="number" label="Listen port" />
                <v-combobox v-model="host.tags" chips closable-chips hint="Write tag name and hit enter" label="Tags" multiple />
                <v-btn color="success" @click="forceFileDownload(host)">
                  Download Config <v-icon end>mdi-cloud-download-outline</v-icon>
                </v-btn>
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
      <v-expansion-panels>
        <v-expansion-panel>
          <v-expansion-panel-title>Advanced Configuration</v-expansion-panel-title>
          <v-expansion-panel-text>
            <div class="d-flex flex-no-wrap justify-space-between">
              <v-col cols="12">
                <v-select
                  return-object
                  v-model="platforms.selected"
                  :items="platforms.items"
                  item-title="text"
                  item-value="value"
                  label="Platform of this host"
                  persistent-hint
                />
                <v-combobox v-model="host.current.address" chips closable-chips hint="Write IPv4 or IPv6 CIDR and hit enter" label="Addresses" multiple />
                <v-combobox v-model="host.current.allowedIPs" chips closable-chips hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple />
                <v-switch v-model="publicSubnets" color="success" inset label="Route all public traffic through tunnel" />
                <v-combobox v-model="host.current.dns" chips closable-chips hint="Enter IP address(es) and hit enter or leave empty." label="DNS servers for this host" multiple />
                <v-text-field v-model="host.current.table" label="Table" />
                <v-text-field v-model="host.id" label="Host ID" disabled />
                <v-text-field v-model="host.current.publicKey" label="Public key" />
                <v-text-field
                  v-model="host.current.privateKey"
                  label="Private key"
                  autocomplete="off"
                  :append-inner-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                  :type="showPrivate ? 'text' : 'password'"
                  hint="Clear this field to have the client manage its private key"
                  @click:append-inner="showPrivate = !showPrivate"
                />
                <v-text-field
                  v-model="host.current.presharedKey"
                  label="Preshared Key"
                  autocomplete="off"
                  :append-inner-icon="showPreshared ? 'mdi-eye' : 'mdi-eye-off'"
                  :type="showPreshared ? 'text' : 'password'"
                  @click:append-inner="showPreshared = !showPreshared"
                />
                <v-text-field v-model="host.hostGroup" label="Host Group" />
                <v-text-field v-model="host.apiKey" label="API Key" />
                <v-text-field v-model="host.current.mtu" type="number" label="Define global MTU" hint="Leave at 0 and let us take care of MTU" />
                <v-text-field v-model="host.current.persistentKeepalive" type="number" label="Persistent keepalive" hint="To disable, set to 0. Recommended value 29 (seconds)" />
                <v-textarea v-model="host.current.postUp" label="PostUp Script" hint="Only applies to linux servers" />
                <v-textarea v-model="host.current.postDown" label="PostDown Script" hint="Only applies to linux servers" />
                <v-switch v-model="host.current.subnetRouting" color="success" inset label="Enable subnet routing" />
                <v-switch v-model="host.current.upnp" color="success" inset label="Enable UPnP" />
                <v-switch v-model="host.current.enableDns" color="success" inset label="Enable Nettica DNS" />
              </v-col>
            </div>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>
      <v-card>
        <v-card-actions>
          <v-btn :disabled="!valid" color="success" @click="update(host)">
            Submit <v-icon end>mdi-check-outline</v-icon>
          </v-btn>
          <v-btn color="primary" @click="dialogUpdate = false">
            Cancel <v-icon end>mdi-close-circle-outline</v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-if="host" v-model="dialogCopy" max-width="550">
      <v-card>
        <v-card-title class="text-h5">Copy Host to Net</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-form ref="formRef" v-model="valid">
                <v-text-field v-model="host.name" label="New name for host" :rules="[v => !!v || 'host name is required']" required />
                <v-select
                  return-object
                  v-model="netList.selected"
                  :items="netList.items"
                  item-title="text"
                  item-value="value"
                  label="Copy to this net"
                  :rules="[v => !!v || 'Net is required']"
                  persistent-hint
                  required
                />
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
      <v-card>
        <v-card-actions>
          <v-btn :disabled="!valid" color="success" @click="copy(host)">
            Submit <v-icon end>mdi-check-outline</v-icon>
          </v-btn>
          <v-btn color="primary" @click="dialogCopy = false">
            Cancel <v-icon end>mdi-close-circle-outline</v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-if="host" v-model="dialogServiceHost" max-width="550">
      <v-card>
        <v-card-title class="text-h5">Manage Service: {{ host.name }}</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="12">
              <v-form ref="formRef" v-model="valid">
                <v-combobox v-model="host.current.allowedIPs" chips closable-chips hint="Enter IPv4 or IPv6 CIDR and press tab" label="Allowed IPs" multiple />
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
      </v-card>
      <v-card>
        <v-card-actions>
          <v-btn :disabled="!valid" color="success" @click="update(host)">
            Submit <v-icon end>mdi-check-outline</v-icon>
          </v-btn>
          <v-btn color="primary" @click="dialogServiceHost = false">
            Cancel <v-icon end>mdi-close-circle-outline</v-icon>
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useHostStore } from '@/stores/host'
import { useNetStore } from '@/stores/net'
import { useAccountStore } from '@/stores/account'
import { useAuthStore } from '@/stores/auth'
import { isCidr } from '@/plugins/cidr'
import { formatDate } from '@/utils/formatDate'

const hostStore = useHostStore()
const netStore = useNetStore()
const accountStore = useAccountStore()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)

const showPrivate = ref(true)
const showPreshared = ref(true)
const listView = ref(true)
const dialogCreate = ref(false)
const dialogUpdate = ref(false)
const dialogCopy = ref(false)
const dialogServiceHost = ref(false)
const host = ref(null)
const valid = ref(false)
const netList = ref({})
const publicSubnets = ref(false)
const search = ref('')
const platforms = ref({
  selected: { text: '', value: '' },
  items: [
    { text: 'Windows', value: 'Windows' },
    { text: 'Linux', value: 'Linux' },
    { text: 'MacOS', value: 'MacOS' },
    { text: 'Apple iOS', value: 'iOS' },
    { text: 'Android', value: 'Android' },
    { text: 'Native WireGuard', value: 'Native' },
  ],
})
const headers = [
  { title: 'Name', key: 'name' },
  { title: 'Status', key: 'status' },
  { title: 'Net', key: 'netName' },
  { title: 'IP addresses', key: 'current.address' },
  { title: 'Endpoint', key: 'current.endpoint' },
  { title: 'Tags', key: 'tags' },
  { title: 'Actions', key: 'action', sortable: false },
]

onMounted(() => {
  if (user.value) accountStore.readAll(user.value.email)
  hostStore.readAll()
  netStore.readAll()
})

function Refresh() {
  if (user.value) accountStore.readAll(user.value.email)
  hostStore.readAll()
  netStore.readAll()
}

function buildNetList(currentNetName) {
  const list = { selected: { text: '', value: '' }, items: [] }
  let selected = 0
  for (let i = 0; i < netStore.nets.length; i++) {
    list.items[i] = { text: netStore.nets[i].netName, value: netStore.nets[i].id }
    if (list.items[i].text === currentNetName) selected = i
  }
  list.selected = list.items[selected] || list.selected
  return list
}

function startCreate() {
  host.value = { name: '', email: user.value?.email, enable: true, tags: [], current: {} }
  netList.value = buildNetList('')
  dialogCreate.value = true
}

function create(h) {
  h.current.listenPort = parseInt(h.current.listenPort, 10)
  if (h.current.endpoint && !h.current.endpoint.includes(':')) {
    if (h.current.listenPort === 0) h.current.listenPort = 51820
    h.current.endpoint = h.current.endpoint + ':' + h.current.listenPort
  }
  h.netName = netList.value.selected.text
  h.netid = netList.value.selected.value
  h.platform = platforms.value.selected.value
  dialogCreate.value = false
  hostStore.create(h)
}

function remove(h) {
  if (confirm(`Do you really want to delete ${h.name} ?`)) {
    hostStore.delete(h)
  }
}

function startUpdate(h) {
  host.value = h
  hostStore.readConfig(h)
  netList.value = buildNetList(h.netName)
  for (const p of platforms.value.items) {
    if (p.value === h.platform) { platforms.value.selected = p; break }
  }
  publicSubnets.value = false
  dialogUpdate.value = true
}

function startUpdateServiceHost(h) {
  host.value = h
  hostStore.readConfig(h)
  dialogServiceHost.value = true
}

function startCopy(h) {
  host.value = h
  hostStore.readConfig(h)
  netList.value = buildNetList(h.netName)
  dialogCopy.value = true
  dialogUpdate.value = false
}

function copy(h) {
  h.current.listenPort = parseInt(h.current.listenPort, 10)
  h.current.persistentKeepalive = parseInt(h.current.persistentKeepalive, 10)
  h.current.mtu = parseInt(h.current.mtu, 10)
  if (h.netid !== netList.value.selected.value) {
    h.netName = netList.value.selected.text
    h.netid = netList.value.selected.value
    h.id = ''
    h.current.endpoint = ''
    h.current.listenPort = 0
    hostStore.create(h)
  }
  hostStore.readAll()
  dialogCopy.value = false
}

function updateEnable(h) {
  hostStore.update(h)
}

function update(h) {
  h.current.listenPort = parseInt(h.current.listenPort, 10)
  if (h.current.endpoint && !h.current.endpoint.includes(':')) {
    if (h.current.listenPort === 0) h.current.listenPort = 51820
    h.current.endpoint = h.current.endpoint + ':' + h.current.listenPort
  }
  h.current.persistentKeepalive = parseInt(h.current.persistentKeepalive, 10)
  h.current.mtu = parseInt(h.current.mtu, 10)

  if (netList.value.selected && h.netid !== netList.value.selected.value) {
    h.netName = netList.value.selected.text
    h.netid = netList.value.selected.value
  }
  if (netList.value.selected) h.netName = netList.value.selected.text
  if (platforms.value.selected) h.platform = platforms.value.selected.value

  if (publicSubnets.value) {
    h.current.allowedIPs.push(
      '0.0.0.0/5','8.0.0.0/7','11.0.0.0/8','12.0.0.0/6','16.0.0.0/4','32.0.0.0/3','64.0.0.0/3',
      '96.0.0.0/4','112.0.0.0/5','120.0.0.0/6','124.0.0.0/7','126.0.0.0/8','128.0.0.0/3',
      '160.0.0.0/5','168.0.0.0/6','172.0.0.0/12','172.32.0.0/11','172.64.0.0/10','172.128.0.0/9',
      '173.0.0.0/8','174.0.0.0/7','176.0.0.0/4','192.0.0.0/9','192.128.0.0/11','192.160.0.0/13',
      '192.169.0.0/16','192.170.0.0/15','192.172.0.0/14','192.176.0.0/12','192.192.0.0/10',
      '193.0.0.0/8','194.0.0.0/7','196.0.0.0/6','200.0.0.0/5','208.0.0.0/4',
      '::/1','8000::/2','c000::/3','e000::/4','f000::/5','f800::/6','fe00::/9','fe80::/10','ff00::/8',
    )
  }

  if (h.current.allowedIPs.length < 1) {
    hostStore.error = 'Please provide at least one valid CIDR address for host allowed IPs'; return
  }
  for (const ip of h.current.allowedIPs) {
    if (isCidr(ip) === 0) { hostStore.error = 'Invalid CIDR detected, please correct before submitting'; return }
  }
  if (h.current.address.length < 1) {
    hostStore.error = 'Please provide at least one valid CIDR address for host'; return
  }
  for (const addr of h.current.address) {
    if (isCidr(addr) === 0) { hostStore.error = 'Invalid CIDR detected, please correct before submitting'; return }
  }

  dialogUpdate.value = false
  dialogServiceHost.value = false
  hostStore.update(h)
}

function forceFileDownload(h) {
  const config = hostStore.gethostConfig(h.id)
  if (!config) { hostStore.error = 'Failed to download host config'; return }
  const url = window.URL.createObjectURL(new Blob([config]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', h.name.split(' ').join('-') + '-' + h.netName.split(' ').join('-') + '.zip')
  document.body.appendChild(link)
  link.click()
}
</script>
