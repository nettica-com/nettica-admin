<template>
    <v-container style="padding-top:0px">
        <div>
            <v-btn class="mb-3 mt-0" @click="Refresh()">
                <v-icon>mdi-refresh</v-icon>
                Refresh
            </v-btn>
        </div>
        <v-card>
            <v-card-title>
                <v-row>
                    <v-col cols="2">
                        Services
                    </v-col>
                    <v-col cols="4">
                        <v-text-field v-if="listView" v-model="search" append-inner-icon="mdi-magnify" label="Search"
                            hide-details></v-text-field>
                    </v-col>
                    <v-col cols="6" class="text-right">
                        <v-btn color="#004000" @click="startCreateService">
                            Add Service
                            <v-icon end>mdi-weather-cloudy</v-icon>
                        </v-btn>&nbsp;
                        <v-btn color="#004000" @click="startCreateMultihop">
                            Add Multihop
                            <v-icon end>mdi-weather-cloudy</v-icon>
                        </v-btn>&nbsp;
                        <v-btn color="#400000" @click="startCreateWilderness">
                            In the Wild
                            <v-icon end>mdi-weather-cloudy</v-icon>
                        </v-btn>
                    </v-col>
                </v-row>
            </v-card-title>
        </v-card>
        <v-card>
            <v-card-title>
                Subscriptions
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table v-if="listView" :headers="headers" :items="subscriptionStore.subscriptions" :search="search">
                <template #no-data>
                    Welcome to Nettica! Order a subscription on the <a href="https://nettica.com">Nettica website</a> to get
                    started.
                </template>
                <template v-slot:item.issued="{ item }">
                    {{ formatDate(item.issued) }}
                </template>
                <template v-slot:item.expires="{ item }">
                    {{ formatDate(item.expires) }}
                </template>
                <template v-slot:item.lastUpdated="{ item }">
                    <v-row>
                        <p>At {{ formatDate(item.lastUpdated) }} by {{ item.updatedBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.action="{ item }">
                    <v-row>
                        <v-icon class="pr-1 pl-1" @click="removeSubscription(item)"
                            title="Remove Subscription (not recommended)">
                            mdi-trash-can-outline
                        </v-icon>
                    </v-row>
                </template>
            </v-data-table>
        </v-card>
        <v-card>
            <v-card-title>
                Services
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table v-if="listView" :headers="bottom_headers" :items="serviceStore.services" :search="search"
                @click:row="startUpdateService">
                <template #no-data>
                    Creating a service host requires a subscription. Order a subscription on the <a
                        href="https://nettica.com">Nettica website</a> to get started.
                </template>
                <template v-slot:item.created="{ item }">
                    {{ formatDate(item.created) }}
                </template>
                <template v-slot:item.updated="{ item }">
                    <v-row>
                        <p>At {{ formatDate(item.updated) }} by {{ item.updatedBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.action="{ item }">
                    <v-row>
                        <v-icon class="pr-1 pl-1" @click="remove(item)" title="Delete Service">
                            mdi-trash-can-outline
                        </v-icon>
                    </v-row>
                </template>
            </v-data-table>
        </v-card>
        <v-dialog v-if="subscriptionStore.subscriptions" v-model="dialogCreateService" max-width="550" persistent @keydown.esc="dialogCreateService = false">
            <v-card>
                <v-card-title class="headline">Create New Service</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-select return-object v-model="svcList.selected" :items="svcList.items" item-title="text"
                                    item-value="value" label="Choose type of Service"
                                    :rules="[v => !!v || 'Service is required']" persistent-hint required />
                                <v-select return-object v-model="serverList.selected" :items="serverList.items"
                                    item-title="text" item-value="value" label="Pick a region"
                                    :rules="[v => !!v || 'Server is required']" persistent-hint required />
                                <v-select return-object v-model="netList.selected" :items="netList.items" item-title="text"
                                    item-value="value" label="To this net" :rules="[v => !!v || 'Net is required']"
                                    persistent-hint required />
                                <v-select return-object v-model="dnsList.selected" :items="dnsList.items" item-title="text"
                                    item-value="value" label="Select a DNS provider"
                                    :rules="[v => !!v || 'DNS is required']" persistent-hint required />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="#004000" @click="create()">
                        Submit
                        <v-icon end>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="#000040" @click="dialogCreateService = false">
                        Cancel
                        <v-icon end>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog v-if="subscriptionStore.subscriptions" v-model="dialogCreateMultihop" max-width="550" persistent @keydown.esc="dialogCreateMultihop = false">
            <v-card>
                <v-card-title class="headline">Create New Multihop Service</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-select return-object v-model="ingressList.selected" :items="ingressList.items"
                                    item-title="text" item-value="value" label="Pick a region for ingress"
                                    :rules="[v => !!v || 'Ingress region is required']" persistent-hint required />
                                <v-select return-object v-model="egressList.selected" :items="egressList.items"
                                    item-title="text" item-value="value" label="Pick a region for egress"
                                    :rules="[v => !!v || 'Egress region is required']" persistent-hint required />
                                <v-select return-object v-model="netList.selected" :items="netList.items" item-title="text"
                                    item-value="value" label="To this net" :rules="[v => !!v || 'Net is required']"
                                    persistent-hint required />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="#004000" @click="create_multihop()">
                        Submit
                        <v-icon end>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="#000040" @click="dialogCreateMultihop = false">
                        Cancel
                        <v-icon end>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog v-if="subscriptionStore.subscriptions" v-model="dialogWilderness" max-width="550" persistent @keydown.esc="dialogWilderness = false">
            <v-card>
                <v-card-title class="headline">In the Wild</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <p v-if="!wild">This will create a new service host from Nettica on your Nettica VPN Server.  We will not
                                store your ApiKey, but will store the device and VPN (and their API keys) for the service host
                                embedded in your network.  Click Login to log in and accept the consent, then navigate back to
                                services to create your service.</p>
                                <v-text-field v-model="wildServer" label="Server Name" :readonly="wild" required></v-text-field>
                                <v-btn v-if="!wild" color="#400000" @click="loginWild">
                                    Login
                                    <v-icon end>mdi-lock</v-icon>
                                </v-btn>&nbsp;&nbsp;
                                <v-btn v-if="!wild" color="#000040" @click="dialogWilderness = false">
                                    Cancel
                                    <v-icon end>mdi-close-circle-outline</v-icon>
                                </v-btn>
                                <div v-if="wild">
                                    <v-select return-object v-model="svcList.selected" :items="svcList.items" item-title="text"
                                        item-value="value" label="Choose type of Service"
                                        :rules="[v => !!v || 'Service is required']" persistent-hint required />
                                    <v-select return-object v-model="serverList.selected" :items="serverList.items"
                                        item-title="text" item-value="value" label="Pick a region"
                                        :rules="[v => !!v || 'Server is required']" persistent-hint required />
                                    <v-select return-object v-model="wildList.selected" :items="wildList.items" item-title="text"
                                        item-value="value" label="To this network" :rules="[v => !!v || 'Network is required']"
                                        persistent-hint required />
                                </div>
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions v-if="wild">
                    <v-spacer />
                    <v-btn :disabled="!valid" color="#004000" @click="createWilderness">
                        Submit
                        <v-icon end>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="#000040" @click="dialogWilderness = false">
                        Cancel
                        <v-icon end>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-container>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { useSubscriptionStore } from '@/stores/subscription'
import { useServiceStore } from '@/stores/service'
import { useServerStore } from '@/stores/server'
import { useNetStore } from '@/stores/net'
import { useWildnetStore } from '@/stores/wildnet'
import { useAuthStore } from '@/stores/auth'
import TokenService from '../services/token.service'
import ApiService from '../services/api.service'
import { formatDate } from '@/utils/formatDate'

const subscriptionStore = useSubscriptionStore()
const serviceStore = useServiceStore()
const serverStore = useServerStore()
const netStore = useNetStore()
const wildnetStore = useWildnetStore()
const authStore = useAuthStore()

const listView = ref(true)
const dialogCreateService = ref(false)
const dialogCreateMultihop = ref(false)
const dialogWilderness = ref(false)
const wildServer = ref('')
const wildList = ref({})
const wild = ref(false)
const inDelete = ref(false)
const credits = ref(0)
const netList = ref({})
const serverList = ref({})
const ingressList = ref({})
const egressList = ref({})
const server = ref(null)
const ingressServer = ref(null)
const egressServer = ref(null)
const subscription = ref(null)
const service = ref(null)
const ingress = ref(null)
const egress = ref(null)
const valid = ref(false)
const refreshing = ref(false)
const search = ref('')

const dnsList = ref({
    items: [
        { text: 'Google DNS', value: ['8.8.8.8', '8.8.4.4'] },
        { text: 'Cloudflare DNS', value: ['1.1.1.1', '1.0.0.1'] },
        { text: 'OpenDNS DNS', value: ['208.67.222.222', '208.67.220.222'] },
        { text: 'OpenDNS Family Shield DNS', value: ['208.67.222.123', '208.67.220.123'] },
        { text: 'Quad9 DNS', value: ['9.9.9.9', '149.112.112.112'] },
        { text: 'AdGuard DNS', value: ['94.140.14.14', '94.140.15.15'] },
    ]
})

const svcList = ref({
    items: [
        { text: 'Tunnel Service (tunnel all traffic through the Service Host)', value: 'Tunnel' },
        { text: 'Relay Service  (allows all machines in net to communicate with each other)', value: 'Relay' },
    ]
})

const headers = [
    { title: 'Name', key: 'name' },
    { title: 'Description', key: 'description' },
    { title: 'Issued', key: 'issued' },
    { title: 'Credits', key: 'credits' },
    { title: 'Status', key: 'status' },
    { title: 'Actions', key: 'action', sortable: false },
]

const bottom_headers = [
    { title: 'Name', key: 'device.name' },
    { title: 'Location', key: 'description' },
    { title: 'Service', key: 'serviceType' },
    { title: 'Created', key: 'created' },
    { title: 'Status', key: 'status' },
    { title: 'Actions', key: 'action', sortable: false },
]

watch(() => wildnetStore.nets, (w) => {
    const wl = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < w.length; i++) {
        wl.items[i] = { text: w[i].netName, value: w[i].id }
    }
    wildList.value = wl
})

watch(() => serverStore.servers, (val) => {
    const sl = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < val.length; i++) {
        sl.items[i] = { text: val[i].description, value: val[i].name }
    }
    serverList.value = sl
})

onMounted(() => {
    ApiService.setServer()
    ApiService.setHeader()

    netStore.readAll()
    serverStore.read()
    subscriptionStore.read(authStore.user.email)
    serviceStore.read(authStore.user.email)

    if (TokenService.getWildServer()) {
        wildServer.value = TokenService.getWildServer()
        wild.value = true
        wildnetStore.readAll()
        if (localStorage.getItem('wilderness_reopen')) {
            localStorage.removeItem('wilderness_reopen')
            dialogWilderness.value = true
        }
    } else {
        wild.value = false
    }
})

function Refresh() {
    subscriptionStore.read(authStore.user.email)
    serviceStore.read(authStore.user.email)
    serverStore.read()
}

function Refreshing() {
    refreshing.value = true
    Refresh()
    for (let i = 0; i < 5; i++) {
        if (refreshing.value) {
            setTimeout(() => {
                Refresh()
            }, (i + 1) * 2000)
            if (i == 4) {
                refreshing.value = false
            }
        } else {
            break
        }
    }
}

function startCreateService() {
    credits.value = 0
    for (let i = 0; i < subscriptionStore.subscriptions.length; i++) {
        if (subscriptionStore.subscriptions[i].status == 'active') {
            credits.value += subscriptionStore.subscriptions[i].credits
        }
    }
    if (credits.value <= serviceStore.services.length) {
        alert('You have exceeded your credit limit. Please purchase more credits to create a new service.')
        return
    }
    dialogCreateService.value = true
    service.value = {
        name: '',
        email: authStore.user.email,
    }
    netList.value = {
        selected: { text: '', value: '' },
        items: []
    }
    netList.value.items[0] = { text: 'New Network', value: '' }
    for (let i = 0; i < netStore.nets.length; i++) {
        netList.value.items[i + 1] = { text: netStore.nets[i].netName, value: netStore.nets[i].id }
    }
    netList.value.selected = netList.value.items[0]

    serverList.value = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < serverStore.servers.length; i++) {
        serverList.value.items[i] = { text: serverStore.servers[i].description, value: serverStore.servers[i].name }
    }
}

function startCreateMultihop() {
    credits.value = 0
    for (let i = 0; i < subscriptionStore.subscriptions.length; i++) {
        if (subscriptionStore.subscriptions[i].status == 'active') {
            credits.value += subscriptionStore.subscriptions[i].credits
        }
    }
    if ((credits.value - 1) <= serviceStore.services.length) {
        alert('Multihop requires 2 credits and exceeds your current limit. Please purchase more credits to create a new multihop service.')
        return
    }
    dialogCreateMultihop.value = true
    ingress.value = { name: '', email: authStore.user.email }
    egress.value = { name: '', email: authStore.user.email }
    netList.value = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < netStore.nets.length; i++) {
        netList.value.items[i] = { text: netStore.nets[i].netName, value: netStore.nets[i].id }
    }
    netList.value.selected = netList.value.items[-1]

    ingressList.value = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < serverStore.servers.length; i++) {
        ingressList.value.items[i] = { text: serverStore.servers[i].description, value: serverStore.servers[i].name }
    }
    egressList.value = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < serverStore.servers.length; i++) {
        egressList.value.items[i] = { text: serverStore.servers[i].description, value: serverStore.servers[i].name }
    }
}

function create() {
    for (let i = 0; i < serverList.value.items.length; i++) {
        if (serverList.value.items[i].value == serverList.value.selected.value) {
            server.value = serverStore.servers[i]
        }
    }

    const range = server.value.portMax - server.value.portMin + 1
    const port = server.value.portMin + Math.floor(Math.random() * range)

    service.value.defaultSubnet = server.value.defaultSubnet
    service.value.servicePort = port
    service.value.net = {}
    service.value.net.netName = netList.value.selected.value
    service.value.net.default = {}
    service.value.net.default.dns = []
    service.value.net.default.dns[0] = dnsList.value.selected.value[0]
    service.value.net.default.dns[1] = dnsList.value.selected.value[1]

    service.value.vpn = {}
    service.value.vpn.netName = serverList.value.selected.value
    service.value.vpn.current = {}
    service.value.vpn.current.dns = []
    service.value.vpn.current.dns[0] = dnsList.value.selected.value[0]
    service.value.vpn.current.dns[1] = dnsList.value.selected.value[1]
    service.value.vpn.current.endpoint = server.value.ipAddress + ':' + port
    service.value.vpn.current.listenPort = port

    service.value.description = server.value.description
    service.value.name = server.value.name
    service.value.serviceGroup = server.value.serviceGroup
    service.value.apiKey = server.value.serviceApiKey
    service.value.serviceType = svcList.value.selected.value

    if (service.value.net.netName != '') {
        service.value.net.id = netList.value.selected.value
    } else {
        service.value.net.id = ''
    }

    serviceStore.create(service.value)
    dialogCreateService.value = false
    Refreshing()
}

function create_multihop() {
    for (let i = 0; i < ingressList.value.items.length; i++) {
        if (ingressList.value.items[i].value == ingressList.value.selected.value) {
            ingressServer.value = serverStore.servers[i]
        }
    }
    for (let i = 0; i < egressList.value.items.length; i++) {
        if (egressList.value.items[i].value == egressList.value.selected.value) {
            egressServer.value = serverStore.servers[i]
        }
    }

    const rangeI = ingressServer.value.portMax - ingressServer.value.portMin + 1
    const portI = ingressServer.value.portMin + Math.floor(Math.random() * rangeI)

    ingress.value.defaultSubnet = ingressServer.value.defaultSubnet
    ingress.value.servicePort = portI
    ingress.value.description = ingressServer.value.description + ' (ingress)'
    ingress.value.name = ingressServer.value.name
    ingress.value.serviceGroup = ingressServer.value.serviceGroup
    ingress.value.apiKey = ingressServer.value.serviceApiKey
    ingress.value.net = {}
    ingress.value.net.netName = netList.value.selected.value
    ingress.value.net.default = {}
    ingress.value.net.default.dns = []
    ingress.value.vpn = {}
    ingress.value.vpn.current = {}
    ingress.value.vpn.current.dns = []
    ingress.value.vpn.current.endpoint = ingressServer.value.ipAddress + ':' + portI
    ingress.value.vpn.current.listenPort = portI
    ingress.value.serviceType = 'Ingress'

    if (ingress.value.net.netName != '') {
        ingress.value.net.id = netList.value.selected.value
    } else {
        ingress.value.net.id = ''
    }

    const rangeE = egressServer.value.portMax - egressServer.value.portMin + 1
    const portE = egressServer.value.portMin + Math.floor(Math.random() * rangeE)

    egress.value.defaultSubnet = ingressServer.value.defaultSubnet
    egress.value.servicePort = portE
    egress.value.description = egressServer.value.description + ' (egress)'
    egress.value.name = egressServer.value.name
    egress.value.serviceGroup = egressServer.value.serviceGroup
    egress.value.apiKey = egressServer.value.serviceApiKey
    egress.value.net = {}
    egress.value.net.netName = netList.value.selected.text
    egress.value.net.default = {}
    egress.value.net.default.dns = []
    egress.value.vpn = {}
    egress.value.vpn.current = {}
    egress.value.vpn.current.dns = []
    egress.value.vpn.current.endpoint = egressServer.value.ipAddress + ':' + portE
    egress.value.vpn.current.listenPort = portE
    egress.value.serviceType = 'Egress'

    if (egress.value.net.netName != '') {
        egress.value.net.id = netList.value.selected.value
    } else {
        egress.value.net.id = ''
    }

    serviceStore.create(ingress.value)
    serviceStore.create(egress.value)
    dialogCreateMultihop.value = false
    Refreshing()
}

function startCreateWilderness() {
    const storedServer = TokenService.getWildServer()
    if (storedServer && wild.value == false) {
        wildServer.value = storedServer
        wild.value = true
        wildnetStore.readAll()
    }
    dialogWilderness.value = true
}

function loginWild() {
    if (wildServer.value == '') {
        serviceStore.error = 'Server name is required'
        return
    }
    if (!wildServer.value.toLowerCase().startsWith('http')) {
        serviceStore.error = 'Server name should start with http or https'
        return
    }
    if (wildServer.value.endsWith('/')) {
        wildServer.value = wildServer.value.slice(0, -1)
    }
    dialogWilderness.value = false
    localStorage.setItem('wilderness_reopen', 'true')
    window.location.replace(window.location.origin + '/api/v1.0/auth/redirect?url=' + wildServer.value)
}

async function createWilderness() {
    for (let i = 0; i < serverList.value.items.length; i++) {
        if (serverList.value.items[i].value == serverList.value.selected.value) {
            server.value = serverStore.servers[i]
        }
    }

    const range = server.value.portMax - server.value.portMin + 1
    const port = server.value.portMin + Math.floor(Math.random() * range)

    service.value = {}
    service.value.defaultSubnet = server.value.defaultSubnet
    service.value.servicePort = port
    service.value.net = {}
    service.value.net.netName = wildList.value.selected.text
    service.value.net.id = wildList.value.selected.value
    service.value.net.default = {}
    service.value.net.default.dns = []

    const vpn = {
        name: server.value.name + '.' + wildList.value.selected.text,
        enable: true,
        netName: wildList.value.selected.text,
        netId: wildList.value.selected.value,
        type: 'Service',
        description: server.value.description,
        current: {
            endpoint: server.value.ipAddress + ':' + port,
            listenPort: port,
            subnetRouting: true,
            persistentKeepalive: 23,
        }
    }

    service.value.description = server.value.description
    service.value.name = server.value.name
    service.value.serviceGroup = server.value.serviceGroup
    service.value.serviceType = svcList.value.selected.value
    service.value.email = authStore.user.email
    service.value.server = wildServer.value

    const device = {
        name: svcList.value.selected.value.toLowerCase() + '.' + server.value.name,
        description: `nettica.com ${svcList.value.selected.value.toLowerCase()} service in ${server.value.description}`,
        enable: true,
        type: 'Service',
    }

    ApiService.setWildServer()
    ApiService.setWildHeader()

    ApiService.get('/net/' + wildList.value.selected.value)
        .then(n => {
            service.value.net = n
            vpn.current.allowedIPs = n.default.allowedIPs
            if (svcList.value.selected.value == 'Tunnel') {
                if (vpn.current.allowedIPs) {
                    vpn.current.allowedIPs.push('0.0.0.0/0')
                } else {
                    vpn.current.allowedIPs = ['0.0.0.0/0']
                }
            }

            ApiService.post('/device', device)
                .then(d => {
                    serviceStore.error = `Device created: ${d.name}`
                    vpn.deviceid = d.id

                    ApiService.post('/vpn', vpn)
                        .then(v => {
                            serviceStore.error = `VPN created: ${v.name}`
                            service.value.device = d
                            service.value.vpn = v

                            ApiService.setServer()
                            ApiService.setHeader()
                            ApiService.post('/service', service.value)
                                .then(s => {
                                    serviceStore.error = `Service created: ${s.name}`
                                    Refreshing()
                                })
                                .catch(error => {
                                    if (error.response) serviceStore.error = error.response.data.error
                                })
                        })
                        .catch(error => {
                            if (error.response) serviceStore.error = error.response.data.error
                        })
                })
                .catch(error => {
                    if (error.response) serviceStore.error = error.response.data.error
                })
        })
        .catch(error => {
            if (error.response) serviceStore.error = error.response.data.error
        })

    dialogWilderness.value = false
}

function removeSubscription(item) {
    inDelete.value = true
    if (confirm(`Do you really want to delete ${item.name} (not recommended)? This is an irreversible action.`)) {
        subscriptionStore.delete(item)
    }
}

function remove(item) {
    inDelete.value = true
    if (confirm(`Do you really want to delete ${item.name} ?`)) {
        serviceStore.delete(item)
    }
}

function startUpdateService(event, { item }) {
    if (inDelete.value == true) {
        inDelete.value = false
        return
    }
    service.value = item
}
</script>
