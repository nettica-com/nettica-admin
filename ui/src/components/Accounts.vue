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
                Account
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
                        @click="startInvite"
                >
                    Invite member
                    <v-icon right dark>mdi-account-group</v-icon>
                </v-btn>
            </v-card-title>
        </v-card>
        <v-card>
            <v-card-title>
                Accounts
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table
                    v-if="listView"
                    :headers="headers"
                    :items="accounts"
                    :search="search"
                     @click:row="startUpdate"
            >
                <template v-slot:item.updated="{ item }">
                    <v-row>
                        <p>At {{ item.updated | formatDate }} by {{ item.updatedBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.action="{ item }">
                    <v-row>
<!--
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                        >
                            mdi-account-switch
                        </v-icon>
-->
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                                title = "Edit   "
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
        <v-card>
            <v-card-title>
                Members
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table
                    v-if="listView"
                    :headers="bottom_headers"
                    :items="members"
                    :search="search"
                    @click:row="startUpdateMember"
            >
                <template v-slot:item.updated="{ item }">
                    <v-row>
                        <p>At {{ item.updated | formatDate }} by {{ item.updatedBy }}</p>
                    </v-row>
                </template>
                <template v-slot:item.action="{ item }">
                    <v-row>
<!--                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                        >
                            mdi-account-switch
                        </v-icon>
-->
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdateMember(item)"
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
                v-if="account"
                v-model="dialogCreate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Invite new member</v-card-title>
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
                                <v-text-field
                                        v-model="account.name"
                                        label="Name"
                                        :rules="[ v => !!v || 'Name is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="toAddress"
                                        label="Enter the email address of user you'd like to invite"
                                        :rules="[ v => !!v || 'Email address is required', ]"
                                        required
                                />
                                <v-switch
                                        v-model="sendEmail"
                                        color="success"
                                        inset
                                        label="Send Email"
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
                            @click="create(toAddress, meshList.selected)"
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
                v-if="user"
                v-model="dialogUpdate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Edit User</v-card-title>
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
                                        v-model="user.accountName"
                                        label="Account Name"
                                        :rules="[ v => !!v || 'Account name is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="user.email"
                                        label="Email Address"
                                        :rules="[ v => !!v || 'Email address is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="user.name"
                                        label="Name"
                                        :rules="[ v => !!v || 'User name is required',]"
                                        required
                                />
                                <v-select
                                    :items="roles"
                                    v-model="user.role"
                                    label="Role"
                                ></v-select>
                                <v-select
                                    :items="statuses"
                                    v-model="user.status"
                                    label="Status"
                                ></v-select>
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                    <v-card-actions>
                        <v-btn
                                :disabled="!valid"
                                color="success"
                                @click="update(user)"
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
                v-if="member"
                v-model="dialogMember"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Edit Member</v-card-title>
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
                                        v-model="member.accountName"
                                        label="Account Name"
                                        :rules="[ v => !!v || 'Account name is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="member.email"
                                        label="Email Address"
                                        :rules="[ v => !!v || 'Email address is required',]"
                                        required
                                />
                                <v-text-field
                                        v-model="member.name"
                                        label="Name"
                                        :rules="[ v => !!v || 'Name is required',]"
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
                                <v-select
                                    :items="roles"
                                    v-model="member.role"
                                    label="Role"
                                ></v-select>
                                <v-select
                                    :items="statuses"
                                    v-model="member.status"
                                    label="Status"
                                ></v-select>
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                    <v-card-actions>
                        <v-btn
                                :disabled="!valid"
                                color="success"
                                @click="updateMember(member)"
                        >
                            Submit
                            <v-icon right dark>mdi-check-outline</v-icon>
                        </v-btn>
                        <v-btn
                                color="primary"
                                @click="dialogMember = false"
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
    name: 'Accounts',

    data: () => ({
      listView: true,
      dialogCreate: false,
      dialogUpdate: false,
      dialogMember : false,
      inDelete: false,
      meshList: {},
      toAddress: "",
      sendEmail: true,
      roles : ["Owner", "Admin", "User"],
      statuses : ["Active", "Pending", "Suspended", "Hidden"],
      user: null,
      member: null,
      account: null,
      panel: 1,
      valid: false,
      search: '',
      headers: [
        { text: 'Account Name', value: 'accountName', },
        { text: 'Name', value: 'name', },
        { text: "Role", value: 'role', },
        { text: 'Mesh', value: 'meshName', },
        { text: 'From', value: 'from', },
        { text: 'Status', value: 'status', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
      bottom_headers: [
        { text: 'Email', value: 'email', },
        { text: 'Name', value: 'name', },
        { text: "Role", value: 'role', },
        { text: 'Mesh', value: 'meshName', },
        { text: 'Account Name', value: 'accountName', },
        { text: 'Status', value: 'status', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
    }),

    computed:{
      ...mapGetters({
        authuser: 'auth/user',
        accounts: 'account/accounts',
        members: 'account/users',
        meshes: 'mesh/meshes',
      }),
    },

    mounted () {
      this.readAllAccounts(this.authuser.email)
      this.readAllMeshes()
    },

    watch: {
      // whenever accounts changes, this function will run
      accounts(newAccounts, oldAccounts) {
          for (let i=0; i<newAccounts.length; i++) {
            if (newAccounts[i].id == newAccounts[i].parent ) {
                this.readUsers(newAccounts[i].id);
            }
        }      
      }
    },

    methods: {
        ...mapActions('account', {
            readAllAccounts: 'readAll',
            readUsers: 'readUsers',
            createAccount: 'create',
            updateAccount: 'update',
            delete: 'delete',
            emailUser: 'email',
        }),

        ...mapActions('mesh', {
            readAllMeshes: 'readAll',
        }),

      Refresh() {
        this.readAllAccounts(this.authuser.email)
        this.readAllMeshes()
      },


      startInvite() {
        this.dialogCreate = true;
        this.account = {
          name: "",
          from: this.authuser.email,
          email: "",
        }
        this.meshList = { selected: { "text": "",  "value": ""},
                          items: [] }

        var selected = 0;
        this.meshList.items[0] = { "text": "All Meshes", "value": ""}
        for (let i=0; i<this.meshes.length; i++) {
            this.meshList.items[i+1]= { "text": this.meshes[i].meshName, "value": this.meshes[i].id }
        }

        this.meshList.selected = this.meshList.items[selected];

      },

      create(toAddress, mesh) {
        this.account.email = toAddress;
        this.account.meshId = mesh.value;
        this.account.meshName = mesh.text;
        this.account.from = this.authuser.email;
        this.account.role = "User"
        this.account.status = "Pending"

        for (let i=0; i<this.accounts.length; i++) {
            if (this.accounts[i].id == this.accounts[i].parent) {
                this.account.parent = this.accounts[i].id;
                this.account.accountName = this.accounts[i].accountName;
                break;        
            }
        }

        var result = this.createAccount(this.account)
        console.log( "result = %s", result)

        if ((result) && (this.sendEmail)) {
            this.emailUser(result)
        }

        this.dialogCreate = false;

      },

      remove(item) {
        this.inDelete = true;
        if (item.role == "Owner") {
            alert("You cannot delete owners")
        } else if (confirm(`Do you really want to delete ${item.name} ?`)){
          this.delete(item)
        }
        this.readAllAccounts(this.authuser.email)
        this.readAllMeshes()

      },

      email(account) {
        this.dialogCreate = false;
        if ( account.Email == "") {
          this.errorUser('email address is not defined')
          return
        }

        this.emailUser(account)

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

      forceFileDownload(user){
        let config = this.getUserConfig(user.userid)
        if (!config) {
          this.errorUser('Failed to download user config');
          return
        }
        const url = window.URL.createObjectURL(new Blob([config]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', user.name.split(' ').join('-') + '.conf') //or any other extension
        document.body.appendChild(link)
        link.click()
      },
    }
  };
</script>
