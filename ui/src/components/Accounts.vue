<template>
    <v-container style="padding-top:0px">
        <v-snackbar v-model="notification.show" :center="true" :bottom="true" :color="notification.color">
            <v-row>
                <v-col cols="9" class="text-center">
                    {{ notification.text }}
                </v-col>
                <v-col cols="3">
                    <v-btn text @click="notification.show = false">close</v-btn>
                </v-col>
            </v-row>
        </v-snackbar>
        <v-row><v-col cols="12">
        <div>
            <v-btn class="mb-3 mt-0" @click="Refresh()">
                <v-icon dark>mdi-refresh</v-icon>
                Refresh
            </v-btn>
        </div>
        <v-card>
            <v-card-title>
                Accounts
                <v-spacer></v-spacer>
                <v-text-field v-model="search" append-icon="mdi-magnify" label="Search" single-line
                    hide-details></v-text-field>
                <v-spacer></v-spacer>
                <v-btn color="success" @click="startInvite">
                    Invite
                    <v-icon right dark>mdi-account-group</v-icon>
                </v-btn>&nbsp;
                <v-btn color="primary" @click="dialogAPI = true">
                    API
                    <v-icon right dark>mdi-key</v-icon>
                </v-btn>
            </v-card-title>
        </v-card>
        <v-card>
            <v-row>
                <v-col cols="6">
                    <v-treeview ref="tree" v-if="showTree" :items="items" :search="search" :active.sync="active"
                        :open.sync="open" activatable open-all hoverable>
                        <template v-slot:prepend="{ item }">
                            <span v-if="item.symbol" class="material-symbols-outlined">{{ item.symbol }}</span>
                            <v-icon v-else>
                                {{ item.icon }}
                            </v-icon>
                        </template>
                    </v-treeview>
                </v-col>
                <v-divider vertical></v-divider>
                <v-col cols="6" class="text-center">
                    <div v-if="!selected" class="text-h6 grey--text text--lighten-1 font-weight-light"
                        style="align-self: center;">
                    </div>
                    <v-card v-else-if="selected.isMember" :key="selected.id" class="px-3 mx-auto" style="align-self: center;" flat>
                        <v-card-text width="550">
                            <v-avatar v-if="selected.member.picture != ''" size="50">
                                <img :src="selected.member.picture" class="mx-auto d-block" width="50" height="50" />
                            </v-avatar>
                            <v-icon v-else size="50" >mdi-account</v-icon>

                            <h3 class="text-h5 mb-2">
                                {{ selected.name }}
                            </h3>
                            <h5 class="text-h6 mb-2">
                                {{ selected.role }}
                            </h5>
                        </v-card-text>
                        <v-divider></v-divider>

                        <v-row class="px-3" width="600">
                            <v-col flex>
                                <v-form ref="form" v-model="valid">
                                    <v-text-field v-model="selected.member.parent" label="Account ID" readonly />
                                    <v-text-field v-model="selected.member.accountName" label="Account Name"
                                        :rules="[v => !!v || 'Account name is required',]" required />
                                    <v-text-field v-model="selected.member.name" label="Name"
                                        :rules="[v => !!v || 'Name is required',]" required />
                                    <v-text-field v-model="selected.member.email" label="Email Address"
                                        :rules="[v => !!v || 'Email address is required',]" required >
                                        <template v-slot:append>
                                                    <v-btn icon @click="resendEmail(selected.member)" >
                                                        <v-icon dark>mdi-refresh</v-icon>
                                                        <v-icon dark>mdi-email-outline</v-icon>
                                                    </v-btn>
                                        </template>
                                    </v-text-field>
                                    <v-text-field v-model="selected.member.picture" label="Picture" />
                                    <v-select  :items="networks" v-model="selected.netName" label="To this net" 
                                        :readonly="selected.isReadOnly" ></v-select>
                                    <v-select :items="roles" v-model="selected.role" label="Role" :readonly="selected.isReadOnly"></v-select>
                                    <v-select :items="statuses" v-model="selected.status" label="Status"></v-select>
                                    <p class="text-caption">Created by {{ selected.member.createdBy }} at {{ selected.member.created | formatDate }}<br/>
                                                            Last update by {{ selected.member.updatedBy }} at {{ selected.member.updated | formatDate }}</p>

                                </v-form>
                            </v-col>
                        </v-row>
                        <v-card-actions>
                            <v-btn color="success" @click="updateMember(selected)">
                                Save
                                <v-icon right dark>mdi-check-outline</v-icon>
                            </v-btn>
                            <v-spacer></v-spacer>
                            <v-btn color="error" @click="remove(selected.member)">
                                Delete
                                <v-icon right dark>mdi-delete-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-col>
            </v-row>
        </v-card>
        </v-col></v-row>
        <v-dialog v-if="account" v-model="dialogCreate" max-width="550">
            <v-card>
                <v-card-title class="headline">Invite new member</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-select return-object v-model="netList.selected" :items="netList.items" item-text="text"
                                    item-value="value" label="To this net" :rules="[v => !!v || 'Net is required',]" single
                                    persistent-hint required />
                                <v-text-field v-model="account.name" label="Name" :rules="[v => !!v || 'Name is required',]"
                                    required />
                                <v-text-field v-model="toAddress"
                                    label="Enter the email address of user you'd like to invite"
                                    :rules="[v => !!v || 'Email address is required',]" required />
                                <v-switch v-model="sendEmail" color="success" inset label="Send Email" />
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="!valid" color="success" @click="create(toAddress, netList.selected)">
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
        <v-dialog v-if="accounts" v-model="dialogAPI" max-width="650">
            <v-card>
                <v-card-title class="headline">API Keys</v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12">
                            <v-form ref="form" v-model="valid">
                                <v-data-table :headers="kheaders" :items="accounts" :items-per-page="5"
                                    class="elevation-1">
                                    <template v-slot:item.action="{ item }">
                                        <v-row>
                                            <v-icon class="pr-1 pl-1" @click="regenerateKey(item)"
                                                title="Regenerate API Key (immediate change)">
                                                mdi-refresh
                                            </v-icon>
                                        </v-row>
                                    </template>
                                </v-data-table>
                            </v-form>
                        </v-col>
                    </v-row>
                </v-card-text>
                <v-card-actions>
                    <v-btn color="primary" @click="dialogAPI = false">
                        OK
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
        notification: {},
        showTree: false,
        items: [],
        inEdit: false,
        open: [],
        active: [],

        dialogCreate: false,
        dialogAPI: false,
        inDelete: false,
        netList: {
            items: [],
            selected: null,
        },
        toAddress: "",
        sendEmail: true,
        networks: ["All Networks"],
        roles: ["Owner", "Admin", "User"],
        statuses: ["Active", "Pending", "Suspended", "Hidden"],
        allusers: [],
        user: null,
        member: null,
        account: null,
        panel: 1,
        valid: false,
        search: '',
        kheaders: [
            { text: 'Account', value: 'accountName', },
            { text: 'API Key', value: 'apiKey', },
            { text: 'Actions', value: 'action', sortable: false, },
        ],
    }),

    computed: {
        selected() {

            let findValue = (arr, val) => {
                for(let obj of arr){
                    if (obj.id === val) {
                        return obj;
                    }
                    if(obj.children){
                        let result = findValue(obj.children, val);
                        if (result) {
                            return result;
                        }
                    }
                }
                return undefined;
            };

            if (!this.active.length) return undefined

            const id = this.active[0]
            console.log("selected item = ", findValue(this.items, id))

            return findValue(this.items, id)
        },

        ...mapGetters({
            authuser: 'auth/user',
            create_result: 'account/account',
            accounts: 'account/accounts',
            members: 'account/members',
            nets: 'net/nets',
            getMembers: 'account/getMembers',
            accountError: 'account/error',
        }),
    },

mounted() {
        this.readAllNetworks()
        this.readAllAccounts(this.authuser.email)

    },

    watch: {
        // whenever accounts changes, this function will run
        accounts(newAccounts, oldAccounts) {
            for (let i = 0; i < newAccounts.length; i++) {
                this.readMembers(newAccounts[i].parent);
            }
        },
        members(newMembers, oldMembers) {
            this.buildTree()
        },
        nets(newNets, oldNets) {
            this.netList.items[0] = { "text": "All Networks", "value": "" }
            this.networks = ["All Networks"]
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i + 1] = { "text": this.nets[i].netName, "value": this.nets[i].id }
                this.networks[i + 1] = this.nets[i].netName
            }
        },
        accountError(newError, oldError) {
            if (newError != "") {
                this.notification = {
                    show: true,
                    text: newError,
                    timeout: 10000,
                    color: "error",
                }
            } else {
                this.notification = {
                    show: true,
                    text: "Changes saved.",
                    timeout: 2000,
                }
            }
        },
    },

    methods: {
        ...mapActions('account', {
            readAllAccounts: 'readAll',
            readUsers: 'readUsers',
            readMembers: 'readMembers',
            createAccount: 'create',
            updateAccount: 'update',
            delete: 'delete',
            emailUser: 'email',
        }),

        ...mapActions('net', {
            readAllNetworks: 'readAll',
        }),

        findValue(arr, val) {
            for(let obj of arr){
                if (obj.id === val) {
                    return obj;
                }
                if(obj.children){
                    let result = findValue(obj.children, val);
                    if (result) {
                        return result;
                    }
                }
            }
            return undefined;
        },

        Refresh() {
            this.readAllAccounts(this.authuser.email)
            this.readAllNetworks()
            this.buildTree()
        },

        buildTree() {
            // build the treeview using the accounts and users/members
            this.items = []
            for (let i = 0; i < this.accounts.length; i++) {
                this.items[i] = {
                    id: this.accounts[i].parent,
                    name: this.accounts[i].accountName,
                    account: this.accounts[i],
                    status: this.accounts[i].status,
                    icon: "mdi-account-group",
                    isAccount: true,
                    children: []
                }
                if (this.accounts[i].id != this.accounts[i].parent) {
                    var name = this.accounts[i].name
                        if (this.accounts[i].netName != "") {
                            name = name + " (" + this.accounts[i].netName + ")"
                        }
                        var netName = this.accounts[i].netName
                        if (netName == "") {
                            netName = "All Networks"
                        }

                    this.items[i].children[0] = {
                        id: "m-" + this.accounts[i].id,
                        name: name,
                        netName: netName,
                        isReadOnly: true,
                        member: this.accounts[i],
                        status: this.accounts[i].status,
                        role: this.accounts[i].role,
                        icon: "mdi-account",
                        isAccount: false,
                        isMember: true,
                        children: []
                    }
                }
                console.log("account: ", this.accounts[i])
                var k = 0
                var members = this.getMembers(this.accounts[i].parent)

                if (members == null || members == undefined) {
                    continue
                }

                for (let j = 0; j < members.length; j++) {
                    if (members[j].parent == this.accounts[i].parent) {
                        var name = members[j].name
                        if (members[j].netName != "") {
                            name = name + " (" + members[j].netName + ")"
                        }
                        var netName = members[j].netName
                        if (netName == "") {
                            netName = "All Networks"
                        }
                        var netList =  {
                                selected: { "text": "", "value": "" },
                                items: []
                            }

                        netList.items[0] = { "text": netName, "value": members[j].netId }
                        netList.selected = { "text": netName, "value": members[j].netId }
                        this.items[i].children[k] = {
                            id: "m-" + members[j].id + "-" + members[j].netName,
                            name: name,
                            member: members[j],
                            status: members[j].status,
                            netName: netName,
                            isReadOnly: false,
                            role: members[j].role,
                            email: members[j].email,
                            icon: "mdi-account",
                            isAccount: false,
                            isMember: true,
                            children: []
                        }
                        k++
                    }
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

            this.showTree = true
            return this.items

        },


        startInvite() {
            this.dialogCreate = true;
            this.account = {
                name: "",
                from: this.authuser.email,
                email: "",
            }
            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = 0;
            this.netList.items[0] = { "text": "All Networks", "value": "" }
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i + 1] = { "text": this.nets[i].netName, "value": this.nets[i].id }
            }

            this.netList.selected = this.netList.items[selected];

        },

        async create(toAddress, net) {
            this.account.email = toAddress;
            this.account.netId = net.value;
            this.account.netName = net.text;
            if (this.account.netName == "All Networks") {
                this.account.netName = ""
                this.account.netId = ""
            }
            this.account.from = this.authuser.email;
            this.account.role = "User"
            this.account.status = "Pending"

            for (let i = 0; i < this.accounts.length; i++) {
                if (this.accounts[i].id == this.accounts[i].parent) {
                    this.account.parent = this.accounts[i].id;
                    this.account.accountName = this.accounts[i].accountName;
                    break;
                }
            }

            var result = await this.createAccount(this.account)

            console.log("result = ", result)
            // sleep one second
            
            await new Promise(r => setTimeout(r, 1000));

            console.log("create_result = ", this.create_result)

            if ((this.create_result) && (this.sendEmail)) {
                this.emailUser(this.create_result)
                this.notification = {
                    show: true,
                    text: "Email sent to " + this.create_result.email,
                    timeout: 5000,
                }
            } else {
                this.notification = {
                    show: true,
                    text: "User created.",
                    timeout: 2000,
                }
            }

            this.dialogCreate = false;
            this.Refresh()


        },

        resendEmail(account) {
            this.emailUser(account)
            this.notification = {
                show: true,
                text: "Email sent to " + account.email,
                timeout: 5000,
            }
        },

        remove(item) {
            this.inDelete = true;
            if (item.role == "Owner") {
                alert("You cannot delete owners")
            } else if (confirm(`Do you really want to delete ${item.name} ?`)) {
                this.delete(item)
            }
            this.readAllAccounts(this.authuser.email)
            this.readAllNetworks()

        },

        email(account) {
            this.dialogCreate = false;
            if (account.Email == "") {
                this.errorUser('email address is not defined')
                return
            }

            this.emailUser(account)

        },

        startUpdate(user) {
            if (this.inDelete == true) {
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

        updateMember(item) {

            this.dialogMember = false;

            console.log( "updateMember: ", item)
                    
            item.member.netName = item.netName;

            if (item.member.netName == "All Networks") {
                item.member.netName = "";
                item.member.netId = ""; 
            }

            item.member.role = item.role;
            item.member.status = item.status;

            console.log( "nets = ", this.nets)
            console.log( "netList = ", this.netList)

            for (let i = 0; i < this.nets.length; i++) {
                if (this.nets[i].netName == item.member.netName) {
                    item.member.netId = this.nets[i].id;
                    break;
                }
            }

            var name = item.member.name
            if (item.member.netName != "") {
                name = name + " (" + item.member.netName + ")"
            }
            item.name = name           

            console.log( "updateAccount: ", item.member)
            this.updateAccount(item.member)
            this.readAllAccounts(this.authuser.email)
            this.buildTree()

            this.notification = {
                show: true,
                text: "Changes saved.",
                timeout: 2000,
            }

        },

        regenerateKey(item) {
            if (confirm(`Do you really want to regenerate the API key for ${item.accountName} ?`)) {
                item.apiKey = "";
                this.updateAccount(item)
                // We have to force a refresh here because the API key
                // updated to blank in the vuex store
                this.Refresh()
                this.readAllAccounts(this.authuser.email)
                this.notification = {
                    show: true,
                    text: "API key regenerated.",
                    timeout: 2000,
                }
            }
        },

        forceFileDownload(user) {
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
