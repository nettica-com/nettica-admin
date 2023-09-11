<template>
    <v-container style="padding-top:0px">
        <div>
            <v-btn class="mb-3 mt-0" @click="Refresh()">
                <v-icon dark>mdi-refresh</v-icon>
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
                        <v-text-field v-model="search" append-icon="mdi-magnify" label="Search" single-line
                            hide-details></v-text-field>
                    </v-col>
                    <v-col cols="4" class="text-right">
                        <v-btn color="success" @click="startCreate">
                            Create
                            <span class="material-symbols-outlined">hub</span>
                        </v-btn>
                    </v-col>
                </v-row>
            </v-card-title>
            <d3-network class="network" :net-nodes="nodes" :net-links="links" :options="options" />
            <v-divider></v-divider>
            <v-row style="padding-top: 12px;">
                <v-col cols="6">
                    <v-treeview v-if="showTree" :items="items" :search="search" :active.sync="active" :open.sync="open"
                        activatable hoverable @update:active="loadNetwork">
                        <template v-slot:prepend="{ item }">
                            <span v-if="item.symbol" class="material-symbols-outlined">{{ item.symbol }}</span>
                            <v-icon v-else>
                                {{ item.icon }}
                            </v-icon>
                        </template>
                        <template v-slot:append="{ item }">
                            <v-spacer></v-spacer>
                            <span v-if="!item.isNet" class="hidden-xs-only" >{{  item.address }}</span>
                        </template>
                    </v-treeview>
                </v-col>
                <v-divider vertical></v-divider>
                <v-col cols="6" class="text-center">
                    <div v-if="!selected" class="text-h6 grey--text text--lighten-1 font-weight-light"
                        style="align-self: center;">

                    </div>
                    <v-card v-else-if="selected.isNet" :key="selected.id" class="px-3 mx-auto" flat>
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
                                <v-combobox v-model="selected.net.tags" chips hint="Enter a tag, hit tab, hit enter."
                                    label="Tags" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="selected.net.tags.splice(selected.net.tags.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox v-model="selected.net.default.address" chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Addresses" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="selected.net.default.address.splice(selected.net.default.address.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox v-model="selected.net.default.allowedIPs" chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="selected.net.default.allowedIPs.splice(selected.net.default.allowedIPs.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox v-model="selected.net.default.dns" chips
                                    hint="Enter IP address(es) and hit enter or leave empty."
                                    label="DNS servers for this network" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select }">

                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="selected.net.default.dns.splice(selected.net.default.dns.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-text-field v-model="selected.net.id" label="Network ID" readonly />
                                <v-text-field v-model="selected.net.default.presharedKey" label="Preshared Key" autocomplete="off"
                                                        :append-icon="showPreshared ? 'mdi-eye' : 'mdi-eye-off'"
                                                        :type="showPreshared ? 'text' : 'password'"
                                                        @click:append="showPreshared = !showPreshared" />

                                <v-text-field type="number" v-model="selected.net.default.mtu"
                                    label="Define default global MTU" hint="Leave at 0 and let us take care of MTU" />
                                <v-text-field type="number" v-model="selected.net.default.persistentKeepalive"
                                    label="Persistent keepalive"
                                    hint="To disable, set to 0.  Recommended value 29 (seconds)" />
                                <v-text-field v-model="selected.net.default.listenPort" type="number" :rules="[
                                    v => !!v || 'Listen port is required',
                                ]" label="Listen port" required />
                                <v-switch v-model="selected.net.policies.userEndpoints" color="success" inset
                                    label="Users can create Endpoints" />
                                <v-switch v-model="selected.net.policies.onlyEndpoints" color="success" inset
                                    label="Clients cannot see other clients" />
                                    <v-switch v-model="selected.net.default.upnp" color="success" inset
                                    label="Enable UPnP where possible" />
                                <v-switch v-model="selected.net.default.enableDns" color="success" inset
                                    label="Enable Nettica DNS" />


                                <p class="text-caption">Created by {{ selected.net.createdBy }} at {{
                                    selected.net.created | formatDate }}<br />
                                    Last update by {{ selected.net.updatedBy }} at {{ selected.net.updated |
                                        formatDate }}</p>
                            </v-col>
                        </v-row>
                        <v-card-actions>
                            <v-container>
                                <v-row>
                                    <v-col>
                                        <v-btn color="success" @click="update(selected.net)">
                                            Save
                                            <v-icon right dark>mdi-check-outline</v-icon>
                                        </v-btn>
                                    </v-col>
                                    <v-col>
                                        <v-btn color="error" @click="remove(selected.net)">
                                            Delete
                                            <v-icon right dark>mdi-delete-outline</v-icon>
                                        </v-btn>
                                    </v-col>
                                </v-row>
                            </v-container>
                        </v-card-actions>
                    </v-card>
                    <v-card v-else-if="!selected.isNet">
                        <v-card-text width="600" class="px-3">
                            <v-icon size="50" class="material-symbols-outlined">network_node</v-icon>
                            <h3 class="text-h5 mb-2">
                                {{ selected.name }}
                            </h3>
                        </v-card-text>
                        <v-divider></v-divider>
                        <v-row class="px-3" width="600">
                            <v-col flex>
                                <v-text-field v-model="selected.vpn.name" label="DNS name" :readonly="!inEdit" />
                                <v-combobox :readonly="!inEdit" v-model="selected.vpn.current.address" chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Addresses" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="selected.vpn.current.address.splice(selected.vpn.current.address.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox :readonly="!inEdit" v-model="selected.vpn.current.dns" chips
                                    hint="Enter IP address(es) and hit enter or leave empty."
                                    label="DNS servers for this device" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select }">

                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="selected.vpn.current.dns.splice(selected.vpn.current.dns.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-text-field :readonly="!inEdit" v-model="selected.vpn.current.endpoint"
                                    label="Public endpoint for clients" />
                                <v-text-field :readonly="!inEdit" v-model="selected.vpn.current.listenPort" type="number"
                                    label="Listen port" />
                                <v-switch v-model="selected.vpn.enable" color="success" inset
                                    :label="selected.vpn.enable ? 'Enabled' : 'Disabled'" :readonly="!inEdit" />
                                <p class="text-caption">Created by {{ selected.vpn.createdBy }} at {{ selected.vpn.created |
                                    formatDate }}<br />
                                    Last update by {{ selected.vpn.updatedBy }} at {{ selected.vpn.updated | formatDate }}</p>

                            </v-col>
                        </v-row>
                        <v-expansion-panels v-if="inEdit">
                            <v-expansion-panel>
                                <v-expansion-panel-header dark>Advanced Configuration</v-expansion-panel-header>
                                <v-expansion-panel-content>
                                    <div class="d-flex flex-no-wrap justify-space-between">
                                        <v-col cols="12">
                                            <v-combobox v-model="selected.vpn.current.allowedIPs" chips
                                                hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple
                                                dark>

                                                <template v-slot:selection="{ attrs, item, select }">
                                                    <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                        @click:close="selected.vpn.current.allowedIPs.splice(selected.vpn.current.allowedIPs.indexOf(item), 1)">
                                                        <strong>{{ item }}</strong>&nbsp;
                                                    </v-chip>
                                                </template>
                                            </v-combobox>
                                            <v-switch v-model="publicSubnets" color="success" inset
                                                label="Route all public traffic through tunnel" />

                                            <v-text-field v-model="selected.vpn.accountid" label="Account ID" readonly />
                                            <v-text-field v-model="selected.vpn.id" label="VPN ID" readonly />
                                            <v-text-field v-model="selected.vpn.netid" label="Network ID" readonly />
                                            <v-text-field v-model="selected.vpn.deviceid" label="Device ID" readonly />
                                            <v-text-field v-model="selected.vpn.current.table" label="Table" />
                                            <v-text-field v-model="selected.vpn.current.publicKey" label="Public key" />
                                            <v-text-field v-model="selected.vpn.current.privateKey" label="Private key"
                                                autocomplete="off" :append-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                                                :type="showPrivate ? 'text' : 'password'"
                                                hint="Clear this field to have the client manage its private key"
                                                @click:append="showPrivate = !showPrivate" />
                                            <v-text-field v-model="selected.vpn.current.presharedKey" label="Preshared Key"
                                                autocomplete="off" :append-icon="showPreshared ? 'mdi-eye' : 'mdi-eye-off'"
                                                :type="showPreshared ? 'text' : 'password'"
                                                @click:append="showPreshared = !showPreshared" />
                                            <v-text-field type="number" v-model="selected.vpn.current.mtu"
                                                label="Define global MTU" hint="Leave at 0 and let us take care of MTU" />
                                            <v-text-field type="number" v-model="selected.vpn.current.persistentKeepalive"
                                                label="Persistent keepalive"
                                                hint="To disable, set to 0.  Recommended value 29 (seconds)" />
                                            <v-textarea v-model="selected.vpn.current.postUp" label="PostUp Script"
                                                hint="Only applies to linux servers" />
                                            <v-textarea v-model="selected.vpn.current.postDown" label="PostDown Script"
                                                hint="Only applies to linux servers" />
                                            <v-switch v-model="selected.vpn.current.subnetRouting" color="success" inset
                                                label="Enable subnet routing" />
                                            <v-switch v-model="selected.vpn.current.upnp" color="success" inset
                                                label="Enable UPnP" />
                                            <v-switch v-model="selected.vpn.current.enableDns" color="success" inset
                                                label="Enable Nettica DNS" />

                                        </v-col>
                                    </div>
                                </v-expansion-panel-content>
                            </v-expansion-panel>
                        </v-expansion-panels>


                        <v-card>
                            <v-card-actions v-if="inEdit">
                                <v-btn color="success" @click="updateVPN(selected.vpn)">
                                    Submit
                                    <v-icon right dark>mdi-check-outline</v-icon>
                                </v-btn>
                                <v-btn color="primary" @click="inEdit = false">
                                    Cancel
                                    <v-icon right dark>mdi-close-circle-outline</v-icon>
                                </v-btn>
                            </v-card-actions>
                            <v-card-actions v-else>
                                <v-container>
                                    <v-row>
                                        <v-col>
                                            <v-btn color="success" @click="forceFileDownload(selected.vpn)">
                                                Download
                                                <v-icon right dark>mdi-cloud-download-outline</v-icon>
                                            </v-btn>
                                        </v-col>
                                        <v-col>
                                            <v-btn class="px-3" color="primary" @click="inEdit = true">
                                                Edit
                                                <v-icon right dark>mdi-pencil-outline</v-icon>
                                            </v-btn>
                                        </v-col>
                                        <v-col>
                                            <v-btn class="px-3" color="error" @click="removeVPN(selected.vpn)">
                                                Delete
                                                <v-icon right dark>mdi-delete-outline</v-icon>
                                            </v-btn>
                                        </v-col>
                                    </v-row>
                                </v-container>
                            </v-card-actions>
                        </v-card>

                    </v-card>
                </v-col>
            </v-row>
        </v-card>
        <v-dialog v-if="net" v-model="dialogCreate" max-width="550">
            <v-card>
                <v-card-title class="headline">Create New Network</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-text-field v-model="net.netName" label="Network friendly name"
                                    :rules="[v => !!v || 'Network name is required',]" required />
                                <v-text-field v-model="net.description" label="Description" />
                                <v-select return-object v-model="acntList.selected" :items="acntList.items" item-text="text"
                                    item-value="value" label="For this account"
                                    :rules="[v => !!v || 'Account is required',]" single persistent-hint />
                                <v-combobox v-model="net.tags" chips hint="Enter a tag, hit tab, hit enter." label="Tags"
                                    multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.tags.splice(net.tags.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox v-model="net.default.address" :items="net.default.address"
                                    label="IP subnet for this network (ex. 10.10.10.0/24)"
                                    :rules="[v => !!v || 'Subnet is required',]" multiple chips persistent-hint required />
                                <v-combobox v-model="net.default.dns" chips
                                    hint="Enter the IP address of a global DNS resolver, hit tab, hit enter."
                                    label="DNS servers for this network" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.default.dns.splice(net.default.dns.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-switch v-model="net.policies.userEndpoints" color="success" inset
                                    label="Users can create Endpoints" />
                                <v-switch v-model="net.policies.onlyEndpoints" color="success" inset
                                    label="Clients cannot see other clients" />
                                <v-switch v-model="net.default.upnp" color="success" inset
                                    label="Enable UPnP where possible" />
                                <v-switch v-model="net.default.enableDns" color="success" inset
                                    label="Enable Nettica DNS" />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="success" @click="create(net)">
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="primary" @click="dialogCreate = false">
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-container>
</template>

<!-- <style src="vue-d3-network/dist/vue-d3-network.css"></style> -->
<style>
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
<script>


var D3Network = window['vue-d3-network']


import { mapActions, mapGetters } from 'vuex'

export default {
    name: 'Networks',

    data: () => ({
        acntList: {},
        showTree: false,
        showPrivate: false,
        showPreshared: false,
        items: [],
        active: [],
        open: [],
        inEdit: false,
        dialogCreate: false,
        publicSubnets: false,
        noEdit: false,
        net: null,
        panel: 1,
        valid: false,
        search: '',
        nodes: [
        ],
        links: [
        ],
        nodeSize: 50,
        canvas: false,


    }),

    computed: {
        selected() {
            if (!this.active.length) return undefined

            const id = this.active[0]
            console.log("selected id = ", id)

            // find the item in the tree
            for (let i = 0; i < this.items.length; i++) {
                if (this.items[i].id == id) {
                    return this.items[i]
                }
                for (let j = 0; j < this.items[i].children.length; j++) {
                    if (this.items[i].children[j].id == id) {
                        return this.items[i].children[j]
                    }
                }
            }

            var vpn = this.vpns.find(vpn => vpn.id === id)
            if (vpn) {
                return vpn
            }


            return this.items.find(item => item.id === id)
        },

        ...mapGetters({
            user: 'auth/user',
            server: 'server/server',
            nets: 'net/nets',
            vpns: 'vpn/vpns',
            hosts: 'host/hosts',
            accounts: 'account/accounts',
            getvpnconfig: "vpn/getVPNConfig",

        }),
        options() {
            return {
                force: 4000,
                size: { w: 400, h: 300 },
                nodeSize: this.nodeSize,
                nodeLabels: true,
                linkLabels: true,
                canvas: this.canvas
            }
        }

    },

    mounted() {
        this.readAllAccounts(this.user.email)
        this.readAllNetworks()
        this.readAllVPNs()
    },

    watch: {
        nets: function (val) {
            console.log("buildTree = ", this.buildTree())
            this.showTree = true
        },
        vpns: function (val) {
            console.log("buildTree = ", this.buildTree())
            this.showTree = true
        },
        accounts: function(val) {
            this.acntList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            for (let i = 0; i < this.accounts.length; i++) {
                this.acntList.items[i] = { "text": this.accounts[i].accountName + " - " + this.accounts[i].parent, "value": this.accounts[i].parent }
            }
        }
    },


    methods: {
        ...mapActions('vpn', {
            readAllVPNs: 'readAll',
            updatevpn: "update",
            deletevpn: "delete",
            readvpnconfig: "readConfig",

        }),
        ...mapActions('net', {
            errorNet: 'error',
            readAllNetworks: 'readAll',
            createNet: 'create',
            updateNet: 'update',
            deleteNet: 'delete',
        }),
        ...mapActions('account', {
            readAllAccounts: 'readAll',
        }),

        Refresh() {
            this.readAllAccounts(this.user.email)
            this.readAllVPNs()
            this.readAllNetworks()
        },

        buildTree() {
            // build the treeview using the networks
            this.items = []
            var k = 0
            for (let i = 0; i < this.nets.length; i++) {
                this.items[i] = {
                    id: this.nets[i].id,
                    name: this.nets[i].netName,
                    net: this.nets[i],
                    icon: "mdi-network-outline",
                    symbol: "hub",
                    isNet: true,
                    children: []
                }
            }

            this.items.sort((a, b) => {
                const nameA = a.name.toUpperCase(); // ignore upper and lowercase
                const nameB = b.name.toUpperCase(); // ignore upper and lowercase
                if (nameA < nameB) {
                    return -1;
                }
                if (nameA > nameB) {
                    return 1;
                }

                // names must be equal
                return 0;
            });

            k = 0
            console.log("this.items = ", this.items)
            console.log("this.vpns = ", this.vpns)
            for (let i = 0; i < this.items.length; i++) {
                for (let j = 0; j < this.vpns.length; j++) {
                    if (this.vpns[j].netName == this.items[i].name) {
                        this.items[i].children[k] = {
                            id: this.vpns[j].id,
                            name: this.vpns[j].name,
                            address: this.vpns[j].current.address[0],
                            net: this.items[i].net,
                            vpn: this.vpns[j],
                            icon: "mdi-network-outline",
                            symbol: "network_node",
                            isNet: false,
                            isNode: true,
                            children: []
                        }
                        k++
                    }
                }
                k = 0
            }

            return this.items

        },


        loadNetwork(id) {
            let item = this.selected
            if (!item) return
            let net = item.net
            console.log("net = ", net)
            let name = net.netName
            let x = 0
            let l = 0
            this.links = []
            this.nodes = []
            let net_hosts = []
            console.log("this.vpns = ", this.vpns)
            for (let i = 0; i < this.vpns.length; i++) {
                if (this.vpns[i].netName == name) {
                    net_hosts[x] = this.vpns[i]
                        this.nodes[x] = { id: x, name: this.vpns[i].name, /* _color:'gray'*/ }
                    if (this.vpns[i].current.endpoint == "") {
                        this.nodes[x]._color = "#34adcd"
                    } else {
                        this.nodes[x]._color = "#83c44d"
                    }
                    if (this.vpns[i].role == "Egress") {
                        this.nodes[x]._color = "#50C878"
                    }
                    if (this.vpns[i].id == id) {
                        this.nodes[x]._color = "#FF8C00"
                    }                    
                    x++
                }
            }
            for (let i = 0; i < net_hosts.length; i++) {
                for (let j = 0; j < net_hosts.length; j++) {
                    if (i != j && net_hosts[j].current.endpoint != "" && net_hosts[j].role != "Egress") {
                        this.links[l] = { sid: i, tid: j, _color: "white" }
                        l++
                    }
                }
            }
        },

        doCopy() {
            this.$copyText(this.net.default.presharedKey).then(function (e) {
                alert('Copied')
                console.log(e)
            }, function (e) {
                alert('Can not copy')
                console.log(e)
            })
        },

        startCreate() {
            this.net = {
                name: "",
                email: this.user.email,
                enable: true,
                netName: "",
                id: "",
                tags: [],
                accountid: ""

            }
            this.net.default = {
                allowedIPs: [],
                address: [],
                enableDns: false,
                upnp: false,
            }
            this.net.policies = {
                userEndpoints: false,
                onlyEndpoints: false,
            }
            this.acntList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = 0;
            for (let i = 0; i < this.accounts.length; i++) {
                this.acntList.items[i] = { "text": this.accounts[i].accountName + " - " + this.accounts[i].parent, "value": this.accounts[i].parent }
                if (this.acntList.items[i].value == this.net.accountid) {
                    selected = i
                }
            }

            this.acntList.selected = this.acntList.items[selected];

            this.dialogCreate = true;
        },

        create(net) {
            this.net = net
            if (net.default.allowedIPs.length < 0) {
                this.errorNet('Please provide at least one valid CIDR address for net allowed IPs')
                return;
            }
            for (let i = 0; i < net.default.allowedIPs.length; i++) {
                if (this.$isCidr(net.default.allowedIPs[i]) === 0) {
                    this.errorNet('Invalid CIDR detected, please correct before submitting')
                    return
                }
            }
            this.net.accountid = this.acntList.selected.value
            this.dialogCreate = false;
            this.createNet(net)
        },

        remove(net) {
            this.noEdit = true
            if (confirm(`Do you really want to delete ${net.netName}?`)) {
                this.deleteNet(net)
            }
        },

        email(net) {
            if (!net.email) {
                this.errorNet('Net email is not defined')
                return
            }

            if (confirm(`Do you really want to send email to ${net.email} with all configurations ?`)) {
                this.emailNet(net)
            }
        },

        startUpdate(net) {
            if (this.noEdit == true) {
                this.noEdit = false;
                return
            }

            this.net = net;
            this.dialogUpdate = true;
        },

        update(net) {
            this.net = net

            this.net.default.listenPort = parseInt(this.net.default.listenPort, 10);
            this.net.default.persistentKeepalive = parseInt(this.net.default.persistentKeepalive, 10);
            this.net.default.mtu = parseInt(this.net.default.mtu, 10);
            this.net.id = net.id
            this.net.netName = net.netName


            // check allowed IPs
            if (net.default.allowedIPs.length < 1) {
                this.errorNet('Please provide at least one valid CIDR address for net allowed IPs');
                return;
            }
            for (let i = 0; i < net.default.allowedIPs.length; i++) {
                if (this.$isCidr(net.default.allowedIPs[i]) === 0) {
                    this.errorNet('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }
            // check address
            if (net.default.address.length < 1) {
                this.errorNet('Please provide at least one valid CIDR address for net');
                return;
            }
            for (let i = 0; i < net.default.address.length; i++) {
                if (this.$isCidr(net.default.address[i]) === 0) {
                    this.errorNet('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }
            // all good, submit
            this.dialogUpdate = false;
            this.updateNet(net)

        },
        updateVPN(vpn) {
            this.vpn = vpn
            this.vpn.current.listenPort = parseInt(this.vpn.current.listenPort, 10);
            // append the port to the endpoint if it is not there
            if (this.vpn.current.endpoint != null && this.vpn.current.endpoint != "" && this.vpn.current.endpoint.indexOf(":") == -1) {
                if (this.vpn.current.listenPort == 0) {
                    this.vpn.current.listenPort = 51820
                }
                this.vpn.current.endpoint = this.vpn.current.endpoint + ":" + this.vpn.current.listenPort.toString()
            }

            this.vpn.current.persistentKeepalive = parseInt(this.vpn.current.persistentKeepalive, 10);
            this.vpn.current.mtu = parseInt(this.vpn.current.mtu, 10);

            if (this.publicSubnets) {
                this.vpn.current.allowedIPs.push("0.0.0.0/5", "8.0.0.0/7",
                    "11.0.0.0/8", "12.0.0.0/6", "16.0.0.0/4", "32.0.0.0/3", "64.0.0.0/3", "96.0.0.0/6",
                    "101.0.0.0/8", "102.0.0.0/7", "104.0.0.0/5", "112.0.0.0/5", "120.0.0.0/6",
                    "124.0.0.0/7", "126.0.0.0/8", "128.0.0.0/3", "160.0.0.0/5", "168.0.0.0/6",
                    "172.0.0.0/12", "172.32.0.0/11", "172.64.0.0/10",
                    "172.128.0.0/9", "173.0.0.0/8", "174.0.0.0/7", "176.0.0.0/4", "192.0.0.0/9", "192.128.0.0/11", "192.160.0.0/13", "192.169.0.0/16",
                    "192.170.0.0/15", "192.172.0.0/14", "192.176.0.0/12", "192.192.0.0/10", "193.0.0.0/8", "194.0.0.0/7", "196.0.0.0/6", "200.0.0.0/5", "208.0.0.0/4")
            }

            // check allowed IPs
            if (this.vpn.current.allowedIPs.length < 1) {
                this.errordevice('Please provide at least one valid CIDR address for device allowed IPs');
                return
            }
            for (let i = 0; i < this.vpn.current.allowedIPs.length; i++) {
                if (this.$isCidr(this.vpn.current.allowedIPs[i]) === 0) {
                    this.errordevice('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }
            // check address
            if (this.vpn.current.address.length < 1) {
                this.errordevice('Please provide at least one valid CIDR address for device');
                return;
            }
            for (let i = 0; i < this.vpn.current.address.length; i++) {
                if (this.$isCidr(this.vpn.current.address[i]) === 0) {
                    this.errordevice('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }

            this.inEdit = false;
            this.updatevpn(this.vpn)
        },

        async removeVPN(vpn) {
            if (confirm(`Do you really want to delete ${vpn.name} from ${vpn.netName}?`)) {
                await this.deletevpn(vpn)
                // refresh the page
                this.Refresh()
            }
        },

        async forceFileDownload(vpn) {
            console.log("vpn = ", vpn)
            await this.readvpnconfig(vpn)
            // sleep for one second
            await new Promise(r => setTimeout(r, 1000));
            let config = this.getvpnconfig(vpn.id)
            if (!config) {
                console.log("failed to get config")
                this.errordevice('Failed to download device config');
                return
            }
            console.log('config', config)

            const url = window.URL.createObjectURL(new Blob([config]))
            const link = document.createElement('a')
            link.href = url
            link.setAttribute('download', vpn.name.split(' ').join('-') + '-' + vpn.netName.split(' ').join('-') + '.zip') //or any other extension
            document.body.appendChild(link)
            link.click()
        },

    }
};
</script>
