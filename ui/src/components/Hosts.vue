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
                Hosts
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
                        @click="startCreate"
                >
                    Add host manually
                    <v-icon right dark>mdi-network-outline</v-icon>
                </v-btn>
            </v-card-title>
            <v-data-table
                    v-if="listView"
                    :headers="headers"
                    :items="hosts"
                    :search="search"
                    :items-per-page="25"
                    no-data-text="No Hosts.  Click above to create your first host, or use Meshify Agent on the host."
                    no-results-text="No results matching your search"
                    :footer-props="footerProps"
            >

                <template v-slot:item.name="{ item }">
                        {{ item.name }}
                </template>
                <template v-slot:item.status="{ item }">
                    <v-icon v-if="item.status == 'Online'" color="green">mdi-check-circle</v-icon>
                    <v-icon v-else-if="item.status == 'Native'" color="blue">mdi-minus-circle</v-icon>
                    <v-icon v-else color="red">mdi-close-circle</v-icon>
                    {{ item.status }}
                </template>
                
                <template v-slot:item.address="{ item }">
                    <v-chip
                            v-for="(ip, i) in item.address"
                            :key="i"
                            color="#336699"
                            text-color="white"
                    >
                        <v-icon left>mdi-ip-network</v-icon>
                        {{ ip }}
                    </v-chip>
                </template>
                <template v-slot:item.tags="{ item }">
                    <v-chip
                            v-for="(tag, i) in item.tags"
                            :key="i"
                            color="blue-grey"
                            text-color="white"
                    >
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
                    <v-row v-if="item.type == 'ServiceHost'" >
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdateServiceHost(item)"
                                title = "Edit Service"
                        >
                            mdi-square-edit-outline
                        </v-icon>
                    </v-row>
                    <v-row v-if="item.type != 'ServiceHost'" >
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                                title = "Edit"
                        >
                            mdi-square-edit-outline
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startCopy(item)"
                                title = "Copy"
                        >
                            mdi-content-copy
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click="remove(item)"
                                title = "Delete"
                        >
                            mdi-trash-can-outline
                        </v-icon>
                        <v-switch
                                dark
                                class="pr-1 pl-1"
                                color="success"
                                v-model="item.enable"
                                v-on:change="updateEnable(item)"
                        />
                    </v-row>
                </template>

            </v-data-table>
        </v-card>
        <v-dialog
                v-if="host"
                v-model="dialogCreate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Add New Host</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                                <v-text-field
                                        v-model="host.name"
                                        label="Host friendly name"
                                        :rules="[ v => !!v || 'host name is required', ]"
                                        required
                                />
                                <v-select return-object
                                        v-model="meshList.selected"
                                        :items="meshList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Join this mesh"
                                        :rules="[ v => !!v || 'Mesh is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                <v-text-field
                                        v-model="host.current.endpoint"
                                        label="Public endpoint for clients"
                                />
                                <v-text-field
                                        v-model="host.current.listenPort"
                                        type="number"
                                        label="Listen port"
                                />

                                <v-combobox
                                        v-model="host.tags"
                                        chips
                                        hint="Enter a tag, hit tab, hit enter."
                                        label="Tags"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="host.tags.splice(host.tags.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-switch
                                        v-model="host.enable"
                                        color="success"
                                        inset
                                        :label="host.enable ? 'Enable host after creation': 'Disable host after creation'"
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
                            @click="create(host)"
                    >
                        Submit
                        <v-icon right dark>mdi-check-outline</v-icon>
                    </v-btn>
                    <v-btn
                            color="primary"
                            @click="dialogCreate = false"
                    >
                        Cancel
                        <v-icon right dark>mdi-close-circle-outline</v-icon>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog
                v-if="host"
                v-model="dialogUpdate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Edit Host</v-card-title>
                <v-card-text>

                    <v-row>
                        <v-col
                                cols="12"
                        >
                            <v-form
                                    ref="form"
                                    v-model="valid"
                            >
                                <v-text-field
                                        v-model="host.name"
                                        label="Friendly name"
                                        :rules="[ v => !!v || 'host name is required',]"
                                        required
                                />

                                <v-select return-object
                                        v-model="meshList.selected"
                                        :items="meshList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Join this mesh"
                                        :rules="[ v => !!v || 'Mesh is required', ]"
                                        single
                                        persistent-hint
                                        required
                                />
                                <v-text-field
                                        v-model="host.current.endpoint"
                                        label="Public endpoint for clients"
                                />
                                <v-text-field
                                        v-model="host.current.listenPort"
                                        type="number"
                                        label="Listen port"
                                />
                                <v-combobox
                                        v-model="host.tags"
                                        chips
                                        hint="Write tag name and hit enter"
                                        label="Tags"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="host.tags.splice(host.tags.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-btn
                                        color="success"
                                        @click="forceFileDownload(host)"
                                >
                                    Download Config
                                    <v-icon right dark>mdi-cloud-download-outline</v-icon>
                                </v-btn>
<!--                                <v-img :src="'data:image/png;base64, ' + gethostQrcode(host.id)"/> -->
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
            <v-expansion-panels>
            <v-expansion-panel>
                <v-expansion-panel-header dark>Advanced Configuration</v-expansion-panel-header>
                    <v-expansion-panel-content>
                        <div class="d-flex flex-no-wrap justify-space-between">
                            <v-col cols="12">
                                <v-select return-object
                                        v-model="platforms.selected"
                                        :items="platforms.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="Platform of this host"
                                        single
                                        persistent-hint
                                />

                                <v-combobox
                                        v-model="host.current.address"
                                        chips
                                        hint="Write IPv4 or IPv6 CIDR and hit enter"
                                        label="Addresses"
                                        multiple
                                        dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="host.current.address.splice(host.current.address.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                        v-model="host.current.allowedIPs"
                                        chips
                                        hint="Write IPv4 or IPv6 CIDR and hit enter"
                                        label="Allowed IPs"
                                        multiple
                                        dark
                                >
                            
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="host.current.allowedIPs.splice(host.current.allowedIPs.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-switch
                                        v-model="publicSubnets"
                                        color="success"
                                        inset
                                        label="Route all public traffic through tunnel"
                               />

                                <v-combobox
                                    v-model="host.current.dns"
                                    chips
                                    hint="Enter IP address(es) and hit enter or leave empty."
                                    label="DNS servers for this host"
                                    multiple
                                    dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
      
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="host.current.dns.splice(host.current.dns.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-text-field
                                        v-model="host.current.table"
                                        label="Table"
                                />
                                <v-text-field
                                        v-model="host.id"
                                        label="Host ID"
                                        disabled
                                />
                                <v-text-field
                                        v-model="host.current.publicKey"
                                        label="Public key"
                                />
                                <v-text-field
                                        v-model="host.current.privateKey"
                                        label="Private key"
                                        autocomplete="off"
                                        :append-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                                        :type="showPrivate ? 'text' : 'password'"
                                        hint="Clear this field to have the client manage its private key"
                                        @click:append="showPrivate = !showPrivate"

                                />
                                <v-text-field
                                        v-model="host.current.presharedKey"
                                        label="Preshared Key"
                                        autocomplete="off"
                                        :append-icon="showPreshared ? 'mdi-eye' : 'mdi-eye-off'"
                                        :type="showPreshared ? 'text' : 'password'"
                                        @click:append="showPreshared = !showPreshared"
                                />                           
                                <v-text-field
                                        v-model="host.hostGroup"
                                        label="Host Group"
                                />
                                <v-text-field
                                        v-model="host.apiKey"
                                        label="API Key"
                                />
                                <v-text-field
                                        type="number"
                                        v-model="host.current.mtu"
                                        label="Define global MTU"
                                        hint="Leave at 0 and let us take care of MTU"
                                />
                                <v-text-field
                                        type="number"
                                        v-model="host.current.persistentKeepalive"
                                        label="Persistent keepalive"
                                        hint="To disable, set to 0.  Recommended value 29 (seconds)"
                                />
                                <v-textarea
                                        v-model="host.current.postUp"
                                        label="PostUp Script"
                                        hint="Only applies to linux servers"
                                />
                                <v-textarea
                                        v-model="host.current.postDown"
                                        label="PostDown Script"
                                        hint="Only applies to linux servers"
                                />
                                <v-switch
                                        v-model="host.current.subnetRouting"
                                        color="success"
                                        inset
                                        label="Enable subnet routing"
                               />
                                <v-switch
                                        v-model="host.current.upnp"
                                        color="success"
                                        inset
                                        label="Enable UPnP"
                               />
                                <v-switch
                                        v-model="host.current.enableDns"
                                        color="success"
                                        inset
                                        label="Enable Meshify DNS"
                               />

                            </v-col>
                         </div>
                    </v-expansion-panel-content>
                </v-expansion-panel>
            </v-expansion-panels>
            <v-card>
                <v-card-actions>
                        <v-btn
                                :disabled="!valid"
                                color="success"
                                @click="update(host)"
                        >
                            Submit
                            <v-icon right dark>mdi-check-outline</v-icon>
                        </v-btn>
                        <v-btn
                                color="primary"
                                @click="dialogUpdate = false"
                        >
                            Cancel
                            <v-icon right dark>mdi-close-circle-outline</v-icon>
                        </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
        <v-dialog
                v-if="host"
                v-model="dialogCopy"
                max-width="550"
        >
            <v-card>
            <v-card-title class="headline">Copy Host to Mesh</v-card-title>
                    <v-card-text>

                        <v-row>
                            <v-col
                                    cols="12"
                            >
                                <v-form
                                        ref="form"
                                        v-model="valid"
                                >
                                    <v-text-field
                                            v-model="host.name"
                                            label="New name for host"
                                            :rules="[ v => !!v || 'host name is required',]"
                                            required
                                    />

                                    <v-select return-object
                                            v-model="meshList.selected"
                                            :items="meshList.items"
                                            item-text = "text"
                                            item-value = "value"
                                            label="Copy to this mesh"
                                            :rules="[ v => !!v || 'Mesh is required', ]"
                                            single
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
                        <v-btn
                                :disabled="!valid"
                                color="success"
                                @click="copy(host)"
                        >
                            Submit
                            <v-icon right dark>mdi-check-outline</v-icon>
                        </v-btn>
                        <v-btn
                                color="primary"
                                @click="dialogCopy = false"
                        >
                            Cancel
                            <v-icon right dark>mdi-close-circle-outline</v-icon>
                        </v-btn>
                </v-card-actions>
            </v-card>

            </v-dialog>
            <v-dialog
                v-if="host"
                v-model="dialogServiceHost"
                max-width="550"
        >
            <v-card>
            <v-card-title class="headline">Manage Service: {{ host.name }}</v-card-title>
                    <v-card-text>

                        <v-row>
                            <v-col
                                    cols="12"
                            >
                                <v-form
                                        ref="form"
                                        v-model="valid"
                                >
                                    <v-combobox
                                            v-model="host.current.allowedIPs"
                                            chips
                                            hint="Enter IPv4 or IPv6 CIDR and press tab"
                                            label="Allowed IPs"
                                            multiple
                                            dark
                                    >
                                
                                        <template v-slot:selection="{ attrs, item, select, selected }">
                                            <v-chip
                                                    v-bind="attrs"
                                                    :input-value="selected"
                                                    close
                                                    @click="select"
                                                    @click:close="host.current.allowedIPs.splice(host.current.allowedIPs.indexOf(item), 1)"
                                            >
                                                <strong>{{ item }}</strong>&nbsp;
                                            </v-chip>
                                        </template>
                                    </v-combobox>
                                </v-form>
                            </v-col>
                        </v-row>
                    </v-card-text>
                </v-card>
                <v-card>
                <v-card-actions>
                        <v-btn
                                :disabled="!valid"
                                color="success"
                                @click="update(host)"
                        >
                            Submit
                            <v-icon right dark>mdi-check-outline</v-icon>
                        </v-btn>
                        <v-btn
                                color="primary"
                                @click="dialogServiceHost = false"
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
    name: 'Hosts',

    data: () => ({
    
      showPrivate: true,
      showPreshared: true,
      footerProps: {'items-per-page-options': [25, 50, 100, -1]},
      listView: true,
      dialogCreate: false,
      dialogUpdate: false,
      dialogCopy: false,
      dialogServiceHost: false,
      host: null,
      mesh: null,
      name: '',
      panel: 1,
      valid: false,
      meshList: {},
      platList: {},
      publicSubnets: false,
      platforms: { selected: { text:"", value:"" },
                   items: [
                        { text: "Windows", value:"Windows",},
                        { text: "Linux",  value: "Linux", },
                        { text: "MacOS" , value:"MacOS", },
                        { text: "Apple iOS" , value:"iOS", },
                        { text: "Android" , value:"Android", },
                        { text: "Native WireGuard", value: "Native", },
                   ],
        },
      selected: '',
      search: '',
      headers: [
        { text: 'Name', value: 'name', },
        { text: 'Status', value: 'status', },
        { text: 'Mesh', value: 'meshName', },
        { text: 'IP addresses', value: 'current.address', },
//        { text: 'ID', value:'id', },
        { text: "Endpoint", value: 'current.endpoint', },
//        { text: 'Created by', value: 'created', sortable: false, },
        { text: 'Tags', value: 'tags', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
    }),

    computed:{
      ...mapGetters({
        gethostQrcode: 'host/gethostQrcode',
        gethostConfig: 'host/gethostConfig',
        user: 'auth/user',
        servers: 'server/servers',
        accounts: 'account/accounts',
        hosts: 'host/hosts',
        meshes: 'mesh/meshes',
        hostQrcodes: 'host/hostQrcodes',
      }),
    },

    mounted () {
      this.readAllAccounts(this.user.email)
      this.readAllHosts()
      this.readAllMeshes()
    },

    methods: {
      ...mapActions('host', {
        errorhost: 'error',
        readAllHosts: 'readAll',
        readQrCode: 'readQrcode',
        readConfig: 'readConfig',
        createhost: 'create',
        updatehost: 'update',
        deletehost: 'delete',
        emailhost: 'email',
      }),
      ...mapActions('mesh', {
        readAllMeshes: 'readAll',
      }),
      ...mapActions('account', {
          readAllAccounts: 'readAll',
      }),

      Refresh() {
        this.readAllAccounts(this.user.email)
        this.readAllHosts()
        this.readAllMeshes()
      },


      startCreate() {
        this.host = {
          name: "",
          email: this.user.email,
          enable: true,
//          meshName: this.meshes[0].meshName,
//          meshID: this.meshes[0].id,
//          allowedIPs: this.meshes[0].default.allowedIPs,
//          address: this.meshes[0].default.address,
//          meshName: this.meshes[0].default.meshName,
//          id: this.meshes[0].default.id,
          tags: [],
          current: {},
        }
        
        this.meshList = { selected: { "text": "",  "value": ""},
                          items: [] }

        var selected = 0;
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
            if (this.meshList.items[i].text == this.host.meshName) {
                selected = i
            }
        }

        this.meshList.selected = this.meshList.items[selected];
        this.dialogCreate = true;
      },

      create(host) {
        this.host.current.listenPort = parseInt(this.host.current.listenPort, 10); 
        // append the port to the endpoint if it is not there
        if (this.host.current.endpoint != null && this.host.current.endpoint != "" && this.host.current.endpoint.indexOf(":") == -1) {
            if (this.host.current.listenPort == 0 ) {
                this.host.current.listenPort = 51820
            }
            this.host.current.endpoint = this.host.current.endpoint + ":" + this.host.current.listenPort.toString()
        }
        this.host.meshName = this.meshList.selected.text
        this.host.meshid = this.meshList.selected.value
        this.host.platform = this.platforms.selected.value
        this.dialogCreate = false;
        this.createhost(host)
      },

      remove(host) {
        if(confirm(`Do you really want to delete ${host.name} ?`)){
          this.deletehost(host)
        }
      },

      email(host) {
        if (!host.email){
          this.errorhost('host email is not defined')
          return
        }

        if(confirm(`Do you really want to send email to ${host.email} with all configurations ?`)){
          this.emailhost(host)
        }
      },

      startUpdate(host) {

        this.host = host;
//        this.readQrCode(this.host);
        this.readConfig(host);

        this.meshList = { selected: { "text": this.host.meshName,  "value": this.host.meshid },
                          items: [] }

        var selected = 0;
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
            if (this.meshList.items[i].text == this.host.meshName) {
                selected = i
            }
        }

        this.meshList.selected = this.meshList.items[selected];

        for (let i=0; i<this.platforms.items.length; i++) {
            if (this.platforms.items[i].value == this.host.platform) {
                this.platforms.selected = this.platforms.items[i]
                break
            }
        }

        this.publicSubnets = false;
        this.dialogUpdate = true;

      },

      startUpdateServiceHost(host) {

        this.host = host;
        this.readConfig(host);
        
        this.dialogServiceHost = true;

      },
      startCopy(host) {

        this.host = host;
        this.readConfig(host);

        this.meshList = { selected: { "text": this.host.meshName,  "value": this.host.meshid },
                        items: [] }

        var selected = 0;
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
            if (this.meshList.items[i].text == this.host.meshName) {
                selected = i
            }
        }

        this.meshList.selected = this.meshList.items[selected];

        this.dialogCopy = true;
        this.dialogUpdate = false;

        },
      copy(host) {

        this.noEdit = true;
        this.host = host;

        this.host.current.listenPort = parseInt(this.host.current.listenPort, 10);
        this.host.current.persistentKeepalive = parseInt(this.host.current.persistentKeepalive, 10);
        this.host.current.mtu = parseInt(this.host.current.mtu, 10);

        var changed = false;
        if (this.host.meshid != this.meshList.selected.value) {
            this.host.meshName = this.meshList.selected.text
            this.host.meshid = this.meshList.selected.value
            changed = true;
        }
        this.host.meshName = this.meshList.selected.text
        this.host.platform = this.platforms.selected.value

        if (changed) {
            this.host.id = ""
            this.host.current.endpoint = ""
            this.host.current.listenPort = 0
            this.host.meshName = this.meshList.selected.text
            this.host.meshid = this.meshList.selected.value
            this.createhost(host)

        }

        this.readAllHosts();

        this.dialogCopy = false;
      },

      updateEnable(host) {
        // the switch automatically updates host.enable to the proper value
        this.updatehost(host)
      },

      update(host) {

        this.noEdit = true;
        this.host = host;

        this.host.current.listenPort = parseInt(this.host.current.listenPort, 10);
        // append the port to the endpoint if it is not there
        if (this.host.current.endpoint != null && this.host.current.endpoint != "" && this.host.current.endpoint.indexOf(":") == -1) {
            if (this.host.current.listenPort == 0 ) {
                this.host.current.listenPort = 51820
            }
            this.host.current.endpoint = this.host.current.endpoint + ":" + this.host.current.listenPort
        }
        this.host.current.persistentKeepalive = parseInt(this.host.current.persistentKeepalive, 10);
        this.host.current.mtu = parseInt(this.host.current.mtu, 10);

        var changed = false;

        if (this.meshList.selected != null && this.host.meshid != this.meshList.selected.value) {
            this.host.meshName = this.meshList.selected.text
            this.host.meshid = this.meshList.selected.value
            changed = true;
        }
        if (this.meshList.selected != null ) {
            this.host.meshName = this.meshList.selected.text
        }
        if (this.platforms.selected != null) {
            this.host.platform = this.platforms.selected.value
        }
        if (this.publicSubnets) {
            this.host.current.allowedIPs.push("0.0.0.0/5","8.0.0.0/7",
            "11.0.0.0/8","12.0.0.0/6", "16.0.0.0/4", "32.0.0.0/3", "64.0.0.0/3", "96.0.0.0/6",
            "101.0.0.0/8", "102.0.0.0/7", "104.0.0.0/5", "112.0.0.0/5", "120.0.0.0/6", 
            "124.0.0.0/7", "126.0.0.0/8", "128.0.0.0/3", "160.0.0.0/5", "168.0.0.0/6",
            "172.0.0.0/12", "172.32.0.0/11","172.64.0.0/10",
            "172.128.0.0/9","173.0.0.0/8","174.0.0.0/7","176.0.0.0/4","192.0.0.0/9","192.128.0.0/11","192.160.0.0/13","192.169.0.0/16",
            "192.170.0.0/15","192.172.0.0/14","192.176.0.0/12","192.192.0.0/10","193.0.0.0/8","194.0.0.0/7","196.0.0.0/6","200.0.0.0/5","208.0.0.0/4")
        }
        if (changed) {
            for (let i=0; i<this.meshes.length; i++) {
                if ( this.host.meshid == this.meshList.items[i].value ) {
                    var template = this.meshes[i]
                    this.host.current.address = []

                    if (host.current.listenPort == host.default.listenPort) {
                        host.current.listenPort = template.listenPort
                    }

                    if (host.current.allowedIPs == host.default.allowedIPs) {
                        host.current.allowedIPs = template.allowedIPs
                    } else {
                        host.current.allowedIPs = []
                    }

                    if (host.current.mtu == host.default.mtu) {
                        host.current.mtu = template.mtu
                    }

                    if (host.current.dns == host.default.dns) {
                        host.current.dns = template.dns
                    } else {
                        host.current.dns = []
                    }

                    if (host.current.persistentKeepalive == host.default.persistentKeepalive) {
                        host.current.persistentKeepalive = template.persistentKeepalive
                    }
                    this.host.default = this.meshes[i].default
                } 
            host.current.preshareKey = this.host.default.preshareKey
            }
        }

        // check allowed IPs
        if (host.current.allowedIPs.length < 1) {
          this.errorhost('Please provide at least one valid CIDR address for host allowed IPs');
          return;
        }
        for (let i = 0; i < host.current.allowedIPs.length; i++){
          if (this.$isCidr(host.current.allowedIPs[i]) === 0) {
            this.errorhost('Invalid CIDR detected, please correct before submitting');
            return
          }
        }
        // check address
        if (host.current.address.length < 1) {
          this.errorhost('Please provide at least one valid CIDR address for host');
          return;
        }
        for (let i = 0; i < host.current.address.length; i++){
          if (this.$isCidr(host.current.address[i]) === 0) {
            this.errorhost('Invalid CIDR detected, please correct before submitting');
            return
          }
        }
        // all good, submit
        this.dialogUpdate = false;
        this.dialogServiceHost = false;
        this.updatehost(host)
      },

      forceFileDownload(host){
        let config = this.gethostConfig(host.id)
        if (!config) {
          this.errorhost('Failed to download host config');
          return
        }
        const url = window.URL.createObjectURL(new Blob([config]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', host.name.split(' ').join('-') + '-' + host.meshName.split(' ').join('-') + '.zip') //or any other extension
        document.body.appendChild(link)
        link.click()
      },
    }
  };
</script>
