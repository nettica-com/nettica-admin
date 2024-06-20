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
                Services
                <v-spacer></v-spacer>
                <v-text-field v-if="listView" v-model="search" append-icon="mdi-magnify" label="Search" single-line
                    hide-details></v-text-field>
                <v-spacer></v-spacer>
                <v-btn color="success" @click="startCreateService">
                    Add Service
                    <v-icon right dark>mdi-weather-cloudy</v-icon>
                </v-btn>&nbsp;
                <v-btn color="success" @click="startCreateMultihop">
                    Add Multihop
                    <v-icon right dark>mdi-weather-cloudy</v-icon>
                </v-btn>&nbsp;
                <v-btn color="red" @click="startCreateWilderness">
                    In the Wild
                    <v-icon right dark>mdi-weather-cloudy</v-icon>
                </v-btn>
            </v-card-title>
        </v-card>
        <v-card>
            <v-card-title>
                Subscriptions
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table v-if="listView" :headers="headers" :items="subscriptions" :search="search">
                <template slot="no-data">
                    Welcome to Nettica! Order a subscription on the <a href="https://nettica.com">Nettica website</a> to get
                    started.
                </template>

                <template v-slot:item.issued="{ item }">
                    {{ item.issued | formatDate }}
                </template>
                <template v-slot:item.expires="{ item }">
                    {{ item.expires | formatDate }}
                </template>
                <template v-slot:item.lastUpdated="{ item }">
                    <v-row>
                        <p>At {{ item.lastUpdated | formatDate }} by {{ item.updatedBy }}</p>
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
            <v-data-table v-if="listView" :headers="bottom_headers" :items="services" :search="search"
                @click:row="startUpdateService">
                <template slot="no-data">
                    Creating a service host requires a subscription. Order a subscription on the <a
                        href="https://nettica.com">Nettica website</a> to get started.
                </template>
                <template v-slot:item.created="{ item }">
                    {{ item.created | formatDate }}
                </template>

                <template v-slot:item.updated="{ item }">
                    <v-row>
                        <p>At {{ item.updated | formatDate }} by {{ item.updatedBy }}</p>
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
        <v-dialog v-if="subscriptions" v-model="dialogCreateService" max-width="550">
            <v-card>
                <v-card-title class="headline">Create New Service</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-select return-object v-model="svcList.selected" :items="svcList.items" item-text="text"
                                    item-value="value" label="Choose type of Service"
                                    :rules="[v => !!v || 'Service is required',]" single persistent-hint required />
                                <v-select return-object v-model="serverList.selected" :items="serverList.items"
                                    item-text="text" item-value="value" label="Pick a region"
                                    :rules="[v => !!v || 'Server is required',]" single persistent-hint required />
                                <v-select return-object v-model="netList.selected" :items="netList.items" item-text="text"
                                    item-value="value" label="To this net" :rules="[v => !!v || 'Net is required',]"
                                    single persistent-hint required />
                                <v-select return-object v-model="dnsList.selected" :items="dnsList.items" item-text="text"
                                    item-value="value" label="Select a DNS provider"
                                    :rules="[v => !!v || 'DNS is required',]" single persistent-hint required />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="success" @click="create()">
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="primary" @click="dialogCreateService = false">
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog v-if="subscriptions" v-model="dialogCreateMultihop" max-width="550">
            <v-card>
                <v-card-title class="headline">Create New Multihop Service</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-select return-object v-model="ingressList.selected" :items="ingressList.items"
                                    item-text="text" item-value="value" label="Pick a region for ingress"
                                    :rules="[v => !!v || 'Ingress region is required',]" single persistent-hint
                                    required />
                                <v-select return-object v-model="egressList.selected" :items="egressList.items"
                                    item-text="text" item-value="value" label="Pick a region for egress"
                                    :rules="[v => !!v || 'Egress region is required',]" single persistent-hint required />
                                <v-select return-object v-model="netList.selected" :items="netList.items" item-text="text"
                                    item-value="value" label="To this net" :rules="[v => !!v || 'Net is required',]"
                                    single persistent-hint required />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="success" @click="create_multihop()">
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="primary" @click="dialogCreateMultihop = false">
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog v-if="subscriptions" v-model="dialogWilderness" max-width="550">
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
                                <v-btn v-if="!wild" color="error" @click="loginWild">
                                    Login
                                    <v-icon right dark>mdi-lock</v-icon>
                                </v-btn>
                                <div v-if="wild">
                                    <v-select return-object v-model="svcList.selected" :items="svcList.items" item-text="text"
                                        item-value="value" label="Choose type of Service"
                                        :rules="[v => !!v || 'Service is required',]" single persistent-hint required />
                                    <v-select return-object v-model="serverList.selected" :items="serverList.items"
                                        item-text="text" item-value="value" label="Pick a region"
                                        :rules="[v => !!v || 'Server is required',]" single persistent-hint required />
                                    <v-select return-object v-model="wildList.selected" :items="wildList.items" item-text="text"
                                        item-value="value" label="To this network" :rules="[v => !!v || 'Network is required',]"
                                        single persistent-hint required />
                                </div>
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions v-if="wild">
                    <v-spacer />
                    <v-btn :disabled="!valid" color="success" @click="createWilderness">
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn color="primary" @click="dialogWilderness = false">
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'
import TokenService from '../services/token.service';
import ApiService from '../services/api.service';

export default {
    name: 'Services',

    data: () => ({
        listView: true,
        dialogCreateService: false,
        dialogCreateMultihop: false,
        dialogUpdate: false,
        dialogWilderness: false,
        showWilderness: false,
        wildServer: "",
        wildApiKey: "",
        wildList: {},
        wild: false,
        inDelete: false,
        credits: 0,
        used: 0,
        netList: {},
        serverList: {},
        ingressList: {},
        egressList: {},
        server: null,
        ingressServer: null,
        egressServer: null,
        user: null,
        subscription: null,
        service: null,
        ingress: null,
        egress: null,
        panel: 1,
        valid: false,
        search: '',
        dnsList: {
            items: [
                { text: "Google DNS", value: "8.8.8.8" },
                { text: "Cloudflare DNS", value: "1.1.1.1" },
                { text: "OpenDNS DNS", value: "208.67.222.222" },
                { text: "Quad9 DNS", value: "9.9.9.9" },
            ]
        },
        svcList: {
            items: [
                { text: "Tunnel Service (tunnel all traffic through the Service Host)", value: "Tunnel" },
                { text: "Relay Service  (allows all machines in net to communicate with each other)", value: "Relay" },
            ]
        },
        headers: [
            { text: 'Name', value: 'name', },
            { text: "Description", value: 'description', },
            { text: 'Issued', value: 'issued', },
            { text: 'Credits', value: 'credits', },
            { text: 'Status', value: 'status', },
            { text: 'Actions', value: 'action', sortable: false, },

        ],
        bottom_headers: [
            { text: 'Name', value: 'device.name', },
            { text: "Location", value: 'description', },
            { text: 'Service', value: 'serviceType' },
            { text: 'Created', value: 'created', },
            { text: 'Status', value: 'status', },
            { text: 'Actions', value: 'action', sortable: false, },

        ],
    }),

    computed: {
        ...mapGetters({
            authuser: 'auth/user',
            subscriptions: 'subscription/subscriptions',
            services: 'service/services',
            servers: 'server/servers',
            nets: 'net/nets',
            wildnets: 'wildnet/nets',
            devices: 'device/devices',
            vpns: 'vpn/vpns',
        }),
    },

    mounted() {

        ApiService.setServer()
        ApiService.setHeader()

        this.readAllNetworks()
        this.readServers()
        this.readSubscriptions(this.authuser.email)
        this.readServices(this.authuser.email)

        if (TokenService.getWildServer()) {
            this.wildServer = TokenService.getWildServer()
            this.wild = true
            this.readWildNets()
        } else {
            this.wild = false
        }

    },

    watch: {
        subscriptions: function (val) {
            console.log("subscriptions", val)
        },
        services: function (val) {
            console.log("services", val)
        },
        wildnets: function (w) {
            console.log("wildnets", w)
            var wildList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            for (let i = 0; i < w.length; i++) {
                wildList.items[i] = { "text": w[i].netName, "value": w[i].id }
            }
            this.wildList = wildList
        },
        servers: function (val) {
            console.log("servers", val)
            var serverList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            for (let i = 0; i < val.length; i++) {
                serverList.items[i] = { "text": val[i].description, "value": val[i].name }
            }
            this.serverList = serverList
        },
    },


    methods: {
        ...mapActions('subscription', {
            readSubscriptions: 'read',
            deleteSubscription: 'delete',
            updateSubscription: 'update',
        }),

        ...mapActions('service', {
            errorService: 'error',
            readServices: 'read',
            createService: 'create',
            updateService: 'update',
            deleteService: 'delete',
        }),

        ...mapActions('device', {
            createDevice: 'create',
        }),

        ...mapActions("vpn", {
            createVPN: 'create',
        }), 

        ...mapActions('server', {
            readServers: 'read',
        }),

        ...mapActions('net', {
            readAllNetworks: 'readAll',
        }),

        ...mapActions('wildnet', {
            readWildNets: 'readAll',
            setServer: 'wildServer',
        }),

        Refresh() {
            this.readSubscriptions(this.authuser.email)
            this.readServices(this.authuser.email)
            this.readServers()
        },

        startCreateService() {
            this.credits = 0;
            for (var i = 0; i < this.subscriptions.length; i++) {
                if (this.subscriptions[i].status == "active") {
                    this.credits += this.subscriptions[i].credits;
                }
            }
            if (this.credits <= this.services.length) {
                alert("You have exceeded your credit limit. Please purchase more credits to create a new service.")
                return
            }
            this.dialogCreateService = true;
            this.service = {
                name: "",
                email: this.authuser.email,
            }
            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = 0;
            this.netList.items[0] = { "text": "New Network", "value": "" }
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i + 1] = { "text": this.nets[i].netName, "value": this.nets[i].id }
            }

            this.netList.selected = this.netList.items[selected];

            this.serverList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            for (let i = 0; i < this.servers.length; i++) {
                this.serverList.items[i] = { "text": this.servers[i].description, "value": this.servers[i].name }
            }
        },

        startCreateMultihop() {
            this.credits = 0;
            for (var i = 0; i < this.subscriptions.length; i++) {
                if (this.subscriptions[i].status == "active") {
                    this.credits += this.subscriptions[i].credits;
                }
            }
            if ((this.credits - 1) <= this.services.length) {
                alert("Multihop requires 2 credits and exceeds your current limit. Please purchase more credits to create a new multihop service.")
                return
            }
            this.dialogCreateMultihop = true;
            this.ingress = {
                name: "",
                email: this.authuser.email,
            }
            this.egress = {
                name: "",
                email: this.authuser.email,
            }
            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = -1;
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i] = { "text": this.nets[i].netName, "value": this.nets[i].id }
            }

            this.netList.selected = this.netList.items[selected];

            this.ingressList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            for (let i = 0; i < this.servers.length; i++) {
                this.ingressList.items[i] = { "text": this.servers[i].description, "value": this.servers[i].name }
            }
            this.egressList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            for (let i = 0; i < this.servers.length; i++) {
                this.egressList.items[i] = { "text": this.servers[i].description, "value": this.servers[i].name }
            }

        },

        create() {

            for (let i = 0; i < this.serverList.items.length; i++) {
                if (this.serverList.items[i].value == this.serverList.selected.value) {
                    this.server = this.servers[i];
                }
            }

            var range = this.server.portMax - this.server.portMin + 1;
            var port = this.server.portMin + Math.floor(Math.random() * range);

            this.service.defaultSubnet = this.server.defaultSubnet;
            this.service.servicePort = port;
            this.service.net = {}
            this.service.net.netName = this.netList.selected.value;
            this.service.net.default = {}
            this.service.net.default.dns = []
            this.service.net.default.dns[0] = this.dnsList.selected.value;

            this.service.vpn = {}
            this.service.vpn.netName = this.serverList.selected.value;
            this.service.vpn.current = {}
            this.service.vpn.current.dns = []
            this.service.vpn.current.dns[0] = this.dnsList.selected.value;
            this.service.vpn.current.endpoint = this.server.ipAddress + ":" + port;
            this.service.vpn.current.listenPort = port;

            this.service.description = this.server.description
            this.service.name = this.server.name
            this.service.serviceGroup = this.server.serviceGroup
            this.service.apiKey = this.server.serviceApiKey

            this.service.serviceType = this.svcList.selected.value;

            if (this.service.net.netName != "") {
                this.service.net.id = this.netList.selected.value;
            }
            else {
                this.service.net.id = "";
            }

            this.createService(this.service);

            this.dialogCreateService = false;

        },

        create_multihop() {

            for (let i = 0; i < this.ingressList.items.length; i++) {
                if (this.ingressList.items[i].value == this.ingressList.selected.value) {
                    this.ingressServer = this.servers[i];
                }
            }

            for (let i = 0; i < this.egressList.items.length; i++) {
                if (this.egressList.items[i].value == this.egressList.selected.value) {
                    this.egressServer = this.servers[i];
                }
            }

            var rangeI = this.ingressServer.portMax - this.ingressServer.portMin + 1;
            var portI = this.ingressServer.portMin + Math.floor(Math.random() * rangeI);

            this.ingress.defaultSubnet = this.ingressServer.defaultSubnet;
            this.ingress.servicePort = portI;
            this.ingress.description = this.ingressServer.description + " (ingress)"
            this.ingress.name = this.ingressServer.name
            this.ingress.serviceGroup = this.ingressServer.serviceGroup
            this.ingress.apiKey = this.ingressServer.serviceApiKey

            this.ingress.net = {}
            this.ingress.net.netName = this.netList.selected.value;
            this.ingress.net.default = {}
            this.ingress.net.default.dns = []
            this.ingress.vpn = {}
            this.ingress.vpn.current = {}
            this.ingress.vpn.current.dns = []
            this.ingress.vpn.current.endpoint = this.ingressServer.ipAddress + ":" + portI;
            this.ingress.vpn.current.listenPort = portI;

            this.ingress.serviceType = "Ingress";

            if (this.ingress.net.netName != "") {
                this.ingress.net.id = this.netList.selected.value;
            }
            else {
                this.ingress.net.id = "";
            }

            var rangeE = this.egressServer.portMax - this.egressServer.portMin + 1;
            var portE = this.egressServer.portMin + Math.floor(Math.random() * rangeE);

            this.egress.defaultSubnet = this.ingressServer.defaultSubnet;
            this.egress.servicePort = portE;
            this.egress.description = this.egressServer.description + " (egress)"
            this.egress.name = this.egressServer.name
            this.egress.serviceGroup = this.egressServer.serviceGroup
            this.egress.apiKey = this.egressServer.serviceApiKey
            this.egress.net = {}
            this.egress.net.netName = this.netList.selected.text;
            this.egress.net.default = {}
            this.egress.net.default.dns = []
            this.egress.vpn = {}
            this.egress.vpn.current = {}
            this.egress.vpn.current.dns = []
            this.egress.vpn.current.endpoint = this.egressServer.ipAddress + ":" + portE;
            this.egress.vpn.current.listenPort = portE;

            this.egress.serviceType = "Egress";

            if (this.egress.net.netName != "") {
                this.egress.net.id = this.netList.selected.value;
            }
            else {
                this.egress.net.id = "";
            }


            this.createService(this.ingress);
            this.createService(this.egress);

            this.dialogCreateMultihop = false;

        },

        startCreateWilderness() {

            var server = TokenService.getWildServer()
            if (server && this.wild == false) {    
                this.wildServer = server
                this.wild = true
                this.readWildNets()
            }

            this.dialogWilderness = true;

        },

        loginWild() {

            if (this.wildServer == "") {
                this.errorService("Server name is required")
                return
            }

            if (!this.wildServer.toLowerCase().startsWith("http")) {
                this.errorService("Server name should start with http or https")
                return
            }

            if (this.wildServer.endsWith("/")) {
                this.wildServer = this.wildServer.slice(0, -1)
            }


            this.dialogWilderness = false;
            
            var url = this.wildServer + "/?referer=" + window.location.origin
            window.open( url )
        },

        async createWilderness() {
            console.log("createWilderness");

            var device = {}

            for (let i = 0; i < this.serverList.items.length; i++) {
                if (this.serverList.items[i].value == this.serverList.selected.value) {
                    this.server = this.servers[i];
                }
            }

            var range = this.server.portMax - this.server.portMin + 1;
            var port = this.server.portMin + Math.floor(Math.random() * range);

            this.service = {}
            this.service.defaultSubnet = this.server.defaultSubnet;
            this.service.servicePort = port;
            this.service.net = {}
            this.service.net.netName = this.wildList.selected.text;
            this.service.net.id = this.wildList.selected.value;  
            this.service.net.default = {}
            this.service.net.default.dns = []

            var vpn = {
                name: this.server.name + "." + this.wildList.selected.text,
                enable: true,
                netName: this.wildList.selected.text,
                netId: this.wildList.selected.value,
                type: "Service",
                description: this.server.description,
                // deviceid: this.devices[0].id,
                current: {
                    endpoint: this.server.ipAddress + ":" + port,
                    listenPort: port,
                    subnetRouting: true,
                }
            }

            //this.service.device = this.devices[0];
            //this.service.vpn = this.vpns[0];
            this.service.description = this.server.description
            this.service.name = this.server.name
            this.service.serviceGroup = this.server.serviceGroup
            this.service.serviceType = this.svcList.selected.value;
            this.service.email = this.authuser.email;
            this.service.server = this.wildServer;

            device.name = this.svcList.selected.value.toLowerCase() + "." + this.server.name;
            device.description = `nettica.com ${this.svcList.selected.value.toLowerCase()} service in ${this.server.description}`;
            device.enable = true;

            // switch to the wilderness
            ApiService.setWildServer();
            ApiService.setWildHeader();
            ApiService.post("/device", device)
                .then( d => {
                this.errorService(`Device created: ${d.name}`);
                console.log("device created: ", d);
                vpn.deviceid = d.id;

                ApiService.setWildServer();
                ApiService.setWildHeader();
                ApiService.post("/vpn", vpn)
                    .then( v => {
                        console.log("vpn created: ", v);
                        this.errorService(`VPN created: ${v.name}`);

                    this.service.device = d;   
                    this.service.vpn = v;  

                    ApiService.setServer();
                    ApiService.setHeader();
                    ApiService.post("/service", this.service)
                        .then( s => {
                            console.log("service created: ", s);
                            this.errorService(`Service created: ${s.name}`);
                            this.Refresh();
                        })
                        .catch(error => {
                            console.log("error: ", error)
                            if (error.response) {
                                this.errorService(error.response.data.error)
                            }
                        });
                    })
                    .catch(error => {
                        console.log("error: ", error)
                        if (error.response) {
                            this.errorService(error.response.data.error)
                        }
                    });
            })
            .catch(error => {
                console.log("error: ", error)
                if (error.response) {
                this.errorService(error.response.data.error)
                }
            });
            this.dialogWilderness = false;
        },


        removeSubscription(item) {
            this.inDelete = true;
            if (confirm(`Do you really want to delete ${item.name} (not recommended)? This is an irreversible action.`)) {
                this.deleteSubscription(item)
            }
        },


        remove(item) {
            this.inDelete = true;
            if (confirm(`Do you really want to delete ${item.name} ?`)) {
                this.deleteService(item)
            }
        },

        startUpdateSubscription(subscription) {
            if (this.inDelete == true) {
                this.inDelete = false;
                return
            }
            this.subscription = subscription;
            this.dialogSubscription = true;
        },

        updateSubscription(subscription) {

            this.dialogSubscription = false;
            this.updateSubscription(subscription)
        },

        startUpdateService(service) {
            if (this.inDelete == true) {
                this.inDelete = false;
                return
            }
            this.service = service;
            this.dialogService = true;
        },

        updateService(service) {
            this.dialogService = false;
            this.updateService(service)
        },

    }
};
</script>
