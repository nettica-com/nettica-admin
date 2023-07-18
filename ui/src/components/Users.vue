<template>
    <v-container>
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
                        @click="startCreate"
                >
                    Invite member
                    <v-icon right dark>mdi-account-group</v-icon>
                </v-btn>
            </v-card-title>
        </v-card>
        <v-card>
            <v-card-title>
                Users
                <v-spacer></v-spacer>
                <v-spacer></v-spacer>
            </v-card-title>
            <v-data-table
                    v-if="listView"
                    :headers="headers"
                    :items="accounts"
                    :search="search"
            >
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
                        >
                            mdi-account-switch
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click.stop="startUpdate(item)"
                        >
                            mdi-square-edit-outline
                        </v-icon>
                        <v-icon
                                class="pr-1 pl-1"
                                @click="remove(item)"
                        >
                            mdi-trash-can-outline
                        </v-icon>
                    </v-row>
                </template>

            </v-data-table>
            <v-card-text v-else>
                <v-row>
                    <v-col
                            v-for="(user, i) in users "
                            :key="i"
                            sm12 lg6
                    >
                        <v-card
                                :color="user.enable ? '#1F7087' : 'warning'"
                                class="mx-auto"
                                raised
                                shaped
                        >
                            <v-list-item>
                                <v-list-item-content>
                                    <v-list-item-title class="headline">{{ user.name }}</v-list-item-title>
                                    <v-list-item-subtitle>{{ user.email }}</v-list-item-subtitle>
                                </v-list-item-content>

                                <v-list-item-avatar
                                        tile
                                        size="150"
                                >
                                </v-list-item-avatar>
                            </v-list-item>
                            <v-card-actions>
                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                v-on:click="forceFileDownload(user)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Download</span>
                                            <v-icon right dark>mdi-cloud-download-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Download</span>
                                </v-tooltip>

                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                @click.stop="startUpdate(user)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Edit</span>
                                            <v-icon right dark>mdi-square-edit-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Edit</span>
                                </v-tooltip>

                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                @click="remove(user)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Delete</span>
                                            <v-icon right dark>mdi-trash-can-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Delete</span>
                                </v-tooltip>

                                <v-tooltip bottom>
                                    <template v-slot:activator="{ on }">
                                        <v-btn
                                                text
                                                @click="email(user)"
                                                v-on="on"
                                        >
                                            <span class="d-none d-lg-flex">Send Email</span>
                                            <v-icon right dark>mdi-email-send-outline</v-icon>
                                        </v-btn>
                                    </template>
                                    <span>Send Email</span>
                                </v-tooltip>
                                <v-spacer/>
                                <v-tooltip right>
                                    <template v-slot:activator="{ on }">
                                        <v-switch
                                                dark
                                                v-on="on"
                                                color="success"
                                                v-model="user.enable"
                                                v-on:change="update(user)"
                                        />
                                    </template>
                                    <span> {{user.enable ? 'Disable' : 'Enable'}} this user</span>
                                </v-tooltip>

                            </v-card-actions>
                        </v-card>
                    </v-col>
                </v-row>
            </v-card-text>
        </v-card>
        <v-dialog
                v-if="user"
                v-model="dialogCreate"
                max-width="550"
        >
            <v-card>
                <v-card-title class="headline">Invite new user</v-card-title>
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
                                        v-model="toAddress"
                                        label="Enter the email address of user you'd like to invite"
                                        :rules="[ v => !!v || 'Email address is required', ]"
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
                            @click="email(toAddress)"
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
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
                    <v-spacer/>
                    <v-row>
                        <v-col cols="12">

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
                        </v-col>
                    </v-row>


        </v-dialog>
    </v-container>
</template>
<script>
  import { mapActions, mapGetters } from 'vuex'

  export default {
    name: 'Users',

    data: () => ({
      listView: true,
      dialogCreate: false,
      dialogUpdate: false,
      toAddress: "",
      user: null,
      panel: 1,
      valid: false,
      search: '',
      headers: [
        { text: 'Email', value: 'email', },
        { text: 'Name', value: 'name', },
        { text: "Role", value: 'role', },
        { text: 'ID', value: 'id', },
//        { text: 'Tags', value: 'tags', },
        { text: 'Actions', value: 'action', sortable: false, },

      ],
    }),

    computed:{
      ...mapGetters({
        authuser: 'auth/user',
        server: 'server/server',
        users: 'user/users',
        accounts: 'account/accounts',
      }),
    },

    mounted () {
      this.readAllAccounts(this.authuser.email)
//      this.readAllUsers()
//      this.readServer()
    },

    methods: {
      ...mapActions('user', {
        errorUser: 'error',
        readAllUsers: 'readAll',
        createUser: 'create',
        updateUser: 'update',
        deleteUser: 'delete',
        emailUser: 'email',
      }),

    ...mapActions('server', {
            readServer: 'read',
        }),
    ...mapActions('account', {
            readAllAccounts: 'readAll'

      }),

      startCreate() {
        this.user = {
          name: "",
          email: this.authuser.email,
          toAddress: "",
          enable: true,
          userName: "",
          userid: "",
          tags: [],
        }
        this.dialogCreate = true;
      },

      create(user) {
        this.dialogCreate = false;
        this.createUser(user)
      },

      remove(user) {
        if(confirm(`Do you really want to delete ${user.name} ?`)){
          this.deleteUser(user)
        }
      },

      email(toAddress) {
        this.dialogCreate = false;
        if (!toAddress) {
          this.errorUser('email address is not defined')
          return
        }

        this.emailUser(toAddress + "/" + this.accounts[0].id)
      },

      startUpdate(user) {
        this.user = user;
        this.dialogUpdate = true;
      },

      update(user) {

        this.dialogUpdate = false;
        user.id = user.email
        this.updateUser(user)
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
