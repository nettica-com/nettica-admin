<template>
    <v-container style="padding-top:0px">
        <v-row><v-col cols="12">
                <div>
                    <v-btn class="mb-3 mt-0" @click="Refresh()">
                        <v-icon>mdi-refresh</v-icon>
                        Refresh
                    </v-btn>
                </div>
                <v-card>
                    <v-card-title>
                        <v-row>
                            <v-col cols="4">Devices</v-col>
                            <v-col cols="4">
                                <v-text-field v-model="search" append-inner-icon="mdi-magnify" label="Search"
                                    hide-details></v-text-field>
                            </v-col>
                            <v-col cols="4" class="text-right">
                                <v-btn color="#004000" @click="startCreate">
                                    Add Device
                                    <span class="material-symbols-outlined">devices</span>
                                </v-btn>
                            </v-col>
                        </v-row>
                    </v-card-title>
                    <div v-if="friendly">
                        <v-alert type="info" color="#336699" closable>
                            No devices found. <a style="color:white;" @click="startCreate">Click here to add your first
                                device.</a>
                        </v-alert>
                    </div>

                    <v-row v-if="!friendly">
                        <v-col cols="6">
                            <VTreeview v-if="showTree" :items="items" :search="search" :filter-fn="filter"
                                v-model:activated="active" v-model:opened="open" item-title="name" item-value="id"
                                activatable hoverable>
                                <template v-slot:prepend="{ item }">
                                    <span v-if="item.symbol && item.status == 'Online'"
                                        class="material-symbols-outlined" style="color:green;">{{ item.symbol }}</span>
                                    <span v-else-if="item.symbol && item.status == 'Offline'"
                                        class="material-symbols-outlined" style="color:red;">{{ item.symbol }}</span>
                                    <span
                                        v-else-if="item.symbol && item.status && item.status != 'Online' && item.status != 'Offline'"
                                        class="material-symbols-outlined" style="color:blue;">{{ item.symbol }}</span>
                                    <span v-else-if="item.symbol" class="material-symbols-outlined">{{ item.symbol
                                    }}</span>
                                    <v-icon v-else style="color:white;">
                                        {{ item.icon }}
                                    </v-icon>
                                </template>
                                <template v-slot:title="{ item }">
                                    <table>
                                        <tbody>
                                            <tr>
                                                <td :style="{ color: item.enabled ? 'white' : 'gray' }">
                                                    {{ item.name }}
                                                </td>
                                            </tr>
                                            <tr v-if="item.isDevice">
                                                <td class="gray" style="font-size: small;">
                                                    {{ item.description }}
                                                </td>
                                            </tr>
                                            <tr v-else>
                                                <td class="gray" style="font-size: small;">
                                                    {{ item.vpn.current.address.join(', ') }}
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </template>
                                <template v-slot:append="{ item }">
                                    <v-btn v-if="item.isDevice" icon @click="startAddVPN(item.device)">
                                        <v-tooltip text="Add network to this device" location="bottom">
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
                            <v-card v-else-if="selected.isDevice" :key="selected.id" class="px-3 mx-auto" flat>
                                <v-form autocomplete="off">
                                    <v-card-text width="600">
                                        <span class="material-symbols-outlined" style="font-size:50px;">{{
                                            selected.symbol }}</span>
                                        <h3 class="text-h5 mb-2">
                                            {{ selected.name }}
                                        </h3>
                                        <div class="mb-2">
                                            <v-icon v-if="selected.device.status == 'Online'"
                                                color="green">mdi-check-circle</v-icon>
                                            <v-icon v-else-if="selected.device.status == 'Offline'"
                                                color="red">mdi-close-circle</v-icon>
                                            <v-icon v-else color="blue">mdi-minus-circle</v-icon>
                                            {{ selected.device.status }}
                                        </div>
                                    </v-card-text>
                                    <v-divider></v-divider>

                                    <v-row class="px-3" width="600">
                                        <v-col flex>
                                            <div :hidden="selected.device.registered">
                                                <v-text-field v-model="selected.device.ezcode" label="EZ-Code"
                                                    :readonly="true" />
                                            </div>
                                            <v-text-field v-model="selected.device.description" label="Description"
                                                :readonly="!inEdit" />
                                            <v-combobox v-model="selected.device.tags" chips closable-chips
                                                hint="Enter a tag, hit tab, hit enter." label="Tags" multiple
                                                :readonly="!inEdit" />
                                            <v-select return-object v-model="selected.platform" :items="platforms.items"
                                                item-title="text" item-value="value" label="Platform of this device"
                                                persistent-hint :readonly="!inEdit" />
                                            <v-text-field v-model="selected.device.version" label="Version" readonly />
                                            <v-switch v-model="selected.device.enable" color="#004000" inset
                                                :label="selected.device.enable ? 'Enabled' : 'Disabled'"
                                                :readonly="!inEdit" />
                                            <v-text-field v-model="selected.device.checkInterval" type="number"
                                                label="Check interval" hint="In seconds" :readonly="!inEdit" />
                                            <v-select return-object v-model="selected.accountid" :items="acntList.items"
                                                item-title="text" item-value="value" label="Account ID" persistent-hint
                                                :readonly="!inEdit" />
                                            <v-text-field v-model="selected.device.id" label="Device ID" readonly />
                                            <v-text-field v-model="selected.device.server" label="Server"
                                                :readonly="!inEdit" />
                                            <v-text-field v-if="!inEdit" label="API Key" readonly
                                                append-inner-icon="mdi-content-copy"
                                                @click:append-inner="copy(selected.device.apiKey)">* * *
                                                *</v-text-field>
                                            <v-text-field v-if="inEdit" v-model="selected.device.apiKey"
                                                label="API Key" />
                                            <v-text-field v-model="selected.device.instanceid"
                                                label="Instance ID (AWS | Oracle | GCP | Azure)" hint="Give your device superpowers. Enter the instance ID from your cloud provider" :readonly="!inEdit" />
                                            <div :hidden="!inEdit">
                                                <v-text-field v-model="selected.device.name"
                                                    label="Device friendly name"
                                                    :rules="[v => !!v || 'device name is required']" required />
                                            </div>
                                            <v-select return-object v-model="selected.logging" :items="logging.items"
                                                item-title="text" item-value="value" label="Logging" persistent-hint
                                                :readonly="!inEdit" />
                                            <p class="text-caption">Created by {{ selected.device.createdBy }} at {{
                                                formatDate(selected.device.created) }}<br />
                                                Last update by {{ selected.device.updatedBy }} at {{
                                                    formatDate(selected.device.updated)
                                                }}<br />
                                                <span v-if="selected.device.lastSeen">Last Seen on {{
                                                    formatDate(selected.device.lastSeen)
                                                }}</span>
                                            </p>
                                        </v-col>
                                    </v-row>
                                    <v-card-actions v-if="inEdit">
                                        <v-btn color="#004000" @click="updateDevice(selected)">
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
                                                    <v-btn color="#004000" @click="copyDeviceConfig(selected.device)">
                                                        Copy
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
                                                    <v-btn class="px-3" color="#400000"
                                                        @click="removeDevice(selected.device)">
                                                        Delete
                                                        <v-icon end>mdi-delete-outline</v-icon>
                                                    </v-btn>
                                                </v-col>
                                            </v-row>
                                        </v-container>
                                    </v-card-actions>
                                </v-form>
                            </v-card>
                            <v-card v-else-if="selected && !selected.isDevice">
                                <v-form autocomplete="off">
                                    <v-card-text width="600" class="px-3">
                                        <v-icon class="material-symbols-outlined" size="50">hub</v-icon>
                                        <h3 class="text-h5 mb-2">
                                            {{ selected.vpn.netName }}
                                        </h3>
                                    </v-card-text>
                                    <v-divider></v-divider>
                                    <v-row class="px-3" width="600">
                                        <v-col flex>
                                            <v-text-field v-model="selected.vpn.name" label="DNS name"
                                                :readonly="!inEdit" :rules="[rules.required, rules.host]" />
                                            <v-combobox :readonly="!inEdit" v-model="selected.vpn.current.address" chips
                                                closable-chips hint="Write IPv4 or IPv6 CIDR and hit enter"
                                                label="Addresses" multiple />
                                            <v-combobox :readonly="!inEdit" v-model="selected.vpn.current.dns" chips
                                                closable-chips
                                                hint="Enter IP address(es) of the DNS servers and hit enter.  Do not leave empty."
                                                label="DNS servers for this device" multiple />
                                            <v-combobox v-if="selected.vpn.type == 'Service'"
                                                v-model="selected.vpn.current.allowedIPs" chips closable-chips
                                                hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs"
                                                multiple :readonly="!inEdit" />
                                            <v-text-field :readonly="!inEdit" v-model="selected.vpn.current.endpoint"
                                                label="Public endpoint for clients" :rules="[rules.ipport]" />
                                            <v-text-field type="number" v-model="selected.vpn.current.mtu" label="MTU"
                                                hint="Leave at 0 for auto, 1350 for IPv6 or if problems occur" />
                                            <v-text-field type="number"
                                                v-model="selected.vpn.current.persistentKeepalive"
                                                label="Persistent keepalive"
                                                hint="To disable, set to 0.  Recommended value 23 (seconds)" />
                                            <v-switch v-model="selected.vpn.enable" color="#004000" inset
                                                :label="selected.vpn.enable ? 'Enabled' : 'Disabled'"
                                                :readonly="!inEdit" />
                                            <p class="text-caption">Created by {{ selected.vpn.createdBy }} at {{
                                                formatDate(selected.vpn.created) }}<br />
                                                Last update by {{ selected.vpn.updatedBy }} at {{
                                                    formatDate(selected.vpn.updated) }}</p>

                                        </v-col>
                                    </v-row>
                                    <v-expansion-panels v-if="inEdit && !(selected.vpn.type == 'Service')">
                                        <v-expansion-panel>
                                            <v-expansion-panel-title>Advanced Configuration</v-expansion-panel-title>
                                            <v-expansion-panel-text>
                                                <div class="d-flex flex-no-wrap justify-space-between">
                                                    <v-col cols="12">
                                                        <v-combobox v-model="selected.vpn.current.allowedIPs" chips
                                                            closable-chips hint="Write IPv4 or IPv6 CIDR and hit enter"
                                                            label="Allowed IPs" multiple />
                                                        <v-text-field v-model="selected.vpn.id" label="VPN ID"
                                                            readonly />
                                                        <v-text-field v-model="selected.vpn.netid" label="Network ID"
                                                            readonly />
                                                        <v-text-field v-model="selected.vpn.deviceid" label="Device ID"
                                                            disabled />
                                                        <v-combobox v-model="selected.vpn.role"
                                                            :items="['', 'Ingress', 'Egress']" label="Role" />
                                                        <v-text-field v-model="selected.vpn.current.table"
                                                            label="Table" />
                                                        <v-text-field v-model="selected.vpn.current.publicKey"
                                                            label="Public key" />
                                                        <v-text-field v-if="!editPrivate" label="Private key" readonly
                                                            append-inner-icon="mdi-square-edit-outline"
                                                            @click:append-inner="editPrivate = true" />
                                                        <v-text-field v-if="editPrivate"
                                                            v-model="selected.vpn.current.privateKey"
                                                            label="Private key"
                                                            hint="Clear this field to have the client manage its private key" />
                                                        <v-text-field label="Preshared Key" readonly
                                                            append-inner-icon="mdi-content-copy"
                                                            @click:append-inner="copy(selected.vpn.current.presharedKey)" />
                                                        <v-textarea v-model="selected.vpn.current.postUp"
                                                            label="PostUp Script"
                                                            hint="Only applies to linux servers" />
                                                        <v-textarea v-model="selected.vpn.current.postDown"
                                                            label="PostDown Script"
                                                            hint="Only applies to linux servers" />
                                                        <v-switch v-model="selected.vpn.current.subnetRouting"
                                                            color="#004000" inset label="Enable subnet routing" />
                                                        <v-divider></v-divider>
                                                        <div style="display:flex; flex-wrap:wrap; gap:0 16px;">
                                                            <div style="display:flex; flex-direction:column; flex:1 1 180px; min-width:0;">
                                                                <v-switch v-model="selected.vpn.current.syncEndpoint" color="#004000" inset label="Sync Endpoint" />
                                                                <v-switch v-model="selected.vpn.current.upnp" color="#004000" inset label="UPnP" />
                                                                <v-switch v-model="selected.vpn.current.failsafe" color="#004000" inset label="FailSafe" />
                                                            </div>
                                                            <div style="display:flex; flex-direction:column; flex:1 1 180px; min-width:0;">
                                                                <v-switch v-model="selected.vpn.current.hasSSH" color="#004000" inset label="SSH" />
                                                                <v-switch v-model="selected.vpn.current.hasRDP" color="#004000" inset label="Remote Desktop" />
                                                                <v-switch v-model="selected.vpn.current.enableDns" color="#004000" inset label="Nettica DNS" />
                                                            </div>
                                                        </div>
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
                                                        <v-btn class="px-3" color="#400000"
                                                            @click="removeVPN(selected.vpn)">
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
                <v-dialog v-if="device" v-model="dialogCreate" max-width="550" persistent
                    @keydown.esc="dialogCreate = false">
                    <v-card>
                        <v-card-title class="headline">Add New Device</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-text-field v-model="device.name" :rules="[rules.required, rules.host]"
                                            required>
                                            <template #label>
                                                <span class="text-red"><strong>* </strong></span>Host friendly name
                                            </template>
                                        </v-text-field>
                                        <v-select return-object v-model="addNet.selected" :items="addNet.items"
                                            item-title="text" item-value="value"
                                            label="Join this network (optional)" persistent-hint />
                                        <v-select return-object v-model="acntList.selected" :items="acntList.items"
                                            item-title="text" item-value="value" label="For this account"
                                            :rules="[v => !!v || 'Account is required']" persistent-hint />
                                        <v-select return-object v-model="platforms.selected" :items="platforms.items"
                                            item-title="text" item-value="value" label="Platform of this device"
                                            persistent-hint />
                                        <v-text-field v-model="device.instanceid" label="AWS or Azure Instance ID" />
                                        <v-switch v-model="device.enable" color="#004000" inset
                                            :label="device.enable ? 'Enable device after creation' : 'Disable device after creation'" />
                                        <v-switch v-model="use_ezcode" color="#004000" inset
                                            :label="use_ezcode ? 'Use EZ-Code' : 'Do not use EZ-Code'" />
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer />
                            <v-btn :disabled="!valid" color="#004000" @click="create(device)">
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
                <v-dialog v-if="device && device.ezcode" v-model="dialogEZCode" max-width="550" persistent
                    @keydown.esc="dialogEZCode = false">
                    <v-card>
                        <v-card-title class="headline">EZ-Code</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-text-field v-model="device.name" label="Host friendly name"
                                            :readonly="true" />
                                        <v-text-field v-model="device.ezcode" label="EZ-Code" :readonly="true" />
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer />
                            <v-btn color="#004000" @click="dialogEZCode = false">
                                OK
                                <v-icon end>mdi-check-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-dialog>
                <v-dialog v-if="vpn" v-model="dialogAddVPN" max-width="550" persistent
                    @keydown.esc="dialogAddVPN = false">
                    <v-card>
                        <v-card-title class="headline">Add VPN</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-select return-object v-model="netList.selected" :items="netList.items"
                                            @update:model-value="updateDefaults" item-title="text" item-value="value"
                                            label="Join this network" :rules="[v => !!v || 'Network is required']"
                                            persistent-hint required />
                                        <v-text-field v-model="vpn.name" label="DNS name for this device"
                                            :rules="[rules.required, rules.host]" required />
                                        <v-text-field v-model="vpn.current.endpoint" label="Public endpoint for clients"
                                            :rules="[rules.ipport]" />
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
                            <v-btn :disabled="!valid" color="#004000" @click="createVPN(vpn)">
                                Submit
                                <v-icon end>mdi-check-outline</v-icon>
                            </v-btn>
                            <v-btn color="#000040" @click="dialogAddVPN = false">
                                Cancel
                                <v-icon end>mdi-close-circle-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-dialog>
                <v-dialog v-if="vpn" v-model="dialogServiceHost" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Manage Service: {{ vpn.name }}</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-combobox v-model="vpn.current.allowedIPs" chips closable-chips
                                            hint="Enter IPv4 or IPv6 CIDR and press tab" label="Allowed IPs" multiple />
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                    </v-card>
                    <v-card>
                        <v-card-actions>
                            <v-btn :disabled="!valid" color="#004000" @click="updateDevice(selected)">
                                Submit
                                <v-icon end>mdi-check-outline</v-icon>
                            </v-btn>
                            <v-btn color="#000040" @click="dialogServiceHost = false">
                                Cancel
                                <v-icon end>mdi-close-circle-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-dialog>
            </v-col></v-row>
    </v-container>
</template>

<style>
.v-treeview-node {
    padding: 10px 0;
}

.v-treeview-node--leaf {
    padding: 0;
}

.gray {
    color: gray;
}
</style>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useDeviceStore } from '@/stores/device'
import { useNetStore } from '@/stores/net'
import { useAccountStore } from '@/stores/account'
import { useVpnStore } from '@/stores/vpn'
import { useAuthStore } from '@/stores/auth'
import ApiService from '@/services/api.service'
import { formatDate } from '@/utils/formatDate'

const deviceStore = useDeviceStore()
const netStore = useNetStore()
const accountStore = useAccountStore()
const vpnStore = useVpnStore()
const authStore = useAuthStore()

const acntList = ref({})
const showPrivate = ref(false)
const showPreshared = ref(false)
const showApiKey = ref(false)
const showTree = ref(false)
const friendly = ref(false)
const editPrivate = ref(false)
const use_ezcode = ref(true)
const dialogCreate = ref(false)
const dialogAddVPN = ref(false)
const dialogServiceHost = ref(false)
const dialogEZCode = ref(false)
const device = ref(null)
const net = ref(null)
const vpn = ref(null)
const items = ref([])
const active = ref([])
const open = ref([])
const inEdit = ref(false)
const valid = ref(false)
const refreshing = ref(false)
const search = ref('')
const netList = ref({})
const addNet = ref({})
const publicSubnets = ref(false)

const logging = ref({
    selected: { text: 'None', value: '' },
    items: [
        { text: 'None', value: '' },
        { text: 'Errors', value: '#400000' },
        { text: 'Info', value: 'info' },
        { text: 'Debug', value: 'debug' },
    ],
})

const platforms = ref({
    selected: { text: '', value: '' },
    items: [
        { text: 'Windows', value: 'Windows' },
        { text: 'Linux', value: 'Linux' },
        { text: 'MacOS', value: 'macos' },
        { text: 'Apple iOS', value: 'ios' },
        { text: 'Raspberry Pi', value: 'raspberry' },
        { text: 'Android', value: 'android' },
        { text: 'Native WireGuard', value: 'Native' },
    ],
})

const rules = {
    required: value => !!value || 'Required.',
    email: v => /.+@.+\..+/.test(v) || 'E-mail must be valid',
    host: v => /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/.test(v) || 'Only letters, numbers, dots and hyphens are allowed. Must start and end with a letter or number.',
    ipport: v => (!v || v && v.length == 0 || /^(((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5})|^$)(,\s+((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5})|^$))*)$/.test(v)) || 'If present, must be valid IPv4 or IPv6 address and port',
}

const selected = computed(() => {
    if (!active.value.length) return undefined
    const id = active.value[0]
    for (let i = 0; i < items.value.length; i++) {
        if (items.value[i].id === id) return items.value[i]
        for (let j = 0; j < items.value[i].children.length; j++) {
            if (items.value[i].children[j].id === id) return items.value[i].children[j]
        }
    }
    return undefined
})

watch(() => deviceStore.devices, () => {
    buildTree()
    showTree.value = true
    friendly.value = deviceStore.devices.length === 0
})

watch(() => accountStore.accounts, (val) => {
    acntList.value = {
        selected: { text: '', value: '' },
        items: []
    }
    for (let i = 0; i < val.length; i++) {
        acntList.value.items[i] = { text: val[i].accountName + ' - ' + val[i].parent, value: val[i].parent }
    }
})

watch(() => netStore.nets, (val) => {
    netList.value = { selected: { text: '', value: '' }, items: [] }
    for (const net of val) {
        if (device.value?.vpns?.some(v => v.netid === net.id)) continue
        netList.value.items.push({ text: net.netName, value: net.id })
    }
})

function onVisibilityChange() {
    if (document.visibilityState === 'visible') Refresh()
}

onMounted(() => {
    document.addEventListener('visibilitychange', onVisibilityChange)
    if (deviceStore.devices.length > 0) {
        buildTree()
        showTree.value = true
    }
    accountStore.readAll(authStore.user.email)
    deviceStore.readAll()
    netStore.readAll()
})

onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange)
})

function Refresh() {
    accountStore.readAll(authStore.user.email)
    deviceStore.readAll()
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

async function asyncRefresh() {
    await accountStore.readAll(authStore.user.email)
    await deviceStore.readAll()
    buildTree()
}

function filter(value, query, item) {
    const raw = item?.raw ?? item
    if (!query) return true
    if (raw.name && raw.name.toLowerCase().includes(query.toLowerCase())) return true
    if (raw.isDevice && raw.device.tags != null) {
        for (const tag of raw.device.tags) {
            if (tag.toLowerCase().includes(query.toLowerCase())) return true
        }
    }
    if (!raw.isDevice && raw.vpn?.current?.address?.length > 0) {
        for (const addr of raw.vpn.current.address) {
            if (addr.toLowerCase().includes(query.toLowerCase())) return true
        }
    }
    return false
}

function buildTree() {
    items.value = []
    for (let i = 0; i < deviceStore.devices.length; i++) {
        const d = deviceStore.devices[i]
        let logText = 'None'
        if (d.logging === '#400000') logText = 'Errors'
        else if (d.logging === 'info') logText = 'Info'
        else if (d.logging === 'debug') logText = 'Debug'

        const treeItem = {
            id: d.id,
            name: d.name,
            description: d.description,
            device: d,
            status: d.status,
            platform: { text: d.platform, value: d.platform },
            accountid: { text: d.accountid, value: d.accountid },
            logging: { text: logText, value: d.logging },
            icon: 'mdi-devices',
            symbol: 'devices',
            isDevice: true,
            enabled: d.enable,
            children: []
        }

        if (d.os === 'windows') { treeItem.icon = 'mdi-window'; treeItem.symbol = 'window' }
        if (d.os === 'linux') { treeItem.icon = 'mdi-dns'; treeItem.symbol = 'dns' }
        if (d.os === 'macos') { treeItem.icon = 'mdi-laptop'; treeItem.symbol = 'laptop' }
        if (d.os === 'ios') { treeItem.icon = 'mdi-phone'; treeItem.symbol = 'phone_iphone' }
        if (d.os === 'android') { treeItem.icon = 'mdi-phone-android'; treeItem.symbol = 'phone_android' }
        if (d.type === 'Service') { treeItem.icon = 'mdi-cloud'; treeItem.symbol = 'cloud' }

        if (d.vpns) {
            for (let j = 0; j < d.vpns.length; j++) {
                treeItem.children[j] = {
                    id: d.vpns[j].id,
                    name: d.vpns[j].netName,
                    vpn: d.vpns[j],
                    icon: 'mdi-network-outline',
                    symbol: 'network_node',
                    isDevice: false,
                    enabled: d.vpns[j].enable,
                }
            }
        }

        items.value[i] = treeItem
    }

    items.value.sort((a, b) => a.name.toUpperCase() < b.name.toUpperCase() ? -1 : a.name.toUpperCase() > b.name.toUpperCase() ? 1 : 0)
}

function startCreate() {
    device.value = {
        name: '',
        accountid: '',
        email: authStore.user.email,
        enable: true,
        tags: [],
        current: {},
        quiet: true,
        debug: false,
    }

    addNet.value = {
        selected: { text: '', value: '' },
        items: [{ text: '', value: '' }]
    }
    for (let i = 0; i < netStore.nets.length; i++) {
        addNet.value.items[i + 1] = { text: netStore.nets[i].netName, value: netStore.nets[i] }
    }

    acntList.value = { selected: { text: '', value: '' }, items: [] }
    let selected = 0
    for (let i = 0; i < accountStore.accounts.length; i++) {
        acntList.value.items[i] = { text: accountStore.accounts[i].accountName + ' - ' + accountStore.accounts[i].parent, value: accountStore.accounts[i].parent }
        if (acntList.value.items[i].value === device.value.accountid) selected = i
    }
    acntList.value.selected = acntList.value.items[selected]
    dialogCreate.value = true
}

function make_ezcode() {
    const chars = 'ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789'
    let result = 'ez-'
    for (let i = 0; i < 4; i++) result += chars.charAt(Math.floor(Math.random() * chars.length))
    return result
}

async function create(dev) {
    if (use_ezcode.value) dev.ezcode = make_ezcode()
    device.value = dev
    device.value.platform = platforms.value.selected.value
    device.value.logging = ''
    device.value.checkInterval = 10
    device.value.accountid = acntList.value.selected.value
    device.value.name = device.value.name.trim()
    dialogCreate.value = false

    ApiService.post('/device', device.value)
        .then(d => {
            if (use_ezcode.value) dialogEZCode.value = true
            if (addNet.value.selected.value === '') {
                Refresh()
                deviceStore.error = `Device ${device.value.name} created`
            } else {
                const newVpn = {
                    name: device.value.name + '.' + addNet.value.selected.text,
                    netName: addNet.value.selected.text,
                    netid: addNet.value.selected.value.id,
                    deviceid: d.id,
                    accountid: addNet.value.selected.value.accountid,
                    enable: true,
                    current: {},
                }
                ApiService.post('/vpn', newVpn)
                    .then(v => {
                        deviceStore.error = `Device ${device.value.name} created and added to ${v.netName}`
                        Refresh()
                    })
                    .catch(error => deviceStore.error = error)
            }
        })
        .catch(error => deviceStore.error = error)
}

function startAddVPN(dev) {
    device.value = dev
    netStore.readAll()
    vpn.value = {
        name: '',
        deviceid: dev.id,
        email: authStore.user.email,
        enable: true,
        tags: [],
        current: {},
    }

    netList.value = { selected: { text: '', value: '' }, items: [] }
    for (const net of netStore.nets) {
        if (dev.vpns?.some(v => v.netid === net.id)) continue
        netList.value.items.push({ text: net.netName, value: net.id })
    }
    dialogAddVPN.value = true
}

function updateDefaults(net) {
    let sel = 0
    for (let i = 0; i < netStore.nets.length; i++) {
        if (netStore.nets[i].id === net.value) { sel = i; break }
    }
    vpn.value.name = device.value.name + '.' + netStore.nets[sel].netName
    vpn.value.current.syncEndpoint = netStore.nets[sel].default.syncEndpoint
    vpn.value.current.hasSSH = netStore.nets[sel].default.hasSSH
    vpn.value.current.hasRDP = netStore.nets[sel].default.hasRDP
    vpn.value.current.upnp = netStore.nets[sel].default.upnp
    vpn.value.current.failsafe = netStore.nets[sel].default.failsafe
    vpn.value.current.enableDns = netStore.nets[sel].default.enableDns
}

async function createVPN(v) {
    vpn.value = v
    vpn.value.current.listenPort = 0
    if (vpn.value.current.endpoint && vpn.value.current.endpoint.includes(':')) {
        const parts = vpn.value.current.endpoint.split(':')
        vpn.value.current.listenPort = parseInt(parts[parts.length - 1], 10)
    }
    vpn.value.netName = netList.value.selected.text
    vpn.value.netid = netList.value.selected.value
    dialogAddVPN.value = false
    await vpnStore.create(vpn.value)
    await new Promise(r => setTimeout(r, 2000))
    await asyncRefresh()
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

function removeDevice(dev) {
    if (confirm(`Do you really want to delete ${dev.name} ?`)) {
        deviceStore.delete(dev)
    }
}

async function removeVPN(v) {
    if (confirm(`Do you really want to delete ${v.name} from ${v.netName}?`)) {
        await vpnStore.delete(v)
        Refreshing()
    }
}

function updateEnable(dev) {
    deviceStore.update(dev)
}


function copyDeviceConfig(dev) {
    const url = `curl "http://localhost:53280/config/?id=${dev.id}&apiKey=${dev.apiKey}&server=${dev.server}"`
    navigator.clipboard.writeText(url).then(() => deviceStore.error = 'Copied to clipboard')
}

async function updateDevice(item) {
    device.value = item.device
    device.value.logging = item.logging.value
    device.value.platform = item.platform.value
    device.value.accountid = item.accountid.value
    device.value.checkInterval = parseInt(device.value.checkInterval, 10)
    dialogServiceHost.value = false
    inEdit.value = false
    deviceStore.update(device.value)
}

function copy(text) {
    navigator.clipboard.writeText(text).then(() => deviceStore.error = 'Copied to clipboard')
}

async function forceFileDownload(v) {
    await vpnStore.readConfig(v)
    await new Promise(r => setTimeout(r, 1000))
    const config = vpnStore.getVPNConfig(v.id)
    if (!config) {
        deviceStore.error = 'Failed to download device config'
        return
    }
    const url = window.URL.createObjectURL(new Blob([config]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', v.name.split(' ').join('-') + '-' + v.netName.split(' ').join('-') + '.zip')
    document.body.appendChild(link)
    link.click()
}
</script>
