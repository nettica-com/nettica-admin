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
                    <v-col cols="4">
                        Networks
                    </v-col>
                    <v-col cols="4">
                        <v-text-field v-model="search" append-inner-icon="mdi-magnify" label="Search"
                            hide-details></v-text-field>
                    </v-col>
                    <v-col cols="4" class="text-right">
                        <v-btn color="#004000" @click="startCreate">
                            Create
                            <span class="material-symbols-outlined">hub</span>
                        </v-btn>
                    </v-col>
                </v-row>
            </v-card-title>
            <div v-if="friendly">
                <v-alert type="info" color="#336699" closable>
                    No networks found. <a style="color:white;" @click="createDefaultNetwork()">Click here to create your first network.</a>
                </v-alert>
            </div>
            <D3NetworkGraph v-if="!friendly" class="network" :net-nodes="nodes" :net-links="links" :options="options" />
            <v-divider v-if="!friendly"></v-divider>
            <v-row v-if="!friendly" style="padding-top: 12px;">
                <v-col cols="6">
                    <VTreeview v-if="showTree" :items="items" :search="search" :filter-fn="filter"
                        v-model:activated="active" v-model:opened="open"
                        item-title="name" item-value="id"
                        activatable hoverable @update:activated="loadNetwork">
                        <template v-slot:prepend="{ item }">
                            <span v-if="item.symbol" class="material-symbols-outlined">{{ item.symbol }}</span>
                            <v-icon v-else>{{ item.icon }}</v-icon>
                        </template>
                        <template v-slot:title="{ item }">
                            <table>
                                <tr><td :style="{ color: item.enabled ? 'white' : 'gray' }">
                                    {{ item.name }}
                                </td></tr>
                                <tr v-if="item.isNet"><td class="gray" style="font-size: small;">
                                    {{ item.description }}
                                </td></tr>
                                <tr v-else><td class="gray" style="font-size: small;">
                                    {{ item.vpn.current.address.join(', ') }}
                                </td></tr>
                            </table>
                        </template>
                        <template v-slot:append="{ item }">
                            <v-spacer></v-spacer>
                            <v-btn v-if="item.isNet" icon @click="startAddDevice(item.net)">
                                <v-tooltip text="Add device to this network" location="bottom">
                                    <template #activator="{ props: tooltipProps }">
                                        <v-icon v-bind="tooltipProps" color="#336699">mdi-plus-circle</v-icon>
                                    </template>
                                </v-tooltip>
                            </v-btn>
                        </template>
                    </VTreeview>
                </v-col>
                <v-divider vertical></v-divider>
                <v-col cols="6" class="text-center">
                    <div v-if="!selected" class="text-h6 text-grey-lighten-1 font-weight-light"
                        style="align-self: center;">
                    </div>
                    <v-card v-else-if="selected.isNet" :key="selected.id" class="px-3 mx-auto" flat>
                        <v-form autocomplete="off">
                        <v-card-text width="550">
                            <v-icon size="50" class="material-symbols-outlined">hub</v-icon>
                            <h3 class="text-h5 mb-2">
                                {{ selected.name }}
                            </h3>
                        </v-card-text>
                        <v-divider></v-divider>

                        <v-row class="px-3" width="600">
                            <v-col flex>
                                <v-text-field v-model="selected.net.description" label="Description" />
                                <v-combobox v-model="selected.net.tags" chips closable-chips
                                    hint="Enter a tag, hit tab, hit enter." label="Tags" multiple />
                                <v-combobox v-model="selected.net.default.address" chips closable-chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Addresses" multiple />
                                <v-combobox v-model="selected.net.default.allowedIPs" chips closable-chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple />
                                <v-combobox v-model="selected.net.default.dns" chips closable-chips
                                    hint="Enter IP address(es) and hit enter or leave empty."
                                    label="DNS servers for this network" multiple />
                                <v-text-field type="number" v-model="selected.net.default.mtu"
                                    label="MTU" hint="Leave at 0 for auto, 1350 for IPv6 or if problems occur" />
                                <v-text-field v-model="selected.net.accountid" label="Account ID" readonly />
                                <v-text-field v-model="selected.net.id" label="Network ID" readonly />
                                <v-text-field label="Preshared Key" readonly>
                                    <template v-slot:append-inner>
                                        <v-icon title="Copy to clipboard" @click="copy(selected.net.default.presharedKey)">mdi-content-copy</v-icon>
                                        <v-icon title="Regenerate preshared key" @click="regenerate(selected.net)">mdi-refresh</v-icon>
                                    </template>
                                </v-text-field>
                                <v-text-field type="number" v-model="selected.net.default.persistentKeepalive"
                                    label="Persistent keepalive"
                                    hint="To disable, set to 0.  Recommended value 29 (seconds)" />
                                <span style="font-size: small;">Network Policy</span><v-divider></v-divider>
                                <table>
                                    <tr>
                                        <td colspan="2">
                                            <v-switch v-model="selected.net.forceUpdate" color="#004000" inset
                                                :label="selected.net.forceUpdate ? 'Force updates immediately' : 'Applies to new devices only'" />
                                        </td>
                                    </tr>
                                    <tr>
                                        <td colspan="2">
                                            <v-switch v-model="selected.net.policies.userEndpoints" color="#004000" inset
                                                :label="selected.net.policies.userEndpoints ? 'Only Admins can create Endpoints' : 'Users can create Endpoints'" />
                                        </td>
                                    </tr>
                                    <tr>
                                        <td colspan="2">
                                            <v-switch v-model="selected.net.policies.onlyEndpoints" color="#004000" inset
                                                :label="selected.net.policies.onlyEndpoints ? 'Clients can only see Endpoints' : 'Clients can see all devices'" />
                                        </td>
                                    </tr>
                                </table>
                                <span style="font-size: small;">Client Defaults</span><v-divider></v-divider>
                                <table width="100%">
                                    <tr>
                                        <td>
                                            <v-switch v-model="selected.net.default.syncEndpoint" color="#004000" inset
                                                label="Sync Endpoint" />
                                        </td>
                                        <td>
                                            <v-switch v-model="selected.net.default.hasSSH" color="#004000" inset
                                                label="SSH" />
                                        </td>
                                    </tr>
                                    <tr>
                                        <td>
                                            <v-switch v-model="selected.net.default.upnp" color="#004000" inset
                                                label="Enable UPnP" />
                                        </td>
                                        <td>
                                            <v-switch v-model="selected.net.default.hasRDP" color="#004000" inset
                                                label="Remote Desktop" />
                                        </td>
                                    </tr>
                                    <tr>
                                        <td>
                                            <v-switch v-model="selected.net.default.failsafe" color="#004000" inset
                                                label="FailSafe" />
                                        </td>
                                        <td>
                                            <v-switch v-model="selected.net.default.enableDns" color="#004000" inset
                                                label="Enable Nettica DNS" />
                                        </td>
                                    </tr>
                                </table>
                                <v-divider v-if="isOwner"></v-divider>
                                <table width="100%" v-if="isOwner">
                                    <tr>
                                        <td colspan="2">
                                            <v-switch v-model="selected.net.critical" color="red" inset
                                                :label="selected.net.critical ? 'This is a Critical Network' : 'This is a Normal Network'" />
                                        </td>
                                    </tr>
                                </table>
                                <v-divider></v-divider>
                                <p class="text-caption">Created by {{ selected.net.createdBy }} at {{
                                    formatDate(selected.net.created) }}<br />
                                    Last update by {{ selected.net.updatedBy }} at {{ formatDate(selected.net.updated) }}</p>
                            </v-col>
                        </v-row>
                        <v-card-actions>
                            <v-container>
                                <v-row>
                                    <v-col>
                                        <v-btn color="#004000" @click="update(selected.net)">
                                            Save
                                            <v-icon end>mdi-check-outline</v-icon>
                                        </v-btn>
                                    </v-col>
                                    <v-col>
                                        <v-btn color="#400000" @click="remove(selected.net)">
                                            Delete
                                            <v-icon end>mdi-delete-outline</v-icon>
                                        </v-btn>
                                    </v-col>
                                </v-row>
                            </v-container>
                        </v-card-actions>
                        </v-form>
                    </v-card>
                    <v-card v-else-if="selected && !selected.isNet">
                        <v-form autocomplete="off">
                        <v-card-text width="600" class="px-3">
                            <v-icon size="50" class="material-symbols-outlined">network_node</v-icon>
                            <h3 class="text-h5 mb-2">
                                {{ selected.name }}
                            </h3>
                        </v-card-text>
                        <v-divider></v-divider>
                        <v-row class="px-3" width="600">
                            <v-col flex>
                                <v-text-field v-model="selected.vpn.name" label="DNS name" :readonly="!inEdit"
                                 :rules="[rules.required, rules.host]" />
                                <v-combobox :readonly="!inEdit" v-model="selected.vpn.current.address" chips closable-chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Addresses" multiple />
                                <v-combobox :readonly="!inEdit" v-model="selected.vpn.current.dns" chips closable-chips
                                    hint="Enter IP address(es) and hit enter or leave empty."
                                    label="DNS servers for this network" multiple />
                                <v-combobox v-if="selected.vpn.type == 'Service'" v-model="selected.vpn.current.allowedIPs" chips closable-chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple
                                    :readonly="!inEdit" />
                                <v-text-field :readonly="!inEdit" v-model="selected.vpn.current.endpoint"
                                    label="Public endpoint for clients" :rules="[ rules.ipport ]" />
                                <v-text-field type="number" v-model="selected.vpn.current.mtu"
                                    label="MTU" hint="Leave at 0 for auto, 1350 for IPv6 or problems occur" />
                                <v-text-field type="number" v-model="selected.vpn.current.persistentKeepalive"
                                    label="Persistent keepalive" hint="To disable, set to 0.  Recommended value 23 (seconds)" />
                                <v-switch v-model="selected.vpn.enable" color="#004000" inset
                                    :label="selected.vpn.enable ? 'Enabled' : 'Disabled'" :readonly="!inEdit" />
                                <p class="text-caption">Created by {{ selected.vpn.createdBy }} at {{ formatDate(selected.vpn.created) }}<br />
                                    Last update by {{ selected.vpn.updatedBy }} at {{ formatDate(selected.vpn.updated) }}</p>
                            </v-col>
                        </v-row>
                        <v-expansion-panels v-if="inEdit && !(selected.vpn.type == 'Service')">
                            <v-expansion-panel>
                                <v-expansion-panel-title>Advanced Configuration</v-expansion-panel-title>
                                <v-expansion-panel-text>
                                    <div class="d-flex flex-no-wrap justify-space-between">
                                        <v-col cols="12">
                                            <v-combobox v-model="selected.vpn.current.allowedIPs" chips closable-chips
                                                hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple />
                                            <v-text-field v-model="selected.vpn.accountid" label="Account ID" readonly />
                                            <v-text-field v-model="selected.vpn.id" label="VPN ID" readonly />
                                            <v-text-field v-model="selected.vpn.netid" label="Network ID" readonly />
                                            <v-text-field v-model="selected.vpn.deviceid" label="Device ID" readonly />
                                            <v-combobox v-model="selected.vpn.role" :items="['', 'Ingress', 'Egress']"
                                                label="Role" />
                                            <v-text-field v-model="selected.vpn.current.table" label="Table" />
                                            <v-text-field v-model="selected.vpn.current.publicKey" label="Public key" />
                                            <v-text-field v-if="!editPrivate" label="Private key" readonly append-inner-icon="mdi-square-edit-outline" @click:append-inner="editPrivate = true" />
                                            <v-text-field v-if="editPrivate" v-model="selected.vpn.current.privateKey" label="Private key"
                                                hint="Clear this field to have the client manage its private key" />
                                            <v-text-field label="Preshared Key" readonly
                                                append-inner-icon="mdi-content-copy" @click:append-inner="copy(selected.vpn.current.presharedKey)" />
                                            <v-textarea v-model="selected.vpn.current.preUp" label="PreUp Script"
                                                hint="Command to run before starting VPN" />
                                            <v-textarea v-model="selected.vpn.current.postUp" label="PostUp Script"
                                                hint="Command to run after starting VPN" />
                                            <v-textarea v-model="selected.vpn.current.preDown" label="PreDown Script"
                                                hint="Command to run before stopping VPN" />
                                            <v-textarea v-model="selected.vpn.current.postDown" label="PostDown Script"
                                                hint="Command to run after stopping VPN" />
                                            <v-switch v-model="selected.vpn.current.subnetRouting" color="#004000" inset
                                                label="Enable subnet routing" />
                                            <v-divider></v-divider>
                                            <table width="100%">
                                                <tr>
                                                    <td>
                                                        <v-switch v-model="selected.vpn.current.syncEndpoint" color="#004000" inset
                                                            label="Sync Endpoint" />
                                                    </td>
                                                    <td>
                                                        <v-switch v-model="selected.vpn.current.hasSSH" color="#004000" inset
                                                            label="SSH" />
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <td>
                                                        <v-switch v-model="selected.vpn.current.upnp" color="#004000" inset
                                                            label="Enable UPnP" />
                                                    </td>
                                                    <td>
                                                        <v-switch v-model="selected.vpn.current.hasRDP" color="#004000" inset
                                                            label="Remote Desktop" />
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <td>
                                                        <v-switch v-model="selected.vpn.current.failsafe" color="#004000" inset
                                                            label="FailSafe" />
                                                    </td>
                                                    <td>
                                                        <v-switch v-model="selected.vpn.current.enableDns" color="#004000" inset
                                                            label="Nettica DNS" />
                                                    </td>
                                                </tr>
                                            </table>
                                        </v-col>
                                    </div>
                                </v-expansion-panel-text>
                            </v-expansion-panel>
                        </v-expansion-panels>
                        <v-card>
                            <v-card-actions v-if="inEdit">
                                <v-btn color="#004000" @click="updateVPN(selected.vpn)">
                                    Submit
                                    <v-icon end>mdi-check-outline</v-icon>
                                </v-btn>
                                <v-btn color="#000040" @click="inEdit = false">
                                    Cancel
                                    <v-icon end>mdi-close-circle-outline</v-icon>
                                </v-btn>
                            </v-card-actions>
                            <v-card-actions v-else>
                                <v-container>
                                    <v-row>
                                        <v-col>
                                            <v-btn color="#004000" @click="forceFileDownload(selected.vpn)">
                                                Download
                                                <v-icon end>mdi-cloud-download-outline</v-icon>
                                            </v-btn>
                                        </v-col>
                                        <v-col>
                                            <v-btn class="px-3" color="#000040" @click="inEdit = true">
                                                Edit
                                                <v-icon end>mdi-pencil-outline</v-icon>
                                            </v-btn>
                                        </v-col>
                                        <v-col>
                                            <v-btn class="px-3" color="#400000" @click="removeVPN(selected.vpn)">
                                                Delete
                                                <v-icon end>mdi-delete-outline</v-icon>
                                            </v-btn>
                                        </v-col>
                                    </v-row>
                                </v-container>
                            </v-card-actions>
                        </v-card>
                        </v-form>
                    </v-card>
                </v-col>
            </v-row>
        </v-card>
        <v-dialog v-if="net" v-model="dialogCreate" max-width="550" persistent @keydown.esc="dialogCreate = false">
            <v-card>
                <v-card-title class="headline">Create New Network</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-text-field v-model="net.netName"
                                    :rules="[ rules.required, rules.host ]" required>
                                    <template #label>
                                      <span class="text-red"><strong>* </strong></span>Network Name
                                    </template>
                                </v-text-field>
                                <v-combobox v-model="net.default.address" :items="net.default.address"
                                    :rules="[rules.required, rules.cidr]" multiple chips closable-chips persistent-hint required>
                                    <template #label>
                                      <span class="text-red"><strong>* </strong></span>IP subnet for this network (ex. 10.10.10.0/24)
                                    </template>
                                </v-combobox>
                                <v-select return-object v-model="acntList.selected" :items="acntList.items" item-title="text"
                                    item-value="value" label="For this account"
                                    :rules="[v => !!v || 'Account is required']" persistent-hint />
                                <v-text-field v-model="net.description" label="Description" />
                                <v-combobox v-model="net.tags" chips closable-chips
                                    hint="Enter a tag, hit tab, hit enter." label="Tags" multiple />
                                <v-combobox v-model="net.default.dns" chips closable-chips
                                    hint="Enter the IP address of a global DNS resolver, hit tab, hit enter."
                                    label="DNS servers for this network" multiple />
                                <v-switch v-model="net.policies.userEndpoints" color="#004000" inset
                                    label="Users can create Endpoints" />
                                <v-switch v-model="net.policies.onlyEndpoints" color="#004000" inset
                                    label="Clients can only see Endpoints" />
                                <v-switch v-model="net.default.upnp" color="#004000" inset
                                    label="Enable UPnP where possible" />
                                <v-switch v-model="net.default.failsafe" color="#004000" inset
                                    label="Enable FailSafe" />
                                <v-switch v-model="net.default.enableDns" color="#004000" inset
                                    label="Enable Nettica DNS" />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="#004000" @click="create(net)">
                        Submit
                        <v-icon end>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="#000040" @click="dialogCreate = false">
                        Cancel
                        <v-icon end>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog v-if="net" v-model="dialogAddDevice" max-width="550" persistent @keydown.esc="dialogAddDevice = false">
            <v-card>
                <v-card-title class="headline">Add Device</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-text-field v-model="net.netName" label="To Network" readonly />
                                <v-select return-object v-model="deviceList.selected" :items="deviceList.items" @update:model-value="updateName"
                                    item-title="text" item-value="value" label="Add this device"
                                    :rules="[v => !!v || 'Device is required']" persistent-hint required />
                                <v-text-field v-model="vpn.name" label="DNS name for this device"
                                    :rules="[rules.required, rules.host]" required />
                                <v-text-field v-model="vpn.current.endpoint" label="Public endpoint for clients"
                                    :rules="[ rules.ipport ]" />
                                <v-switch v-model="vpn.enable" color="#004000" inset
                                    :label="vpn.enable ? 'Enable VPN after creation' : 'Disable VPN after creation'" />
                                <v-switch v-model="vpn.current.syncEndpoint" color="#004000" inset
                                    :label="vpn.current.syncEndpoint ? 'Automatically sync endpoint using server' : 'Do not sync endpoint using server'"
                                    :disabled="!(vpn.current.endpoint && vpn.current.endpoint.length > 0)" />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="#004000" @click="addDevice()">
                        Submit
                        <v-icon end>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="#000040" @click="dialogAddDevice = false">
                        Cancel
                        <v-icon end>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-container>
</template>

<style>
.gray {
    color: gray;
}
text {
    font-size: 12px;
    color: white;
    fill: white;
}
.node {
    fill: #336699;
    stroke: #5b81a7;
}
.link {
    color: white;
}
.net-svg {
    margin: 0 auto;
}
.network {
    display: flex;
    justify-content: center;
}
</style>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import D3NetworkGraph from './D3NetworkGraph.vue'
import { useVpnStore } from '@/stores/vpn'
import { useNetStore } from '@/stores/net'
import { useAccountStore } from '@/stores/account'
import { useDeviceStore } from '@/stores/device'
import { useAuthStore } from '@/stores/auth'
import ApiService from '@/services/api.service'
import { formatDate } from '@/utils/formatDate'

const vpnStore = useVpnStore()
const netStore = useNetStore()
const accountStore = useAccountStore()
const deviceStore = useDeviceStore()
const authStore = useAuthStore()

const acntList = ref({})
const showTree = ref(false)
const friendly = ref(false)
const items = ref([])
const active = ref([])
const open = ref([])
const deviceList = ref({})
const isOwner = ref(false)
const inEdit = ref(false)
const editPrivate = ref(false)
const dialogCreate = ref(false)
const dialogAddDevice = ref(false)
const publicSubnets = ref(false)
const noEdit = ref(false)
const net = ref(null)
const vpn = ref({
    name: '',
    current: { endpoint: '' }
})
const device = ref(null)
const valid = ref(false)
const refreshing = ref(false)
const search = ref('')
const nodes = ref([])
const links = ref([])
const nodeSize = ref(50)

const rules = {
    cidr: v => /((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b\/\b[0-9]{1,2}\b)|(\b(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}\b\/\b[0-9]{1,3}\b))(?: ((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b\/\b[0-9]{1,2}\b)|(\b(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}\b\/\b[0-9]{1,3}\b)))*/.test(v) || 'Enter a valid subnet',
    required: value => !!value || 'Required.',
    email: v => /.+@.+\..+/.test(v) || 'E-mail must be valid',
    host: v => /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/.test(v) || 'Only letters, numbers, dots and hyphens are allowed. Must start and end with a letter or number.',
    ipport: v => (!v || v && v.length == 0 || /^(((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5})|^$)(,\s+((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5})|^$))*)$/.test(v)) || 'If present, must be valid IPv4 or IPv6 address and port',
}

const options = computed(() => ({
    force: 4000,
    size: { w: Math.min(window.innerWidth, 1000), h: 300 },
    nodeSize: nodeSize.value,
    nodeLabels: true,
    linkLabels: true,
    canvas: false,
}))

const selected = computed(() => {
    if (!active.value.length) return undefined
    const id = active.value[0]
    for (let i = 0; i < items.value.length; i++) {
        if (items.value[i].id === id) return items.value[i]
        for (let j = 0; j < items.value[i].children.length; j++) {
            if (items.value[i].children[j].id === id) return items.value[i].children[j]
        }
    }
    const foundVpn = vpnStore.vpns.find(v => v.id === id)
    if (foundVpn) return foundVpn
    return items.value.find(item => item.id === id)
})

watch(() => netStore.nets, () => {
    buildTree()
    showTree.value = true
    friendly.value = netStore.nets.length === 0
})

watch(() => vpnStore.vpns, () => {
    buildTree()
    showTree.value = true
})

watch(() => accountStore.accounts, (val) => {
    acntList.value = { selected: { text: '', value: '' }, items: [] }
    for (let i = 0; i < val.length; i++) {
        if (val[i].parent === val[i].id) isOwner.value = true
        acntList.value.items[i] = { text: val[i].accountName + ' - ' + val[i].parent, value: val[i].parent }
    }
})

onMounted(() => {
    accountStore.readAll(authStore.user.email)
    netStore.readAll()
    vpnStore.readAll()
    deviceStore.readAll()
})

function Refresh() {
    accountStore.readAll(authStore.user.email)
    vpnStore.readAll()
    netStore.readAll()
}

function Refreshing() {
    refreshing.value = true
    Refresh()
    for (let i = 0; i < 5; i++) {
        if (refreshing.value) {
            setTimeout(() => { Refresh() }, (i + 1) * 2000)
            if (i === 4) refreshing.value = false
        } else {
            break
        }
    }
}

function filter(value, query, item) {
    const raw = item?.raw ?? item
    if (!query) return true
    if (raw.name && raw.name.toLowerCase().includes(query.toLowerCase())) return true
    if (raw.isNet && raw.net.tags != null) {
        for (const tag of raw.net.tags) {
            if (tag.toLowerCase().includes(query.toLowerCase())) return true
        }
    }
    if (!raw.isNet && raw.vpn?.name?.toLowerCase().includes(query.toLowerCase())) return true
    if (!raw.isNet && raw.vpn?.current?.address?.length > 0) {
        for (const addr of raw.vpn.current.address) {
            if (addr.toLowerCase().includes(query.toLowerCase())) return true
        }
    }
    return false
}

function buildTree() {
    items.value = []
    for (let i = 0; i < netStore.nets.length; i++) {
        items.value[i] = {
            id: netStore.nets[i].id,
            name: netStore.nets[i].netName,
            description: netStore.nets[i].description,
            net: netStore.nets[i],
            icon: 'mdi-network-outline',
            symbol: 'hub',
            isNet: true,
            enabled: true,
            children: []
        }
    }

    items.value.sort((a, b) => a.name.toUpperCase() < b.name.toUpperCase() ? -1 : a.name.toUpperCase() > b.name.toUpperCase() ? 1 : 0)

    for (let i = 0; i < items.value.length; i++) {
        let k = 0
        for (let j = 0; j < vpnStore.vpns.length; j++) {
            if (vpnStore.vpns[j].netName === items.value[i].name) {
                items.value[i].children[k] = {
                    id: vpnStore.vpns[j].id,
                    name: vpnStore.vpns[j].name,
                    address: vpnStore.vpns[j].current.address[0],
                    net: items.value[i].net,
                    vpn: vpnStore.vpns[j],
                    icon: 'mdi-network-outline',
                    symbol: 'network_node',
                    isNet: false,
                    isNode: true,
                    enabled: vpnStore.vpns[j].enable,
                }
                k++
            }
        }
    }
}

function loadNetwork() {
    const item = selected.value
    if (!item) return
    const n = item.net
    if (!n) return
    const name = n.netName
    let x = 0
    let l = 0
    links.value = []
    nodes.value = []
    const net_hosts = []
    for (let i = 0; i < vpnStore.vpns.length; i++) {
        if (vpnStore.vpns[i].netName === name) {
            net_hosts[x] = vpnStore.vpns[i]
            nodes.value[x] = { id: x, name: vpnStore.vpns[i].name }
            if (!vpnStore.vpns[i].current.endpoint) {
                nodes.value[x]._color = '#34adcd'
            } else {
                nodes.value[x]._color = '#83c44d'
            }
            if (vpnStore.vpns[i].role === 'Egress') nodes.value[x]._color = '#50C878'
            if (vpnStore.vpns[i].id === active.value[0]) nodes.value[x]._color = '#FF8C00'
            x++
        }
    }
    for (let i = 0; i < net_hosts.length; i++) {
        for (let j = 0; j < net_hosts.length; j++) {
            if (i !== j && net_hosts[j].current.endpoint !== '' && net_hosts[j].role !== 'Egress') {
                links.value[l] = { sid: i, tid: j, _color: 'white' }
                l++
            }
        }
    }
}

function startCreate() {
    net.value = {
        name: '',
        email: authStore.user.email,
        enable: true,
        netName: '',
        id: '',
        tags: [],
        accountid: '',
        default: { allowedIPs: [], address: [], enableDns: false, upnp: false, failsafe: false },
        policies: { userEndpoints: false, onlyEndpoints: false },
    }

    acntList.value = { selected: { text: '', value: '' }, items: [] }
    let sel = 0
    for (let i = 0; i < accountStore.accounts.length; i++) {
        acntList.value.items[i] = { text: accountStore.accounts[i].accountName + ' - ' + accountStore.accounts[i].parent, value: accountStore.accounts[i].parent }
        if (acntList.value.items[i].value === net.value.accountid) sel = i
    }
    acntList.value.selected = acntList.value.items[sel]
    dialogCreate.value = true
}

function create(n) {
    net.value = n
    net.value.accountid = acntList.value.selected.value
    dialogCreate.value = false
    netStore.create(n)
}

function createDefaultNetwork() {
    net.value = {
        name: 'Nettica',
        description: authStore.user.email + "'s first network",
        email: authStore.user.email,
        enable: true,
        netName: 'Nettica',
        id: '',
        tags: [],
        accountid: '',
        default: {
            allowedIPs: ['10.10.10.0/24'],
            address: ['10.10.10.0/24'],
            dns: ['8.8.8.8'],
            enableDns: false,
            upnp: true,
            failsafe: true,
        },
        policies: { userEndpoints: true, onlyEndpoints: false },
    }

    for (const acct of accountStore.accounts) {
        if (acct.id === acct.parent) {
            net.value.accountid = acct.parent
            break
        }
    }

    if (!net.value.accountid) {
        netStore.error = 'Cannot create a network for this account'
        return
    }

    netStore.create(net.value)
}

function remove(n) {
    noEdit.value = true
    if (confirm(`Do you really want to delete ${n.netName}?`)) {
        netStore.delete(n)
    }
}

function startAddDevice(n) {
    net.value = n
    vpn.value = {
        name: '',
        netid: n.id,
        accountid: n.accountid,
        email: authStore.user.email,
        enable: true,
        tags: [],
        current: { syncEndpoint: false },
    }

    deviceList.value = { selected: { text: '', value: '' }, items: [] }
    for (const d of deviceStore.devices) {
        let found = false
        if (d.vpns) {
            for (const v of d.vpns) {
                if (v.netid === n.id) { found = true; break }
            }
        }
        if (!found) deviceList.value.items.push({ text: d.name, value: d.id })
    }
    deviceList.value.items.sort((a, b) => a.text.toUpperCase() < b.text.toUpperCase() ? -1 : a.text.toUpperCase() > b.text.toUpperCase() ? 1 : 0)
    dialogAddDevice.value = true
}

function updateName(item) {
    device.value = deviceStore.devices.find(d => d.id === item.value)
    vpn.value.name = device.value.name + '.' + net.value.netName
}

async function addDevice() {
    vpn.value.current.listenPort = 0
    if (vpn.value.current.endpoint && vpn.value.current.endpoint.includes(':')) {
        const parts = vpn.value.current.endpoint.split(':')
        vpn.value.current.listenPort = parseInt(parts[parts.length - 1], 10)
    }
    vpn.value.netName = net.value.netName
    vpn.value.netid = net.value.id
    vpn.value.accountid = net.value.accountid
    vpn.value.deviceid = device.value.id
    dialogAddDevice.value = false
    vpnStore.create(vpn.value)
}

function update(n) {
    net.value = n
    net.value.default.listenPort = 0
    net.value.default.persistentKeepalive = parseInt(net.value.default.persistentKeepalive, 10)
    net.value.default.mtu = parseInt(net.value.default.mtu, 10)

    for (const addr of n.default.address) {
        if (typeof addr === 'string' && !addr.includes('/')) {
            netStore.error = 'Invalid CIDR detected, please correct before submitting'
            return
        }
    }

    if (n.default.allowedIPs.length < 1) {
        net.value.default.allowedIPs = net.value.default.address
    }

    netStore.update(n)
    netStore.update_net(n)
}

function updateVPN(v) {
    vpn.value = v
    vpn.value.current.listenPort = 0
    if (vpn.value.current.endpoint && vpn.value.current.endpoint.includes(':')) {
        const parts = vpn.value.current.endpoint.split(':')
        vpn.value.current.listenPort = parseInt(parts[parts.length - 1], 10)
    }
    vpn.value.current.persistentKeepalive = parseInt(vpn.value.current.persistentKeepalive, 10)
    vpn.value.current.mtu = parseInt(vpn.value.current.mtu, 10)
    inEdit.value = false
    editPrivate.value = false
    vpnStore.update(vpn.value)
}

async function removeVPN(v) {
    if (confirm(`Do you really want to delete ${v.name} from ${v.netName}?`)) {
        await vpnStore.delete(v)
        Refreshing()
    }
}

function copy(text) {
    navigator.clipboard.writeText(text).then(() => netStore.error = 'Copied to clipboard')
}

function regenerate(n) {
    if (confirm(`Do you really want to regenerate the preshared key for ${n.netName}? This will take effect immediately.`)) {
        const chars = 'ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789'
        let key = ''
        for (let i = 0; i < 32; i++) key += chars.charAt(Math.floor(Math.random() * chars.length))
        key = btoa(key)
        n.default.presharedKey = key
        n.forceUpdate = true

        ApiService.patch(`/net/${n.id}`, n)
            .then(updated => {
                netStore.update_net(updated)
                netStore.error = 'Preshared key has been regenerated and distributed.  Clients may need to reconnect.'
                Refresh()
            })
            .catch(() => netStore.error = 'Failed to regenerate preshared key')
    }
}

async function forceFileDownload(v) {
    ApiService.getWithConfig(`/vpn/${v.id}/config?qrcode=false`, { responseType: 'arraybuffer' })
        .then(config => {
            const url = window.URL.createObjectURL(new Blob([config]))
            const link = document.createElement('a')
            link.href = url
            link.setAttribute('download', v.name.split(' ').join('-') + '-' + v.netName.split(' ').join('-') + '.zip')
            document.body.appendChild(link)
            link.click()
        })
        .catch(() => netStore.error = 'Failed to download device config')
}
</script>
