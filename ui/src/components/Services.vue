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
                Services & Subscriptions
                <v-spacer></v-spacer>
                <v-text-field
                        v-if="listView"
                        v-model="search"
                        append-icon="mdi-magnify"
                        label="Search"
                        single-line
                        hide-details
                ></v-text-field>
                <v-spacer></v-spacer>
                <v-btn
                        color="success"
                        @click="startCreateService"
                >
                    Add Service Host
                    <v-icon right dark>mdi-weather-cloudy</v-icon>
                </v-btn>&nbsp; 
                <v-btn
                        color="success"
                        @click="startCreateMultihop"
                >
                    Add Multihop
                    <v-icon right dark>mdi-weather-cloudy</v-icon>
                </v-btn>
            </v-card-title>
        </v-card>
        <v-card>
            <v-card-title>
                Subscriptions
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table
                    v-if="listView"
                    :headers="headers"
                    :items="subscriptions"
                    :search="search"
            >
                <template slot="no-data">
                    Welcome to Nettica!  Order a subscription on the <a href="https://nettica.com">Nettica website</a> to get started.
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
                        <v-icon
                                class="pr-1 pl-1"
                                @click="removeSubscription(item)"
                                title = "Remove Subscription (not recommended)"
                        >
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
            <v-data-table
                    v-if="listView"
                    :headers="bottom_headers"
                    :items="services"
                    :search="search"
                    @click:row="startUpdateService"
            >
                <template slot="no-data">
                    Creating a service host requires a subscription.  Order a subscription on the <a href="https://nettica.com">Nettica website</a> to get started.
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

                        <v-icon
                                class="pr-1 pl-1"
                                @click="remove(item)"
                                title = "Delete Service Host"
                        >
                            mdi-trash-can-outline
                        </v-icon>
                    </v-row>
                </template>

            </v-data-table>
        </v-card>
        <v-dialog
                v-if="subscriptions"
                v-model="dialogCreateService"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Create New Service Host</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                                <v-select return-object
                                        v-model="serverList.selected"
                                        :items="serverList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Pick a region"
                                        :rules="[ v => !!v || 'Server is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                <v-select return-object
                                        v-model="meshList.selected"
                                        :items="meshList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="To this mesh"
                                        :rules="[ v => !!v || 'Mesh is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                <v-select return-object
                                        v-model="dnsList.selected"
                                        :items="dnsList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Select a DNS provider"
                                        :rules="[ v => !!v || 'DNS is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                 <v-select return-object
                                        v-model="svcList.selected"
                                        :items="svcList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Choose type of Service"
                                        :rules="[ v => !!v || 'Service is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />

                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer/>
                    <v-btn
                            :disabled="!valid"
                            color="success"
                            @click="create()"
                    >
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn
                            color="primary"
                            @click="dialogCreateService = false"
                    >
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog
                v-if="subscriptions"
                v-model="dialogCreateMultihop"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Create New Multihop Service</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                            <v-select return-object
                                        v-model="ingressList.selected"
                                        :items="ingressList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Pick a region for ingress"
                                        :rules="[ v => !!v || 'Ingress region is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                <v-select return-object
                                        v-model="egressList.selected"
                                        :items="egressList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Pick a region for egress"
                                        :rules="[ v => !!v || 'Egress region is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                <v-select return-object
                                        v-model="meshList.selected"
                                        :items="meshList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="To this mesh"
                                        :rules="[ v => !!v || 'Mesh is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer/>
                    <v-btn
                            :disabled="!valid"
                            color="success"
                            @click="create_multihop()"
                    >
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn
                            color="primary"
                            @click="dialogCreateMultihop = false"
                    >
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

  export default {
    name: 'Services',

    data: () => ({
      listView: true,
      dialogCreateService: false,
      dialogCreateMultihop: false,
      dialogUpdate: false,
      dialogMember : false,
      inDelete: false,
      credits : 0,
      used: 0,
      meshList: {},
      serverList: {},
      ingressList: {},
      egressList: {},
      server: null,
      ingressServer: null,
      egressServer: null,
      roles : ["Owner", "Admin", "User"],
      statuses : ["Active", "Pending", "Suspended", "Hidden"],
      user: null,
      member: null,
      account: null,
      subscription: null,
      service: null,
      ingress: null,
      egress : null,
      panel: 1,
      valid: false,
      search: '',
      dnsList:{ items: [
        { text: "Google DNS", value: "8.8.8.8" },
        { text: "Cloudflare DNS", value: "1.1.1.1" },
        { text: "OpenDNS DNS", value: "208.67.222.222"},
        { text: "Quad9 DNS", value: "9.9.9.9"},
      ]},
      svcList:{ items: [
        { text: "Relay Service  (allows all machines in mesh to communicate with each other)", value: "Relay" },
        { text: "Tunnel Service (tunnel all traffic through the Service Host)", value: "Tunnel" },
      ]},
      headers: [
        { text: 'Name', value: 'name', },
        { text: "Description", value: 'description', },
        { text: 'Issued', value: 'issued', },
        { text: 'Credits', value: 'credits', },
        { text: 'Status', value: 'status', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
      bottom_headers: [
        { text: 'Name', value: 'relayHost.name', },
        { text: "Location", value: 'description', },
        { text: 'Service', value: 'serviceType'},
        { text: 'Created', value: 'created', },
        { text: 'Status', value: 'status', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
    }),

    computed:{
      ...mapGetters({
        authuser: 'auth/user',
        accounts: 'account/accounts',
        members: 'account/users',
        subscriptions: 'subscription/subscriptions',
        services: 'service/services',
        servers: 'server/servers',
        meshes: 'mesh/meshes',
      }),
    },

    mounted () {
      this.readAllMeshes()
      this.readSubscriptions(this.authuser.email)
      this.readServices(this.authuser.email)
      this.readServers()

    },

    watch: {
      // whenever accounts changes, this function will run
        //      subscriptions(newSubscriptions, oldSubscriptions) {
          //      this.readSubscriptions(this.authuser.email);
      //}
    },

    methods: {
        ...mapActions('account', {
            readAllAccounts: 'readAll',
            readUsers: 'readUsers',
            createAccount: 'create',
            updateAccount: 'update',
            emailUser: 'email',
        }),

        ...mapActions('subscription', {
            readSubscriptions: 'read',
            deleteSubscription: 'delete',
            updateSubscription: 'update',
        }),

        ...mapActions('service', {
            readServices: 'read',
            createService: 'create',
            updateService: 'update',
            deleteService: 'delete',
        }),

        ...mapActions('server', {
            readServers: 'read',
        }),

        ...mapActions('mesh', {
            readAllMeshes: 'readAll',
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
        this.meshList = { selected: { "text": "",  "value": ""},
                          items: [] }

        var selected = 0;
        this.meshList.items[0] = { "text": "New Mesh", "value": ""}
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i+1]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
        }

        this.meshList.selected = this.meshList.items[selected];

        this.serverList = { selected: { "text": "",  "value": ""},
                          items: [] }
        for (let i=0; i<this.servers.length; i++) {
            this.serverList.items[i]= { "text": this.servers[i].description, "value": this.servers[i].name }
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
        this.meshList = { selected: { "text": "",  "value": ""},
                          items: [] }

        var selected = -1;
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
        }

        this.meshList.selected = this.meshList.items[selected];

        this.ingressList = { selected: { "text": "",  "value": ""},
                          items: [] }
        for (let i=0; i<this.servers.length; i++) {
            this.ingressList.items[i]= { "text": this.servers[i].description, "value": this.servers[i].name }
        }
        this.egressList = { selected: { "text": "",  "value": ""},
                          items: [] }
        for (let i=0; i<this.servers.length; i++) {
            this.egressList.items[i]= { "text": this.servers[i].description, "value": this.servers[i].name }
        }

    },

    create() {

    for (let i=0; i<this.serverList.items.length; i++) {
        if (this.serverList.items[i].value == this.serverList.selected.value) {
            this.server = this.servers[i];
        }
    }

    var range = this.server.portMax - this.server.portMin + 1;
    var port = this.server.portMin + Math.floor(Math.random() * range);

    this.service.defaultSubnet = this.server.defaultSubnet;
    this.service.servicePort = port;
    this.service.relayHost = {}
    this.service.relayHost.meshName = this.serverList.selected.value;
    this.service.relayHost.current = {}
    this.service.relayHost.current.dns = []
    this.service.relayHost.current.dns[0] = this.dnsList.selected.value;
    this.service.relayHost.current.endpoint = this.server.ipAddress + ":" + port;
    this.service.relayHost.current.listenPort = port;
    this.service.description = this.server.description
    this.service.name = this.server.name
    this.service.serviceGroup = this.server.serviceGroup
    this.service.apiKey = this.server.serviceApiKey

    this.service.serviceType = this.svcList.selected.value;

    if (this.service.relayHost.meshName != "") {
        this.service.relayHost.meshId = this.meshList.selected.value;
    }
    else {
        this.service.relayHost.meshId = "";
    }

    this.createService(this.service);

    this.dialogCreateService = false;

    },

    create_multihop() {

    for (let i=0; i<this.ingressList.items.length; i++) {
        if (this.ingressList.items[i].value == this.ingressList.selected.value) {
            this.ingressServer = this.servers[i];
        }
    }

    for (let i=0; i<this.egressList.items.length; i++) {
        if (this.egressList.items[i].value == this.egressList.selected.value) {
            this.egressServer = this.servers[i];
        }
    }

    var rangeI = this.ingressServer.portMax - this.ingressServer.portMin + 1;
    var portI = this.ingressServer.portMin + Math.floor(Math.random() * rangeI);

    this.ingress.defaultSubnet = this.ingressServer.defaultSubnet;
    this.ingress.servicePort = portI;
    this.ingress.relayHost = {}
    this.ingress.relayHost.meshName = this.meshList.selected.value;
    this.ingress.relayHost.current = {}
    this.ingress.relayHost.current.dns = []
    this.ingress.relayHost.current.dns[0] = "8.8.8.8";
    this.ingress.relayHost.current.endpoint = this.ingressServer.ipAddress + ":" + portI;
    this.ingress.relayHost.current.listenPort = portI;
    this.ingress.description = this.ingressServer.description + " (ingress)"
    this.ingress.name = this.ingressServer.name
    this.ingress.serviceGroup = this.ingressServer.serviceGroup
    this.ingress.apiKey = this.ingressServer.serviceApiKey

    this.ingress.serviceType = "Ingress";

    if (this.ingress.relayHost.meshName != "") {
        this.ingress.relayHost.meshId = this.meshList.selected.value;
    }
    else {
        this.ingress.relayHost.meshId = "";
    }

    var rangeE = this.egressServer.portMax - this.egressServer.portMin + 1;
    var portE = this.egressServer.portMin + Math.floor(Math.random() * rangeE);

    this.egress.defaultSubnet = this.ingressServer.defaultSubnet;
    this.egress.servicePort = portE;
    this.egress.relayHost = {}
    this.egress.relayHost.meshName = this.meshList.selected.value;
    this.egress.relayHost.current = {}
    this.egress.relayHost.current.dns = []
    this.egress.relayHost.current.dns[0] = "8.8.8.8";
    this.egress.relayHost.current.endpoint = this.egressServer.ipAddress + ":" + portE;
    this.egress.relayHost.current.listenPort = portE;
    this.egress.description = this.egressServer.description + " (egress)"
    this.egress.name = this.egressServer.name
    this.egress.serviceGroup = this.egressServer.serviceGroup
    this.egress.apiKey = this.egressServer.serviceApiKey

    this.egress.serviceType = "Egress";

    if (this.egress.relayHost.meshName != "") {
        this.egress.relayHost.meshId = this.meshList.selected.value;
    }
    else {
        this.egress.relayHost.meshId = "";
    }


    this.createService(this.ingress);
    this.createService(this.egress);

    this.dialogCreateMultihop = false;

    },

      removeSubscription(item) {
        this.inDelete = true;
        if (confirm(`Do you really want to delete ${item.name} ?`)){
          this.deleteSubscription(item)
        }
        this.readAllAccounts(this.authuser.email)
        this.readAllMeshes()
        this.readServices(this.authuser.email)

      },


      remove(item) {
        this.inDelete = true;
        if (confirm(`Do you really want to delete ${item.name} ?`)){
          this.deleteService(item)
        }
        this.readAllAccounts(this.authuser.email)
        this.readAllMeshes()
        this.readServices(this.authuser.email)

      },

      email(toAddress, mesh) {
        this.dialogCreateService = false;
        if (!toAddress) {
          this.errorUser('email address is not defined')
          return
        }

        for (let i=0; i<this.accounts.length; i++) {
            if (this.accounts[i].id == this.accounts[i].parent) {
                this.emailUser(toAddress + "/" + this.accounts[i].id);
                break;        
            }
        }

      },

      startUpdate(user) {
        if (this.inDelete == true ) {
            this.inDelete = false;
            return
        }
        this.user = user;
        this.dialogUpdate = true;
      },

      update(user) {

        this.dialogUpdate = false;
        this.updateAccount(user)
      },

      updateMember(member) {

        this.dialogMember = false;
        this.member.meshName = this.meshList.selected.text;
        this.member.meshId = this.meshList.selected.value;

        this.updateAccount(member)
      },      

      startUpdateMember(member) {
        if (this.inDelete == true ) {
            this.inDelete = false;
            return
        }

        var selected = 0;
        this.meshList.items = [];
        this.meshList.items[0] = { "text": "All Meshes", "value": ""};
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i+1]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
            if (this.meshes[i].id == member.meshId) {
                selected = i+1;
            }
        }
        this.meshList.selected = this.meshList.items[selected];

        this.member = member;
        this.dialogMember = true;

      },

      startUpdateSubscription(subscription) {
        if (this.inDelete == true ) {
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
        if (this.inDelete == true ) {
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
