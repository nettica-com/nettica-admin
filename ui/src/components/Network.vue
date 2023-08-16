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
                Networks
                <v-spacer></v-spacer>
                <v-text-field v-if="listView" v-model="search" append-icon="mdi-magnify" label="Search" single-line
                    hide-details></v-text-field>
                <v-spacer></v-spacer>
                <v-btn color="success" @click="startCreate">
                    Create New Network
                    <span class="material-symbols-outlined">hub</span>
                </v-btn>
            </v-card-title>
            <d3-network class="network" :net-nodes="nodes" :net-links="links" :options="options" />
            <v-data-table v-if="listView" :headers="headers" :items="nets" :search="search"
                no-data-text="Welcome to Nettica!  Click 'Create New Network' above to get started."
                no-results-text="No results matching your search" @click:row="loadNetwork">
                <template v-slot:item.default.address="{ item }">
                    <v-chip v-for="(ip, i) in item.default.address" :key="i" color="#336699" text-color="white">
                        <v-icon left>mdi-ip-network</v-icon>
                        {{ ip }}
                    </v-chip>
                </template>
                <template v-slot:item.tags="{ item }">
                    <v-chip v-for="(tag, i) in item.tags" :key="i" color="blue-grey" text-color="white">
                        <v-icon left>mdi-tag</v-icon>
                        {{ tag }}
                    </v-chip>
                </template>
                <template v-slot:item.created="{ item }">
                    <v-row>
                        <p>{{ item.createdBy }} at {{ item.created | formatDate }}</p>
                    </v-row>
                </template>
                <template v-slot:item.updated="{ item }">
                    <v-row>
                        <p>At {{ item.updated | formatDate }} by {{ item.updatedBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.action="{ item }">
                    <v-row>
                        <v-icon class="pr-1 pl-1" @click.stop="startUpdate(item)" title="Edit">
                            mdi-square-edit-outline
                        </v-icon>
                        <v-icon class="pr-1 pl-1" @click="remove(item)" title="Delete">
                            mdi-trash-can-outline
                        </v-icon>
                    </v-row>
                </template>

            </v-data-table>
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


                                <v-combobox v-model="net.default.address" :items="net.default.address"
                                    label="IP subnet for this network (ex. 10.10.10.0/24)"
                                    :rules="[v => !!v || 'Subnet is required',]" multiple chips persistent-hint
                                    required />
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
                                <v-combobox v-model="net.tags" chips hint="Enter a tag, hit tab, hit enter." label="Tags"
                                    multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.tags.splice(net.tags.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
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
        <v-dialog v-if="net" v-model="dialogUpdate" max-width="550">
            <v-card>
                <v-card-title class="headline">Edit Net</v-card-title>
                <v-card-text>

                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-text-field v-model="net.id" label="Id" :rules="[v => !!v || 'Net id is required',]"
                                    required />
                                <v-text-field v-model="net.netName" label="Friendly name"
                                    :rules="[v => !!v || 'Net name is required',]" required />
                                <v-text-field v-model="net.description" label="Description" />
                                <v-combobox v-model="net.default.address" chips hint="Write IPv4 or IPv6 CIDR and hit enter"
                                    label="Addresses" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.default.address.splice(net.default.address.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox v-model="net.default.dns" chips
                                    hint="Write IP address(es) and hit enter or leave empty.  If not empty, be sure to include your local resolver."
                                    label="DNS servers for this net" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.default.dns.splice(net.default.dns.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox v-model="net.tags" chips hint="Write tag name and hit enter" label="Tags"
                                    multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.tags.splice(net.tags.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
            <v-expansion-panels>
                <v-expansion-panel>
                    <v-expansion-panel-header dark>Advanced configuration</v-expansion-panel-header>
                    <v-expansion-panel-content>
                        <div class="d-flex flex-no-wrap justify-space-between">
                            <v-col cols="12">
                                <v-text-field v-model="net.default.presharedKey" label="Preshared Key" />
                                <v-combobox v-model="net.default.allowedIPs" chips
                                    hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple dark>
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                            @click:close="net.default.allowedIPs.splice(net.default.allowedIPs.indexOf(item), 1)">
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>

                                <v-text-field type="number" v-model="net.default.mtu" label="Define default global MTU"
                                    hint="Leave at 0 and let us take care of MTU" />
                                <v-text-field type="number" v-model="net.default.persistentKeepalive"
                                    label="Persistent keepalive"
                                    hint="To disable, set to 0.  Recommended value 29 (seconds)" />
                                <v-text-field v-model="net.default.listenPort" type="number" :rules="[
                                    v => !!v || 'Listen port is required',
                                ]" label="Listen port" required />
                                <v-switch v-model="net.default.upnp" color="success" inset
                                    label="Enable UPnP where possible" />
                                <v-switch v-model="net.default.enableDns" color="success" inset
                                    label="Enable Nettica DNS" />

                            </v-col>
                        </div>
                    </v-expansion-panel-content>
                </v-expansion-panel>
            </v-expansion-panels>
            <v-card>
                <v-card-actions>
                    <v-btn :disabled="!valid" color="success" @click="update(net)">
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="primary" @click="dialogUpdate = false">
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
        listView: true,
        dialogCreate: false,
        dialogUpdate: false,
        noEdit: false,
        net: null,
        panel: 1,
        valid: false,
        search: '',
        headers: [
            { text: 'Name', value: 'netName', },
            { text: 'Description', value: 'description' },
            { text: 'Subnet', value: 'default.address', },
            { text: 'Created', value: 'created', sortable: false, },
            { text: 'Tags', value: 'tags', },
            { text: 'Actions', value: 'action', sortable: false, },

        ],
        nodes: [
        ],
        links: [
        ],
        nodeSize: 50,
        canvas: false,


    }),

    computed: {
        ...mapGetters({
            user: 'auth/user',
            server: 'server/server',
            nets: 'net/nets',
            hosts: 'host/hosts',
            accounts: 'account/accounts',

        }),
        options() {
            return {
                force: 4000,
                size: { w: 1000, h: 500 },
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

    methods: {
        ...mapActions('vpn', {
            readAllVPNs: 'readAll',
        }),
        ...mapActions('net', {
            errorNet: 'error',
            readAllNetworks: 'readAll',
            createNet: 'create',
            updateNet: 'update',
            deleteNet: 'delete',
        }),
        ...mapActions('server', {
            readServer: 'read',
        }),
        ...mapActions('account', {
            readAllAccounts: 'readAll',
        }),

        Refresh() {
            this.readAllAccounts(this.user.email)
            this.readAllVPNs()
            this.readAllNetworks()
        },

        loadNetwork(net) {
            let name = net.netName
            let x = 0
            let l = 0
            this.links = []
            this.nodes = []
            let net_hosts = []
            for (let i = 0; i < this.hosts.length; i++) {
                if (this.hosts[i].netName == name) {
                    net_hosts[x] = this.hosts[i]
                    this.nodes[x] = { id: x, name: this.hosts[i].name, /* _color:'gray'*/ }
                    if (this.hosts[i].current.endpoint == "") {
                        this.nodes[x]._color = "#34adcd"
                    } else {
                        this.nodes[x]._color = "#83c44d"
                    }
                    if (this.hosts[i].role == "Egress") {
                        this.nodes[x]._color = "#50C878"
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
                net.id = net.id
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

            this.net.default.listenPort = parseInt(this.net.default.listenPort, 10);
            this.net.default.persistentKeepalive = parseInt(this.net.default.persistentKeepalive, 10);
            this.net.default.mtu = parseInt(this.net.default.mtu, 10);
            this.net.id = net.id
            this.net.netName = net.netName


            // check allowed IPs
            if (net.default.allowedIPs.length < 0) {
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

        forceFileDownload(net) {
            let config = this.getNetConfig(net.id)
            if (!config) {
                this.errorNet('Failed to download net config');
                return
            }
            const url = window.URL.createObjectURL(new Blob([config]))
            const link = document.createElement('a')
            link.href = url
            link.setAttribute('download', net.netName.split(' ').join('-') + '.conf') //or any other extension
            document.body.appendChild(link)
            link.click()
        },
    }
};
</script>
