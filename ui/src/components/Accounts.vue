<template>
    <v-container style="padding-top:0px">
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
                            <v-treeview ref="tree" v-if="showTree" :items="items" :search="search" :filter="filter"
                                :active.sync="active" :open.sync="open" activatable open-all hoverable>
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
                            <v-card v-else-if="selected.hasLimits" :key="selected.pid" class="px-3 mx-auto"
                                style="align-self: center;" flat>
                                <v-card-text width="550">
                                    <v-icon right dark size="50">mdi-account-group</v-icon>
                                    <h3 class="text-h5 mb-2">
                                        {{ selected.account.accountName }}
                                    </h3>
                                    <h5 class="text-h6 mb-2">
                                        {{ selected.account.parent }}
                                    </h5>
                                </v-card-text>
                                <v-divider></v-divider>

                                <table width="100%" style="text-align: left;" >
                                    <th>Limit</th>
                                    <th>Current</th>
                                    <th>Max</th>
                                    <tr>
                                        <td>Members</td>
                                        <td>{{ selected.limits.members }}</td>
                                        <td>{{ selected.limits.maxmembers }}</td>
                                    </tr>
                                    <tr>
                                        <td>Devices</td>
                                        <td>{{ selected.limits.devices }}</td>
                                        <td>{{ selected.limits.maxdevices }}</td>
                                    </tr>
                                    <tr>
                                        <td>Networks</td>
                                        <td>{{ selected.limits.networks }}</td>
                                        <td>{{ selected.limits.maxnetworks }}</td>
                                    </tr>
                                    <tr>
                                        <td>Services</td>
                                        <td>{{ selected.limits.services }}</td>
                                        <td>{{ selected.limits.maxservices }}</td>
                                    </tr>
                                </table>
                            </v-card>
                            <v-card v-else-if="selected.isMember" :key="selected.id" class="px-3 mx-auto"
                                style="align-self: center;" flat>
                                <v-form autocomplete="off">
                                    <v-card-text width="550">
                                        <v-avatar v-if="selected.member.picture != ''" size="50">
                                            <img :src="selected.member.picture" class="mx-auto d-block" width="50"
                                                height="50" />
                                        </v-avatar>
                                        <v-icon v-else size="50">mdi-account</v-icon>

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
                                                <v-text-field v-model="selected.member.parent" label="Account ID"
                                                    readonly />
                                                <v-text-field v-model="selected.member.accountName" label="Account Name"
                                                    :rules="[v => !!v || 'Account name is required',]" required />
                                                <v-text-field v-model="selected.member.name" label="Name"
                                                    :rules="[v => !!v || 'Name is required',]" required />
                                                <v-text-field v-model="selected.member.email" label="Email Address"
                                                    :rules="[v => !!v || 'Email address is required',]" required>
                                                    <template v-slot:append>
                                                        <v-btn icon @click="resendEmail(selected.member)">
                                                            <v-icon dark>mdi-refresh</v-icon>
                                                            <v-icon dark>mdi-email-outline</v-icon>
                                                        </v-btn>
                                                    </template>
                                                </v-text-field>
                                                <v-text-field v-model="selected.member.picture" label="Picture" />
                                                <v-select :items="networks" v-model="selected.netName" label="To this net"
                                                    :readonly="selected.isReadOnly"></v-select>
                                                <v-select :items="roles" v-model="selected.role" label="Role"
                                                    :readonly="selected.isReadOnly"></v-select>
                                                <v-select :items="statuses" v-model="selected.status"
                                                    label="Status"></v-select>
                                                <p class="text-caption">Created by {{ selected.member.createdBy }} at {{
                                                    selected.member.created | formatDate }}<br />
                                                    Last update by {{ selected.member.updatedBy }} at {{
                                                        selected.member.updated | formatDate }}</p>

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
                                </v-form>
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
                                <v-select return-object v-model="acntList.selected" :items="acntList.items" item-text="text"
                                    item-value="value" label="To this account" :rules="[v => !!v || 'Account is required',]"
                                    single persistent-hint required />
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
                                <v-data-table :headers="kheaders" :items="accounts" :items-per-page="5" class="elevation-1">
                                    <template v-slot:item.action="{ item }">
                                        <v-row>
                                            <v-icon class="pr-1 pl-1" @click="regenerateKey(item)"
                                                title="Regenerate API Key (immediate change)">
                                                mdi-refresh
                                            </v-icon>
                                            <v-icon class="pr-1 pl-1" @click="copy(item.apiKey)"
                                                title="Copy API Key to clipboard">
                                                mdi-content-copy
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
        showTree: false,
        items: [],
        inEdit: false,
        open: [],
        active: [],
        acntList: {},
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
                for (let obj of arr) {
                    if (obj.id === val) {
                        return obj;
                    }
                    if (obj.children) {
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
            limits: 'account/limits',
            nets: 'net/nets',
            getMembers: 'account/getMembers',
            getLimits: 'account/getLimits',
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
                if (newAccounts[i].role == "Owner" || newAccounts[i].role == "Admin") {
                    console.log("readLimits: ", newAccounts[i].parent)
                    this.readLimits(newAccounts[i].parent);
                }
            }
            this.buildTree()
        },
        members(newMembers, oldMembers) {
            this.buildTree()
        },
        limits(newLimits, oldLimits) {
            console.log("Limits:", this.limits)
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
    },

    methods: {
        ...mapActions('account', {
            errorAccount: 'error',
            readAllAccounts: 'readAll',
            readUsers: 'readUsers',
            readMembers: 'readMembers',
            readLimits: 'readLimits',
            createAccount: 'create',
            updateAccount: 'update',
            delete: 'delete',
            emailUser: 'email',
        }),

        ...mapActions('net', {
            readAllNetworks: 'readAll',
        }),

        findValue(arr, val) {
            for (let obj of arr) {
                if (obj.id === val) {
                    return obj;
                }
                if (obj.children) {
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

        filter(item) {
            if (item.name.toLowerCase().indexOf(this.search.toLowerCase()) > -1) {
                return true;
            }
            if (item.isMember && item.netName.toLowerCase().indexOf(this.search.toLowerCase()) > -1) {
                return true;
            }
            if (item.isMember && item.member.email.toLowerCase().indexOf(this.search.toLowerCase()) > -1) {
                return true;
            }
            if (item.isAccount && item.account.email.toLowerCase().indexOf(this.search.toLowerCase()) > -1) {
                return true;
            }

            return false;
        },

        buildTree() {
            // build the treeview using the accounts and users/members
            this.items = []
            // create a root node for each parent account
            let x = 0;
            for (let i = 0; i < this.accounts.length; i++) {
                let found = false;
                for (let j = 0; j < this.items.length; j++) {
                    if (this.items[j].idx == this.accounts[i].parent) {
                        found = true;
                        break;
                    }
                }
                if (found == false) {
                    let hasLimits = false;
                    let limits = this.getLimits(this.accounts[i].parent)

                    if (limits != null && limits != undefined) {
                        hasLimits = true;
                    }

                    this.items[x] = {
                        id: "p-" + this.accounts[i].parent,
                        pid: "p-" + this.accounts[i].parent,
                        idx: this.accounts[i].parent,
                        name: this.accounts[i].accountName,
                        account: this.accounts[i],
                        status: this.accounts[i].status,
                        icon: "mdi-account-group",
                        isAccount: true,
                        hasLimits: hasLimits,
                        limits: limits,
                        children: []
                    }
                    x++;
                }
            }

            // create a child node for each account
            let child = 0;
            for (let i = 0; i < this.accounts.length; i++) {
                for (let j = 0; j < this.items.length; j++) {
                    if (this.accounts[i].parent == this.items[j].idx) {
                        // append the account to the children of the parent
                        var name = this.accounts[i].name
                        if (this.accounts[i].netName != "") {
                            name = name + " (" + this.accounts[i].netName + ")"
                        }
                        var netName = this.accounts[i].netName
                        if (netName == "") {
                            netName = "All Networks"
                        }
                        var isReadOnly = true;
                        if (this.accounts[i].id == this.accounts[i].parent) {
                            isReadOnly = false;
                        }
                        this.items[j].children[child] = {
                            id: "c-" + this.accounts[i].id,
                            idx: this.accounts[i].id,
                            name: name,
                            netName: netName,
                            isReadOnly: isReadOnly,
                            member: this.accounts[i],
                            status: this.accounts[i].status,
                            role: this.accounts[i].role,
                            icon: "mdi-account",
                            isAccount: false,
                            isMember: true,
                            children: []
                        }
                    }
                }
            }

            // now loop through the members and add them if they are not already in the tree
            for (let i = 0; i < this.items.length; i++) {

                console.log("account: ", this.accounts[i])
                var members = this.getMembers(this.accounts[i].parent)

                if (members == null || members == undefined) {
                    continue
                }
                child = this.items[i].children.length
                for (let j = 0; j < members.length; j++) {
                    if (members[j].parent == this.items[i].idx) {
                        var found = false;
                        for (let k = 0; k < this.items[i].children.length; k++) {
                            if (this.items[i].children[k].idx == members[j].id) {
                                found = true;
                                break;
                            }
                        }
                        if (found == true) {
                            continue
                        }
                        var name = members[j].name
                        if (members[j].netName != "") {
                            name = name + " (" + members[j].netName + ")"
                        }
                        var netName = members[j].netName
                        if (netName == "") {
                            netName = "All Networks"
                        }
                        var netList = {
                            selected: { "text": "", "value": "" },
                            items: []
                        }

                        netList.items[0] = { "text": netName, "value": members[j].netId }
                        netList.selected = { "text": netName, "value": members[j].netId }
                        this.items[i].children[child] = {
                            id: "m-" + members[j].id + "-" + members[j].netName,
                            idx: members[j].id,
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
                        child++
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

            // now sort the children of each parent
            for (let i = 0; i < this.items.length; i++) {
                this.items[i].children.sort((a, b) => {
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
            }

            this.showTree = true
            return this.items

        },


        startInvite() {
            this.account = {
                name: "",
                from: this.authuser.email,
                email: "",
            }
            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }
            this.acntList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = 0;
            this.netList.items[0] = { "text": "All Networks", "value": "" }
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i + 1] = { "text": this.nets[i].netName, "value": this.nets[i].id }
            }

            this.netList.selected = this.netList.items[selected];

            selected = 0;
            var x = 0;
            console.log("accounts = ", this.accounts)
            for (let i = 0; i < this.accounts.length; i++) {
                if (this.accounts[i].role == "Owner" || this.accounts[i].role == "Admin") {
                    this.acntList.items[x] = { "text": this.accounts[i].accountName + " - " + this.accounts[i].parent, "value": this.accounts[i].parent }
                    if (this.acntList.items[x].value == this.accounts[i].id) {
                        selected = x;
                    }
                    x++;
                }
            }

            this.acntList.selected = this.acntList.items[selected];

            if (x == 0) {
                this.errorAccount("You must be an admin or owner to invte a member")
                return
            }
            this.dialogCreate = true;

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
            }

            this.dialogCreate = false;
            this.Refresh()


        },

        resendEmail(account) {
            this.emailUser(account)
        },

        remove(item) {
            this.inDelete = true;
            if (item.role == "Owner") {
                alert("You cannot delete owners")
            } else if (confirm(`Do you really want to delete ${item.name} ?`)) {
                this.delete(item)
            }
            this.readAllAccounts(this.authuser.email)

        },

        email(account) {
            this.dialogCreate = false;
            if (account.Email == "") {
                this.errorAccount('email address is not defined')
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

            console.log("updateMember: ", item)

            item.member.netName = item.netName;

            if (item.member.netName == "All Networks") {
                item.member.netName = "";
                item.member.netId = "";
            }

            item.member.role = item.role;
            item.member.status = item.status;

            console.log("nets = ", this.nets)
            console.log("netList = ", this.netList)

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

            console.log("updateAccount: ", item.member)
            this.updateAccount(item.member)
            this.readAllAccounts(this.authuser.email)


        },

        regenerateKey(item) {
            if (confirm(`Do you really want to regenerate the API key for ${item.accountName} ?`)) {
                item.apiKey = "";
                this.updateAccount(item)
                // We have to force a refresh here because the API key
                // updated to blank in the vuex store
                this.Refresh()
                this.readAllAccounts(this.authuser.email)
                this.errorAccount("API key regenerated")
            }
        },

        copy(text) {
            navigator.clipboard
                .writeText(text)
                .then(() => {
                    this.errorAccount("API key copied to clipboard")
                })
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
