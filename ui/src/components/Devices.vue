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
                        <v-row>
                            <v-col cols="4">Devices</v-col>
                            <v-col cols="4">

                                <v-text-field v-model="search" append-icon="mdi-magnify" label="Search" single-line
                                    hide-details></v-text-field>
                            </v-col>
                            <v-col cols="4" class="text-right">
                                <v-btn color="success" @click="startCreate">
                                    Add Device
                                    <span class="material-symbols-outlined">devices</span>
                                </v-btn>
                            </v-col>
                        </v-row>
                    </v-card-title>
                    <div v-if="friendly">
                        <v-alert type="info" color="#336699" dismissible>
                            No devices found. <a style="color:white;" @click="startCreate">Click here to add your first device.</a>
                        </v-alert>
                    </div>

                    <v-row v-if="!friendly">
                        <v-col cols="6">
                            <v-treeview v-if="showTree" :items="items" :search="search" :filter="filter" :active.sync="active"
                                :open.sync="open" activatable hoverable>
                                <template v-slot:prepend="{ item }">
                                    <span v-if="item.symbol && item.status == 'Online'" class="material-symbols-outlined" style="color:green;">{{ item.symbol }}</span>
                                    <span v-if="item.symbol && item.status == 'Offline'" class="material-symbols-outlined" style="color:red;">{{ item.symbol }}</span>
                                    <span v-if="item.symbol && item.status == 'Native'" class="material-symbols-outlined" style="color:blue;">{{ item.symbol }}</span>
                                    <span v-if="item.symbol && !item.status" class="material-symbols-outlined">{{ item.symbol }}</span>
                                    
                                    <v-icon v-if="!item.symbol">
                                        {{ item.icon }}
                                    </v-icon>
                                </template>
                                <template v-slot:append="{ item }">
                                    <v-btn v-if="item.isDevice" icon @click="startAddVPN(item.device)">
                                        <v-tooltip bottom>
                                            <template v-slot:activator="{ on }">
                                                <v-icon v-on="on" color="#336699">mdi-plus-circle</v-icon>
                                            </template>
                                            <span>Add network to this device</span>
                                        </v-tooltip>
                                    </v-btn>
                                </template>
                            </v-treeview>
                        </v-col>
                        <v-divider vertical></v-divider>
                        <v-col cols="6" class="text-center">
                            <div v-if="!selected" class="text-h6 grey--text text--lighten-1 font-weight-light"
                                style="align-self: center;">

                            </div>
                            <v-card v-else-if="selected.isDevice" :key="selected.id" class="px-3 mx-auto" flat>
                                <v-form autocomplete="off">
                                <v-card-text width="600">
                                    <v-icon v-if="selected.device.type=='Service'" size="50">mdi-cloud</v-icon>
                                    <v-icon v-else size="50">mdi-devices</v-icon>
                                    <h3 class="text-h5 mb-2">
                                        {{ selected.name }}
                                    </h3>
                                    <div class="mb-2">
                                        <template>
                                            <v-icon v-if="selected.device.status == 'Online'"
                                                color="green">mdi-check-circle</v-icon>
                                            <v-icon v-else-if="selected.device.status == 'Native'"
                                                color="blue">mdi-minus-circle</v-icon>
                                            <v-icon v-else color="red">mdi-close-circle</v-icon>
                                            {{ selected.device.status }}
                                        </template>

                                    </div>
                                </v-card-text>
                                <v-divider></v-divider>

                                <v-row class="px-3" width="600">
                                    <v-col flex>
                                        <div :hidden="selected.device.registered">
                                            <v-text-field v-model="selected.device.ezcode" label="EZ-Code"
                                                :readonly="true" />
                                        </div>
                                        <v-text-field v-model="selected.device.description" label="Description" :readonly="!inEdit" />
                                        <v-combobox v-model="selected.device.tags" chips
                                            hint="Enter a tag, hit tab, hit enter." label="Tags" multiple dark
                                            :readonly="!inEdit">
                                            <template v-slot:selection="{ attrs, item, select }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="selected.device.tags.splice(selected.device.tags.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-select return-object v-model="selected.platform" :items="platforms.items"
                                            item-text="text" item-value="value" label="Platform of this device" single
                                            persistent-hint :readonly="!inEdit" />
                                        <v-switch v-model="selected.device.enable" color="success" inset
                                            :label="selected.device.enable ? 'Enabled' : 'Disabled'" :readonly="!inEdit" />
                                        <v-text-field v-model="selected.device.checkInterval" type="number"
                                            label="Check interval" hint="In seconds" :readonly="!inEdit" />
                                        <v-select return-object v-model="selected.accountid" :items="acntList.items"
                                            item-text="text" item-value="value" label="Account ID" single
                                            persistent-hint :readonly="!inEdit" />
                                        <v-text-field v-model="selected.device.id" label="Device ID" readonly />
                                        <v-text-field v-model="selected.device.server" label="Server" :readonly="!inEdit" />
                                        <v-text-field v-model="selected.device.apiKey" label="API Key" readonly
                                            :append-icon="showApiKey ? 'mdi-eye' : 'mdi-eye-off'"
                                            :type="showApiKey ? 'text' : 'password'"
                                            @click:append="showApiKey = !showApiKey" />
                                        <v-text-field v-model="selected.device.instanceid" label="AWS or Azure Instance ID" :readonly="!inEdit" />  
                                        <div :hidden="!inEdit">
                                            <v-text-field v-model="selected.device.name" label="Device friendly name"
                                                :rules="[v => !!v || 'device name is required',]" required />
                                        </div>
                                        <p class="text-caption">Created by {{ selected.device.createdBy }} at {{ selected.device.created | formatDate }}<br/>
                                                            Last update by {{ selected.device.updatedBy }} at {{ selected.device.updated | formatDate }}</p>
                                    </v-col>
                                </v-row>
                                <v-card-actions v-if="inEdit">
                                    <v-btn color="success" @click="updateDevice(selected)">
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
                                                <v-btn color="success" @click="copyDeviceConfig(selected.device)">
                                                    Copy
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
                                                <v-btn class="px-3" color="error" @click="removeDevice(selected.device)">
                                                    Delete
                                                    <v-icon right dark>mdi-delete-outline</v-icon>
                                                </v-btn>
                                            </v-col>
                                        </v-row>
                                    </v-container>
                                </v-card-actions>
                                </v-form>
                            </v-card>
                            <v-card v-else-if="!selected.isDevice">
                                <v-form autocomplete="off">
                                <v-card-text width="600" class="px-3">
                                    <v-icon class="material-symbols-outlined" size="50">hub</v-icon>
                                    <h3 class="text-h5 mb-2">
                                        {{ selected.vpn.netName }}
                                    </h3>
                                </v-card-text>
                                <v-divider></v-divider>
                                <v-row class="px-3" width="600">
                                    <v-col flex>
                                        <v-text-field v-model="selected.vpn.name" label="DNS name" :readonly="!inEdit"
                                          :rules="[ rules.required, rules.host ]" />
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
                                        <v-combobox v-if="selected.vpn.type=='Service'" v-model="selected.vpn.current.allowedIPs" chips
                                                hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs" multiple
                                                dark :readonly="!inEdit" >
                                            <template v-slot:selection="{ attrs, item, select }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="selected.vpn.current.allowedIPs.splice(selected.vpn.current.allowedIPs.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-text-field :readonly="!inEdit" v-model="selected.vpn.current.endpoint"
                                            label="Public endpoint for clients" :rules="[ rules.ipport ]" />
                                        <v-text-field type="number" v-model="selected.vpn.current.mtu"
                                            label="MTU" hint="Leave at 0 for auto, 1350 for IPv6 or if problems occur" />
                                        <v-switch v-model="selected.vpn.enable" color="success" inset
                                        :label="selected.vpn.enable ? 'Enabled' : 'Disabled'" :readonly="!inEdit" />
                                        <p class="text-caption">Created by {{ selected.vpn.createdBy }} at {{ selected.vpn.created | formatDate }}<br/>
                                                            Last update by {{ selected.vpn.updatedBy }} at {{ selected.vpn.updated | formatDate }}</p>

                                    </v-col>
                                </v-row>
                                <v-expansion-panels v-if="inEdit &&!(selected.vpn.type=='Service')">
                                    <v-expansion-panel>
                                        <v-expansion-panel-header dark>Advanced Configuration</v-expansion-panel-header>
                                        <v-expansion-panel-content>
                                            <div class="d-flex flex-no-wrap justify-space-between">
                                                <v-col cols="12">
                                                    <v-combobox v-model="selected.vpn.current.allowedIPs" chips
                                                        hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs"
                                                        multiple dark>

                                                        <template v-slot:selection="{ attrs, item, select }">
                                                            <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                                @click:close="selected.vpn.current.allowedIPs.splice(selected.vpn.current.allowedIPs.indexOf(item), 1)">
                                                                <strong>{{ item }}</strong>&nbsp;
                                                            </v-chip>
                                                        </template>
                                                    </v-combobox>
                                                    <v-text-field v-model="selected.vpn.id" label="VPN ID" readonly />
                                                    <v-text-field v-model="selected.vpn.netid" label="Network ID" readonly />
                                                    <v-text-field v-model="selected.vpn.deviceid" label="Device ID" disabled />
                                                    <v-combobox v-model="selected.vpn.role" :items="['', 'Ingress', 'Egress']"
                                                        label="Role" single dark />
                                                    <v-text-field v-model="selected.vpn.current.table" label="Table" />
                                                    <v-text-field v-model="selected.vpn.current.publicKey" label="Public key" />
                                                    <v-text-field v-model="selected.vpn.current.privateKey" label="Private key"
                                                        autocomplete="off"
                                                        :append-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                                                        :type="showPrivate ? 'text' : 'password'"
                                                        hint="Clear this field to have the client manage its private key"
                                                        @click:append="showPrivate = !showPrivate" />
                                                    <v-text-field v-model="selected.vpn.current.presharedKey"
                                                        label="Preshared Key" autocomplete="off"
                                                        :append-icon="showPreshared ? 'mdi-eye' : 'mdi-eye-off'"
                                                        :type="showPreshared ? 'text' : 'password'"
                                                        @click:append="showPreshared = !showPreshared" />
                                                    <v-text-field type="number"
                                                        v-model="selected.vpn.current.persistentKeepalive"
                                                        label="Persistent keepalive"
                                                        hint="To disable, set to 0.  Recommended value 29 (seconds)" />
                                                    <v-textarea v-model="selected.vpn.current.postUp" label="PostUp Script"
                                                        hint="Only applies to linux servers" />
                                                    <v-textarea v-model="selected.vpn.current.postDown" label="PostDown Script"
                                                        hint="Only applies to linux servers" />
                                                    <v-switch v-model="selected.vpn.current.subnetRouting" color="success" inset
                                                        label="Enable subnet routing" />
                                                    <v-divider></v-divider>
                                                    <table width="100%">
                                                        <tr>
                                                            <td>
                                                                <v-switch v-model="selected.vpn.current.syncEndpoint" color="success" inset
                                                                    label="Sync Endpoint" />
                                                            </td>
                                                            <td>
                                                                <v-switch v-model="selected.vpn.current.hasSSH" color="success" inset
                                                                    label="SSH" />
                                                            </td>
                                                        </tr>
                                                        <tr>
                                                            <td>
                                                                <v-switch v-model="selected.vpn.current.upnp" color="success" inset
                                                                    label="UPnP" />
                                                            </td>
                                                            <td>
                                                                <v-switch v-model="selected.vpn.current.hasRDP" color="success" inset
                                                                    label="Remote Desktop" />
                                                            </td>
                                                        </tr>
                                                        <tr>
                                                            <td>
                                                                <v-switch v-model="selected.vpn.current.failsafe" color="success" inset
                                                                    label="FailSafe" />
                                                            </td>
                                                            <td>
                                                                <v-switch v-model="selected.vpn.current.enableDns" color="success" inset
                                                                    label="Nettica DNS" />
                                                            </td>
                                                        </tr>
                                                    </table>
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
                                </v-form>
                            </v-card>
                        </v-col>
                    </v-row>

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
                        <v-chip v-for="(ip, i) in item.address" :key="i" color="#336699" text-color="white">
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
                        <v-row v-if="item.type == 'ServiceHost'">
                            <v-icon class="pr-1 pl-1" @click.stop="startUpdateServiceHost(item)" title="Edit Service">
                                mdi-square-edit-outline
                            </v-icon>
                        </v-row>
                        <v-row v-if="item.type != 'ServiceHost'">
                            <v-icon class="pr-1 pl-1" @click.stop="startUpdate(item)" title="Edit">
                                mdi-square-edit-outline
                            </v-icon>
                            <v-icon class="pr-1 pl-1" @click.stop="startCopy(item)" title="Copy">
                                mdi-content-copy
                            </v-icon>
                            <v-icon class="pr-1 pl-1" @click="remove(item)" title="Delete">
                                mdi-trash-can-outline
                            </v-icon>
                            <v-switch dark class="pr-1 pl-1" color="success" v-model="item.enable"
                                v-on:change="updateEnable(item)" />
                        </v-row>
                    </template>

                </v-card>
                <v-dialog v-if="device" v-model="dialogCreate" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Add New Device</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-text-field v-model="device.name" label="Host friendly name"
                                            :rules="[rules.required, rules.host]" required />
                                        <v-select return-object v-model="acntList.selected" :items="acntList.items" item-text="text"
                                            item-value="value" label="For this account"
                                            :rules="[v => !!v || 'Account is required',]" single persistent-hint />
                                        <v-select return-object v-model="platforms.selected" :items="platforms.items"
                                            item-text="text" item-value="value" label="Platform of this device" single
                                            persistent-hint />
                                        <v-text-field v-model="device.instanceid" label="AWS or Azure Instance ID" />
                                        <v-switch v-model="device.enable" color="success" inset
                                            :label="device.enable ? 'Enable device after creation' : 'Disable device after creation'" />
                                        <v-switch v-model="use_ezcode" color="success" inset
                                            :label="use_ezcode ? 'Use EZ-Code' : 'Do not use EZ-Code'" />

                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer />
                            <v-btn :disabled="!valid" color="success" @click="create(device)">
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
                <v-dialog v-if="device && device.ezcode" v-model="dialogEZCode" max-width="550">
                    <v-card>
                        <v-card-title class="headline">EZ-Code</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-text-field v-model="device.name" label="Host friendly name" :readonly="true" />
                                        <v-text-field v-model="device.ezcode" label="EZ-Code" :readonly="true" />
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer />
                            <v-btn color="success" @click="dialogEZCode=false;">
                                OK
                                <v-icon right dark>mdi-check-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-dialog>
                <v-dialog v-if="vpn" v-model="dialogAddVPN" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Add VPN</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-select return-object v-model="netList.selected" :items="netList.items" v-on:change="updateDefaults"
                                            item-text="text" item-value="value" label="Join this network"
                                            :rules="[v => !!v || 'Network is required',]" single persistent-hint required />
                                        <v-text-field v-model="vpn.name" label="DNS name for this device"
                                            :rules="[rules.required, rules.host]" required />
                                        <v-text-field v-model="vpn.current.endpoint" label="Public endpoint for clients"
                                            :rules="[ rules.ipport ]" />
                                        <v-switch v-model="vpn.enable" color="success" inset
                                            :label="vpn.enable ? 'Enable VPN after creation' : 'Disable VPN after creation'" />
                                        <v-switch v-model="vpn.current.syncEndpoint" color="success" inset
                                            :label="vpn.current.syncEndpoint ? 'Automatically sync endpoint using server' : 'Do not sync endpoint using server'"
                                            :disabled="!(vpn.current.endpoint && vpn.current.endpoint.length > 0)" />
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                        <v-card-actions>
                            <v-spacer />
                            <v-btn :disabled="!valid" color="success" @click="createVPN(vpn)">
                                Submit
                                <v-icon right dark>mdi-check-outline</v-icon>
                            </v-btn>
                            <v-btn color="primary" @click="dialogAddVPN = false">
                                Cancel
                                <v-icon right dark>mdi-close-circle-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-dialog>
                <v-dialog v-if="vpn" v-model="dialogServiceHost" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Manage Service: {{ vpn.name }}</v-card-title>
                        <v-card-text>

                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-combobox v-model="vpn.current.allowedIPs" chips
                                            hint="Enter IPv4 or IPv6 CIDR and press tab" label="Allowed IPs" multiple dark>

                                            <template v-slot:selection="{ attrs, item, select, selected }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="vpn.current.allowedIPs.splice(vpn.current.allowedIPs.indexOf(item), 1)">
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
                            <v-btn :disabled="!valid" color="success" @click="update(device)">
                                Submit
                                <v-icon right dark>mdi-check-outline</v-icon>
                            </v-btn>
                            <v-btn color="primary" @click="dialogServiceHost = false">
                                Cancel
                                <v-icon right dark>mdi-close-circle-outline</v-icon>
                            </v-btn>
                        </v-card-actions>
                    </v-card>

                </v-dialog>
            </v-col></v-row>
    </v-container>
</template>
<script>
import { mapActions, mapGetters } from 'vuex'


export default {
    name: 'Devices',

    data: () => ({

        acntList: {},
        showPrivate: false,
        showPreshared: false,
        showApiKey: false,
        showTree: false,
        friendly: false,
        use_ezcode: true,
        footerProps: { 'items-per-page-options': [25, 50, 100, -1] },
        dialogCreate: false,
        dialogAddVPN: false,
        dialogServiceHost: false,
        dialogEZCode: false,
        device: null,
        net: null,
        vpn: null,
        items: [],
        active: [],
        open: [],
        inEdit: false,
        name: '',
        panel: 1,
        valid: false,
        rules: {
            required: value => !!value || 'Required.',
            not_required: value => !value || (value && value.length == 0) || 'If present, must be valid IPv4 or IPv6 address and port',
            email: v => /.+@.+\..+/.test(v) || 'E-mail must be valid',
            host: v => /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/.test(v) || 'Only letters, numbers, dots and hyphens are allowed. Must start and end with a letter or number.',
            ipv4port: v => (!v || v && v.length == 0 || /^((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|^$)$/.test(v)) || 'IPv4 address and port required',
            ipv6port: v => (!v || v && v.length == 0 || /^(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5}|^$)$/.test(v)) || 'IPv6 address and port required',
            ipport: v => (!v || v && v.length == 0 || /^(((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5})|^$)(,\s+((\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}\b)|(\[([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\]:[0-9]{1,5})|^$))*)$/.test(v)) || 'If present, must be valid IPv4 or IPv6 address and port',
        },
        netList: {},
        platList: {},
        publicSubnets: false,
        platforms: {
            selected: { text: "", value: "" },
            items: [
                { text: "Windows", value: "Windows", },
                { text: "Linux", value: "Linux", },
                { text: "MacOS", value: "MacOS", },
                { text: "Apple iOS", value: "iOS", },
                { text: "Raspberry Pi", value: "raspberry"},
                { text: "Android", value: "Android", },
                { text: "Native WireGuard", value: "Native", },
            ],
        },
        search: '',
    }),

    computed: {
        selected() {
            if (!this.active.length) return undefined

            const id = this.active[0]
            console.log("selected id = ", id)

            for (let i = 0; i < this.items.length; i++) {
                if (this.items[i].id == id) {
                    return this.items[i]
                }
                for (let j = 0; j < this.items[i].children.length; j++) {
                    if (this.items[i].children[j].id == id) {
                        console.log("vpn.current.address[0] = ", this.items[i].children[j].vpn.current.address[0])
                        return this.items[i].children[j]
                    }
                }
            }
        },
        ...mapGetters({
            user: 'auth/user',
            servers: 'server/servers',
            accounts: 'account/accounts',
            devices: 'device/devices',
            nets: 'net/nets',
            vpns: 'vpn/vpns',
            deviceQrcodes: 'device/deviceQrcodes',
            getvpnconfig: "vpn/getVPNConfig",

        }),
    },

    mounted() {
        this.readAllAccounts(this.user.email)
        this.readAllDevices()

    },

    watch: {
        devices: function (val) {
            console.log("buildTree = ", this.buildTree())
            this.showTree = true
            if (this.devices.length == 0) {
                this.friendly = true
            } else {
                this.friendly = false
            }
        },

        accounts: function (val) {
            this.acntList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            for (let i = 0; i < this.accounts.length; i++) {
                this.acntList.items[i] = { "text": this.accounts[i].accountName + " - " + this.accounts[i].parent, "value": this.accounts[i].parent }
            }           
        },
        nets: function (val) {
            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            for (let i = 0; i < this.nets.length; i++) {
                // check if this net is already on the device
                if (this.device != null && this.device.vpns != null) {
                    var found = false
                    for (let j = 0; j < this.device.vpns.length; j++) {
                        if (this.device.vpns[j].netid == this.nets[i].id) {
                            found = true
                            break
                        }
                    }
                    if (found) {
                        continue
                    }
                }
                this.netList.items[i] = { "text": this.nets[i].netName, "value": this.nets[i].id }
            }           
        },
    },

    methods: {
        ...mapActions('device', {
            errorDevice: 'error',
            readAllDevices: 'readAll',
            readQrCode: 'readQrcode',
            readConfig: 'readConfig',
            createdevice: 'create',
            updatedevice: 'update',
            deletedevice: 'delete',
            updatedevice_vpn: 'update_vpn',
        }),
        ...mapActions('net', {
            readAllNetworks: 'readAll',
        }),
        ...mapActions('account', {
            readAllAccounts: 'readAll',
        }),
        ...mapActions("vpn", {
            createvpn: "create",
            readAllVPNs: "readAll",
            updatevpn: "update",
            deletevpn: "delete",
            readvpnconfig: "readConfig",
        }),

        Refresh() {
            this.readAllAccounts(this.user.email)
            this.readAllDevices()
        },

        Refreshing() {
            this.refreshing = true
            this.Refresh()

            for (let i = 0; i < 5; i++) {
                if (this.refreshing) {
                    setTimeout(() => {
                        console.log("Refreshing", i)
                        this.Refresh()
                    }, (i+1) * 2000)
                if (i == 4) {
                    this.refreshing = false
                }   
                } else {
                    break
                }
            }
        },


        async asyncRefresh() {
            await this.readAllAccounts(this.user.email)
            await this.readAllDevices()
            console.log("buildTree = ", this.buildTree())
        },
        filter(item) {
            if (this.search == "") {
                return true
            }

            if (item.name.toLowerCase().includes(this.search.toLowerCase())) {
                return true
            }
            if (item.isDevice && item.device.tags != null) {
                for (let i = 0; i < item.device.tags.length; i++) {
                    if (item.device.tags[i].toLowerCase().includes(this.search.toLowerCase())) {
                        return true
                    }
                }
            }
            if (!item.isDevice && item.vpn.current.address.length > 0) {
                for (let i = 0; i < item.vpn.current.address.length; i++) {
                    if (item.vpn.current.address[i].toLowerCase().includes(this.search.toLowerCase())) {
                        return true
                    }
                }
            }

            return false;
        },

        buildTree() {
            // build the treeview using the devices and vpns
            this.items = []
            var k = 0
            for (let i = 0; i < this.devices.length; i++) {

                this.items[i] = {
                    id: this.devices[i].id,
                    name: this.devices[i].name,
                    device: this.devices[i],
                    status: this.devices[i].status,
                    platform: { "text": this.devices[i].platform, "value": this.devices[i].platform },
                    accountid: { "text": this.devices[i].accountid, "value": this.devices[i].accountid },
                    icon: "mdi-devices",
                    symbol: "devices",
                    isDevice: true,
                    children: []
                }
 
                if (this.devices[i].type == "Service") {
                    this.items[i].icon = "mdi-cloud"
                    this.items[i].symbol = "cloud"
                }
                if (this.devices[i].vpns == null) {
                    continue
                }
                for (let j = 0; j < this.devices[i].vpns.length; j++) {
                    this.vpns[k] = this.devices[i].vpns[j]
                    k++
                    this.items[i].children[j] = {
                        id: this.devices[i].vpns[j].id,
                        name: this.devices[i].vpns[j].netName,
                        vpn: this.devices[i].vpns[j],
                        icon: "mdi-network-outline",
                        symbol: "network_node",
                        isDevice: false,
                        children: []
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

            return this.items

        },

        startCreate() {
            this.device = {
                name: "",
                accountid: "",
                email: this.user.email,
                enable: true,
                tags: [],
                current: {},
            }

            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = 0;
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i] = { "text": this.nets[i].netName, "value": this.nets[i].id }
                if (this.netList.items[i].text == this.device.netName) {
                    selected = i
                }
            }

            this.netList.selected = this.netList.items[selected];

            this.acntList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            selected = 0;
            for (let i = 0; i < this.accounts.length; i++) {
                this.acntList.items[i] = { "text": this.accounts[i].accountName + " - " + this.accounts[i].parent, "value": this.accounts[i].parent }
                if (this.acntList.items[i].value == this.device.accountid) {
                    selected = i
                }
            }

            this.acntList.selected = this.acntList.items[selected];

            this.dialogCreate = true;
        },

        make_ezcode() {
            const length = 4;
            let result = 'ez-';
            const characters = 'ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789';
            const charactersLength = characters.length;
            let counter = 0;
            while (counter < length) {
                result += characters.charAt(Math.floor(Math.random() * charactersLength));
                counter += 1;
            }
            return result;
        },

        async create(device) {

            if (this.use_ezcode) {
                device.ezcode = this.make_ezcode()
                console.log("ezcode = ", device.ezcode)
            }
            this.device = device
            this.device.platform = this.platforms.selected.value
            this.device.accountid = this.acntList.selected.value
            this.device.name = this.device.name.trim()

            this.dialogCreate = false;
            await this.createdevice(this.device)
            if (this.use_ezcode) {
                this.dialogEZCode = true;
            }
        },

        startAddVPN(device) {
            this.device = device
            this.readAllNetworks()
            this.vpn = {
                name: "",
                deviceid: device.id,
                email: this.user.email,
                enable: true,
                tags: [],
                current: {},
            }


            this.netList = {
                selected: { "text": "", "value": "" },
                items: []
            }

            var selected = 0;
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i] = { "text": this.nets[i].netName, "value": this.nets[i].id }
                if (this.netList.items[i].text == this.device.netName) {
                    selected = i
                    break
                }
            }

            this.netList.selected = this.netList.items[selected];
            this.dialogAddVPN = true;
        },

        updateDefaults(net) {
            console.log("updateDefaults", net)
            var selected = 0;
            for (let i = 0; i < this.nets.length; i++) {
                if (this.nets[i].id == net.value) {
                    selected = i
                    break
                }
            }

            // Change the host name to be the device name + the network name
            this.vpn.name = this.device.name + "." + this.nets[selected].netName

            this.vpn.current.syncEndpoint = this.nets[selected].default.syncEndpoint
            this.vpn.current.hasSSH = this.nets[selected].default.hasSSH
            this.vpn.current.hasRDP = this.nets[selected].default.hasRDP
            this.vpn.current.upnp = this.nets[selected].default.upnp
            this.vpn.current.failsafe = this.nets[selected].default.failsafe
            this.vpn.current.enableDns = this.nets[selected].default.enableDns
            console.log("updateDefaults = ", this.vpn, this.nets[selected])

        },

        async createVPN(vpn) {
            this.vpn = vpn
            this.vpn.current.listenPort = 0

            // get the listen port from the endpoint field if it is there
            if (this.vpn.current.endpoint != null && this.vpn.current.endpoint != "" && this.vpn.current.endpoint.indexOf(":") != -1) {
                let parts = this.vpn.current.endpoint.split(":")
                this.vpn.current.listenPort = parseInt(parts[parts.length-1], 10)
            }

            this.vpn.name = this.vpn.name
            this.vpn.netName = this.netList.selected.text
            this.vpn.netid = this.netList.selected.value
            this.dialogAddVPN = false;
            await this.createvpn(vpn)
            // wait a second
            await new Promise(r => setTimeout(r, 2000));
            await this.asyncRefresh()

        },

        updateVPN(vpn) {
            this.vpn = vpn
            this.vpn.current.listenPort = 0

            // get the listen port from the endpoint field if it is there
            if (this.vpn.current.endpoint != null && this.vpn.current.endpoint != "" && this.vpn.current.endpoint.indexOf(":") != -1) {
                let parts = this.vpn.current.endpoint.split(":")
                this.vpn.current.listenPort = parseInt(parts[parts.length-1], 10)
            }

            this.vpn.current.persistentKeepalive = parseInt(this.vpn.current.persistentKeepalive, 10);
            this.vpn.current.mtu = parseInt(this.vpn.current.mtu, 10);

            if (this.publicSubnets) {
                this.vpn.current.allowedIPs.push("0.0.0.0/5", "8.0.0.0/7", "11.0.0.0/8", "12.0.0.0/6", "16.0.0.0/4", "32.0.0.0/3", "64.0.0.0/3",
				"96.0.0.0/4", "112.0.0.0/5", "120.0.0.0/6", "124.0.0.0/7", "126.0.0.0/8", "128.0.0.0/3", "160.0.0.0/5", "168.0.0.0/6",
				"172.0.0.0/12", "172.32.0.0/11", "172.64.0.0/10", "172.128.0.0/9", "173.0.0.0/8", "174.0.0.0/7", "176.0.0.0/4", "192.0.0.0/9", "192.128.0.0/11",
				"192.160.0.0/13", "192.169.0.0/16", "192.170.0.0/15", "192.172.0.0/14", "192.176.0.0/12", "192.192.0.0/10", "193.0.0.0/8", "194.0.0.0/7",
				"196.0.0.0/6", "200.0.0.0/5", "208.0.0.0/4", "::/1", "8000::/2", "c000::/3", "e000::/4", "f000::/5", "f800::/6", "fe00::/9", "fe80::/10", "ff00::/8")
            }

            // check allowed IPs
            for (let i = 0; i < this.vpn.current.allowedIPs.length; i++) {
                if (this.$isCidr(this.vpn.current.allowedIPs[i]) === 0) {
                    this.errorDevice('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }
            // check address
            for (let i = 0; i < this.vpn.current.address.length; i++) {
                if (this.$isCidr(this.vpn.current.address[i]) === 0) {
                    this.errorDevice('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }
            vpn = this.vpn
            this.inEdit = false;
            this.updatevpn(this.vpn)
    //        this.updatedevice_vpn(this.vpn)
//            this.Refreshing()
        },


        removeDevice(device) {
            if (confirm(`Do you really want to delete ${device.name} ?`)) {
                this.deletedevice(device)
            }
        },

        async removeVPN(vpn) {
            if (confirm(`Do you really want to delete ${vpn.name} from ${vpn.netName}?`)) {
                this.deletevpn(vpn)
                // refresh the page
                this.Refreshing()
            }
        },

        startUpdate(device) {

            this.device = device;
            //        this.readQrCode(this.device);
            //        this.readConfig(device);

            this.netList = {
                selected: { "text": this.device.netName, "value": this.device.netid },
                items: []
            }

            var selected = 0;
            for (let i = 0; i < this.nets.length; i++) {
                this.netList.items[i] = { "text": this.nets[i].netName, "value": this.nets[i].id }
                if (this.netList.items[i].text == this.device.netName) {
                    selected = i
                }
            }

            this.netList.selected = this.netList.items[selected];

            for (let i = 0; i < this.platforms.items.length; i++) {
                if (this.platforms.items[i].value == this.device.platform) {
                    this.platforms.selected = this.platforms.items[i]
                    break
                }
            }

            this.publicSubnets = false;
            this.dialogUpdate = true;

        },

        startUpdateServiceHost(device) {

            this.device = device;
            //        this.readConfig(device);

            this.dialogServiceHost = true;

        },


        updateEnable(device) {
            // the switch automatically updates device.enable to the proper value
            this.updatedevice(device)
        },

        copyDeviceConfig(device) {
            var url = "curl \"http://localhost:53280/config/?id=" + device.id + "&apiKey=" + device.apiKey + "&server=" + device.server + "\""

            // copy url to clipboard 

            navigator.clipboard
                .writeText(url)
                .then(() => {
                    this.errorDevice('Copied to clipboard');
                })

        },

        async updateDevice(item) {
            console.log("updateDevice = ", item)
            this.noEdit = true;
            this.device = item.device;

            this.device.platform = item.platform.value
            console.log("platform = ", this.device.platform)

            // set the account id
            this.device.accountid = item.accountid.value
            console.log("accountid = ", this.device.accountid)

            // fixup the checkInterval
            this.device.checkInterval = parseInt(this.device.checkInterval, 10);

            // all good, submit
            this.dialogUpdate = false;
            this.dialogServiceHost = false;
            this.inEdit = false;

            this.updatedevice(this.device)
            this.Refreshing()
        },

        async forceFileDownload(vpn) {
            console.log(vpn)
            await this.readvpnconfig(vpn)
            // sleep for one second
            await new Promise(r => setTimeout(r, 1000));
            let config = this.getvpnconfig(vpn.id)
            if (!config) {
                console.log("failed to get config")
                this.errorDevice('Failed to download device config');
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
