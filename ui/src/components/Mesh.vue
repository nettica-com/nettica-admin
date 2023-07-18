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
                Meshes
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
                    Create new mesh
                    <img class="ml-1" :src="require('../assets/nettica.png')" height="32" width="32" alt="nettica"/>
                </v-btn>
            </v-card-title>
            <d3-network class="network" :net-nodes="nodes" :net-links="links" :options="options" />
            <v-data-table
                    v-if="listView"
                    :headers="headers"
                    :items="meshes"
                    :search="search"
                    no-data-text="Welcome to Nettica!  Click 'Create New Network' above to get started."
                    no-results-text="No results matching your search"
                     @click:row="loadNetwork"
            >
                <template v-slot:item.default.address="{ item }">
                    <v-chip
                            v-for="(ip, i) in item.default.address"
                            :key="i"
                            color="#336699"
                            text-color="white"
                    >
                        <v-icon left>mdi-ip-network</v-icon>
                        {{ ip }}
                    </v-chip>
                </template>
                <template v-slot:item.default.tags="{ item }">
                    <v-chip
                            v-for="(tag, i) in item.default.tags"
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
                    <v-row>
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                                title = "Edit"
                        >
                            mdi-square-edit-outline
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click="remove(item)"
                                title = "Delete"
                        >
                            mdi-trash-can-outline
                        </v-icon>
                    </v-row>
                </template>

            </v-data-table>
        </v-card>
        <v-dialog
                v-if="mesh"
                v-model="dialogCreate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Create New Mesh</v-card-title>
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
                                        v-model="mesh.meshName"
                                        label="Mesh friendly name"
                                        :rules="[ v => !!v || 'Mesh name is required', ]"
                                        required
                                />
                                <v-text-field
                                    v-model="mesh.description"
                                    label="Description"
                                />
                                <v-select return-object
                                        v-model="acntList.selected"
                                        :items="acntList.items"
                                        item-text = "text"
                                        item-value = "value"
                                        label="For this account"
                                        :rules="[ v => !!v || 'Account is required', ]"
                                        single
                                        persistent-hint
                                />


                                <v-combobox
                                        v-model="mesh.default.address"
                                        :items="mesh.default.address"
                                        label="IP subnet for this mesh (ex. 10.10.10.0/24)"
                                        :rules="[ v => !!v || 'Network is required', ]"
                                        multiple
                                        chips
                                        persistent-hint
                                        required
                                />
                                <v-combobox
                                    v-model="mesh.default.dns"
                                    chips
                                    hint="Enter the IP address of a global DNS resolver, hit tab, hit enter."
                                    label="DNS servers for this mesh"
                                    multiple
                                    dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="mesh.default.dns.splice(mesh.default.dns.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                        v-model="mesh.default.tags"
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
                                                @click:close="mesh.default.tags.splice(mesh.default.tags.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-switch
                                        v-model="mesh.default.upnp"
                                        color="success"
                                        inset
                                        label="Enable UPnP where possible"
                               />
                                    <v-switch
                                        v-model="mesh.default.enableDns"
                                        color="success"
                                        inset
                                        label="Enable Meshify DNS"
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
                            @click="create(mesh)"
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
                v-if="mesh"
                v-model="dialogUpdate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Edit Mesh</v-card-title>
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
                                        v-model="mesh.id"
                                        label="Id"
                                        :rules="[ v => !!v || 'Mesh id is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="mesh.meshName"
                                        label="Friendly name"
                                        :rules="[ v => !!v || 'Mesh name is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="mesh.description"
                                        label="Description"
                                />
                                <v-combobox
                                        v-model="mesh.default.address"
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
                                                @click:close="mesh.default.address.splice(mesh.default.address.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                    v-model="mesh.default.dns"
                                    chips
                                    hint="Write IP address(es) and hit enter or leave empty.  If not empty, be sure to include your local resolver."
                                    label="DNS servers for this mesh"
                                    multiple
                                    dark
                                >
                                    <template v-slot:selection="{ attrs, item, select, selected }">
                                        <v-chip
                                                v-bind="attrs"
                                                :input-value="selected"
                                                close
                                                @click="select"
                                                @click:close="mesh.default.dns.splice(mesh.default.dns.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>
                                <v-combobox
                                        v-model="mesh.default.tags"
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
                                                @click:close="mesh.default.tags.splice(mesh.default.tags.indexOf(item), 1)"
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
            <v-expansion-panels>
            <v-expansion-panel>
                <v-expansion-panel-header dark>Advanced configuration</v-expansion-panel-header>
                    <v-expansion-panel-content>
                        <div class="d-flex flex-no-wrap justify-space-between">
                            <v-col cols="12">
                                <v-text-field
                                        v-model="mesh.default.presharedKey"
                                        label="Preshared Key"
                                />
                                <v-combobox
                                        v-model="mesh.default.allowedIPs"
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
                                                @click:close="mesh.default.allowedIPs.splice(mesh.default.allowedIPs.indexOf(item), 1)"
                                        >
                                            <strong>{{ item }}</strong>&nbsp;
                                        </v-chip>
                                    </template>
                                </v-combobox>

                                <v-text-field
                                        type="number"
                                        v-model="mesh.default.mtu"
                                        label="Define default global MTU"
                                        hint="Leave at 0 and let us take care of MTU"
                                />
                                <v-text-field
                                        type="number"
                                        v-model="mesh.default.persistentKeepalive"
                                        label="Persistent keepalive"
                                        hint="To disable, set to 0.  Recommended value 29 (seconds)"
                                />
                                <v-text-field
                                        v-model="mesh.default.listenPort"
                                        type="number"
                                        :rules="[
                                        v => !!v || 'Listen port is required',
                            ]"
                                        label="Listen port"
                                        required
                                />
                                    <v-switch
                                        v-model="mesh.default.upnp"
                                        color="success"
                                        inset
                                        label="Enable UPnP where possible"
                               />
                                    <v-switch
                                        v-model="mesh.default.enableDns"
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
                                @click="update(mesh)"
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
    </v-container>
</template>

<!-- <style src="vue-d3-network/dist/vue-d3-network.css"></style> -->
<style>
text { font-size:12px; color:white; fill:white; }
.node { fill:#336699; stroke:#5b81a7; }
.link { color: white;}
.net-svg { margin: 0 auto; }
.network { display: flex; justify-content: center;}
</style>
<script>


var D3Network = window['vue-d3-network']


  import { mapActions, mapGetters } from 'vuex'

  export default {
    name: 'Meshes',

    data: () => ({
      acntList : {},
      listView: true,
      dialogCreate: false,
      dialogUpdate: false,
      noEdit : false,
      mesh: null,
      panel: 1,
      valid: false,
      search: '',
      headers: [
        { text: 'Name', value: 'meshName', },
        { text: 'Description', value:'description'},
        { text: 'Subnet', value: 'default.address', },
        { text: 'Created', value: 'created', sortable: false, },
        { text: 'Tags', value: 'tags', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
      nodes: [
      ],
      links: [
      ],
      nodeSize:50,
      canvas:false,


    }),

    computed:{
      ...mapGetters({
        user: 'auth/user',
        server: 'server/server',
        meshes: 'mesh/meshes',
        hosts: 'host/hosts',
        accounts: 'account/accounts',

      }),
      options(){
      return{
        force: 4000,
        size:{ w:1000, h:500},
        nodeSize: this.nodeSize,
        nodeLabels: true,
        linkLabels:true,
        canvas: this.canvas
      }
    }

    },

    mounted () {
      this.readAllAccounts(this.user.email)
      this.readAllMeshes()
      this.readAllHosts()
    },

    methods: {
        ...mapActions('host', {
        readAllHosts: 'readAll',
      }),
      ...mapActions('mesh', {
        errorMesh: 'error',
        readAllMeshes: 'readAll',
        createMesh: 'create',
        updateMesh: 'update',
        deleteMesh: 'delete',
        emailMesh: 'email',
      }),
      ...mapActions('server', {
        readServer: 'read',
      }),
      ...mapActions('account', {
          readAllAccounts: 'readAll',
      }),

      Refresh() {
        this.readAllAccounts(this.user.email)
        this.readAllHosts()
        this.readAllMeshes()
      },

      loadNetwork(mesh) {
          let name = mesh.meshName
          let x = 0
          let l = 0
          this.links = []
          this.nodes = []
          let mesh_hosts = []
          for (let i=0; i<this.hosts.length; i++) {
              if (this.hosts[i].meshName == name) {
                    mesh_hosts[x] = this.hosts[i]
                    this.nodes[x] = { id: x, name: this.hosts[i].name, /* _color:'gray'*/}
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
          for (let i=0; i<mesh_hosts.length; i++) {
              for (let j=0; j<mesh_hosts.length; j++) {
                  if (i != j && mesh_hosts[j].current.endpoint != "" && mesh_hosts[j].role !="Egress") {
                      this.links[l] = { sid: i, tid: j, _color: "white"}
                      l++
                  }
              }
          }
        },

      doCopy() {
            this.$copyText(this.mesh.default.presharedKey).then(function (e) {
                alert('Copied')
                console.log(e)
            }, function (e) {
                alert('Can not copy')
                console.log(e)
            })
      },

      startCreate() {
        this.mesh = {
          name: "",
          email: this.user.email,
          enable: true,
          meshName: "",
          id: "",
          tags: [],
          accountid : ""

        }
        this.mesh.default = {
          allowedIPs: [],
          address: [],
          enableDns: false,
          upnp : false,
        }
        this.acntList = { selected: { "text": "",  "value": ""},
                          items: [] }

        var selected = 0;
        for (let i=0; i<this.accounts.length; i++) {
            this.acntList.items[i]= { "text": this.accounts[i].accountName + " - " + this.accounts[i].parent, "value": this.accounts[i].parent }
            if (this.acntList.items[i].value == this.mesh.accountid) {
                selected = i
            }
        }

        this.acntList.selected = this.acntList.items[selected];

        this.dialogCreate = true;
      },

      create(mesh) {
        this.mesh = mesh
        if (mesh.default.allowedIPs.length < 0) {
          this.errorMesh('Please provide at least one valid CIDR address for mesh allowed IPs')
          return;
        }
        for (let i = 0; i < mesh.default.allowedIPs.length; i++){
          if (this.$isCidr(mesh.default.allowedIPs[i]) === 0) {
            this.errorMesh('Invalid CIDR detected, please correct before submitting')
            return
          }
        }
        this.mesh.accountid = this.acntList.selected.value
        this.dialogCreate = false;
        this.createMesh(mesh)
      },

      remove(mesh) {
        this.noEdit = true
        if(confirm(`Do you really want to delete ${mesh.meshName}?`)){
            mesh.id = mesh.id
            this.deleteMesh(mesh)
        }
      },

      email(mesh) {
        if (!mesh.email){
          this.errorMesh('Mesh email is not defined')
          return
        }

        if(confirm(`Do you really want to send email to ${mesh.email} with all configurations ?`)){
          this.emailMesh(mesh)
        }
      },

      startUpdate(mesh) {
        if (this.noEdit == true ) {
            this.noEdit = false;
            return
        }

        this.mesh = mesh;
        this.dialogUpdate = true;
      },

      update(mesh) {

        this.mesh.default.listenPort = parseInt(this.mesh.default.listenPort, 10);
        this.mesh.default.persistentKeepalive = parseInt(this.mesh.default.persistentKeepalive, 10);
        this.mesh.default.mtu = parseInt(this.mesh.default.mtu, 10);
        this.mesh.id = mesh.id
        this.mesh.meshName = mesh.meshName


        // check allowed IPs
        if (mesh.default.allowedIPs.length < 0) {
          this.errorMesh('Please provide at least one valid CIDR address for mesh allowed IPs');
          return;
        }
        for (let i = 0; i < mesh.default.allowedIPs.length; i++){
          if (this.$isCidr(mesh.default.allowedIPs[i]) === 0) {
            this.errorMesh('Invalid CIDR detected, please correct before submitting');
            return
          }
        }
        // check address
        if (mesh.default.address.length < 1) {
          this.errorMesh('Please provide at least one valid CIDR address for mesh');
          return;
        }
        for (let i = 0; i < mesh.default.address.length; i++){
          if (this.$isCidr(mesh.default.address[i]) === 0) {
            this.errorMesh('Invalid CIDR detected, please correct before submitting');
            return
          }
        }
        // all good, submit
        this.dialogUpdate = false;
        this.updateMesh(mesh)
      },

      forceFileDownload(mesh){
        let config = this.getMeshConfig(mesh.id)
        if (!config) {
          this.errorMesh('Failed to download mesh config');
          return
        }
        const url = window.URL.createObjectURL(new Blob([config]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', mesh.name.split(' ').join('-') + '.conf') //or any other extension
        document.body.appendChild(link)
        link.click()
      },
    }
  };
</script>
