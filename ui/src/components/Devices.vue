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
                        <v-row>
                            <v-col cols="4">Devices</v-col>
                            <v-col cols="4">

                                <v-text-field v-if="listView" v-model="search" append-icon="mdi-magnify" label="Search" single-line
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
                    <v-row>
                        <v-col cols="6">
                            <v-treeview v-if="showTree" :items="items" :search="search" :active.sync="active"
                                :open.sync="open" activatable open-all hoverable>
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
                                        <v-icon color="#336699">mdi-plus-circle</v-icon>
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
                                <v-card-text width="600">
                                    <v-icon v-if="selected.device.type=='Service'">mdi-cloud</v-icon>
                                    <v-icon v-else>mdi-devices</v-icon>
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

                                <v-row class="px-3" width="300">
                                    <v-col flex>
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
                                        <v-text-field v-model="selected.device.description" label="Description" :readonly="!inEdit" />
                                        <v-select return-object v-model="selected.platform" :items="platforms.items"
                                            item-text="text" item-value="value" label="Platform of this device" single
                                            persistent-hint :readonly="!inEdit" />

                                        <v-switch v-model="selected.device.enable" color="success" inset
                                            :label="selected.device.enable ? 'Enabled' : 'Disabled'" :readonly="!inEdit" />
                                        <v-text-field v-model="selected.device.server" label="Server" :readonly="!inEdit" />
                                        <v-text-field v-model="selected.device.id" label="Device ID" readonly />
                                        <v-text-field v-model="selected.device.apiKey" label="API Key" readonly />
                                        <div :hidden="!inEdit">
                                            <v-text-field v-model="selected.device.name" label="Device friendly name"
                                                :rules="[v => !!v || 'device name is required',]" required />
                                        </div>
                                        <p class="text-caption">Created by {{ selected.device.createdBy }} at {{ selected.device.created | formatDate }}<br/>
                                                            Last update by {{ selected.device.updatedBy }} at {{ selected.device.updated | formatDate }}</p>
                                    </v-col>
                                </v-row>
                                <v-card-actions v-if="inEdit">
                                    <v-btn color="success" @click="updateDevice(selected.device)">
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
                            </v-card>
                            <v-card v-else-if="!selected.isDevice">
                                <v-card-text width="600" class="px-3">
                                    <span class="material-symbols-outlined">hub</span>
                                    <h3 class="text-h5 mb-2">
                                        {{ selected.netName }}
                                    </h3>
                                </v-card-text>
                                <v-divider></v-divider>
                                <v-row class="px-3" width="600">
                                    <v-col flex>
                                        <v-text-field v-model="selected.name" label="DNS name" :readonly="!inEdit" />
                                        <v-combobox :readonly="!inEdit" v-model="selected.current.address" chips
                                            hint="Write IPv4 or IPv6 CIDR and hit enter" label="Addresses" multiple dark>
                                            <template v-slot:selection="{ attrs, item, select }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="selected.current.address.splice(selected.current.address.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-combobox :readonly="!inEdit" v-model="selected.current.dns" chips
                                            hint="Enter IP address(es) and hit enter or leave empty."
                                            label="DNS servers for this device" multiple dark>
                                            <template v-slot:selection="{ attrs, item, select }">

                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="selected.current.dns.splice(selected.current.dns.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-text-field :readonly="!inEdit" v-model="selected.current.endpoint"
                                            label="Public endpoint for clients" />
                                        <v-text-field :readonly="!inEdit" v-model="selected.current.listenPort"
                                            type="number" label="Listen port" />
                                        <v-switch v-model="selected.enable" color="success" inset
                                        :label="selected.enable ? 'Enabled' : 'Disabled'" :readonly="!inEdit" />
                                        <p class="text-caption">Created by {{ selected.createdBy }} at {{ selected.created | formatDate }}<br/>
                                                            Last update by {{ selected.updatedBy }} at {{ selected.updated | formatDate }}</p>

                                    </v-col>
                                </v-row>
                                <v-expansion-panels v-if="inEdit">
                                    <v-expansion-panel>
                                        <v-expansion-panel-header dark>Advanced Configuration</v-expansion-panel-header>
                                        <v-expansion-panel-content>
                                            <div class="d-flex flex-no-wrap justify-space-between">
                                                <v-col cols="12">
                                                    <v-combobox v-model="selected.current.allowedIPs" chips
                                                        hint="Write IPv4 or IPv6 CIDR and hit enter" label="Allowed IPs"
                                                        multiple dark>

                                                        <template v-slot:selection="{ attrs, item, select }">
                                                            <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                                @click:close="selected.current.allowedIPs.splice(selected.current.allowedIPs.indexOf(item), 1)">
                                                                <strong>{{ item }}</strong>&nbsp;
                                                            </v-chip>
                                                        </template>
                                                    </v-combobox>
                                                    <v-switch v-model="publicSubnets" color="success" inset
                                                        label="Route all public traffic through tunnel" />

                                                    <v-text-field v-model="selected.id" label="VPN ID" readonly />
                                                    <v-text-field v-model="selected.netid" label="Network ID" readonly />
                                                    <v-text-field v-model="selected.deviceid" label="Device ID" disabled />
                                                    <v-text-field v-model="selected.current.table" label="Table" />
                                                    <v-text-field v-model="selected.current.publicKey" label="Public key" />
                                                    <v-text-field v-model="selected.current.privateKey" label="Private key"
                                                        autocomplete="off"
                                                        :append-icon="showPrivate ? 'mdi-eye' : 'mdi-eye-off'"
                                                        :type="showPrivate ? 'text' : 'password'"
                                                        hint="Clear this field to have the client manage its private key"
                                                        @click:append="showPrivate = !showPrivate" />
                                                    <v-text-field v-model="selected.current.presharedKey"
                                                        label="Preshared Key" autocomplete="off"
                                                        :append-icon="showPreshared ? 'mdi-eye' : 'mdi-eye-off'"
                                                        :type="showPreshared ? 'text' : 'password'"
                                                        @click:append="showPreshared = !showPreshared" />
                                                    <v-text-field type="number" v-model="selected.current.mtu"
                                                        label="Define global MTU"
                                                        hint="Leave at 0 and let us take care of MTU" />
                                                    <v-text-field type="number"
                                                        v-model="selected.current.persistentKeepalive"
                                                        label="Persistent keepalive"
                                                        hint="To disable, set to 0.  Recommended value 29 (seconds)" />
                                                    <v-textarea v-model="selected.current.postUp" label="PostUp Script"
                                                        hint="Only applies to linux servers" />
                                                    <v-textarea v-model="selected.current.postDown" label="PostDown Script"
                                                        hint="Only applies to linux servers" />
                                                    <v-switch v-model="selected.current.subnetRouting" color="success" inset
                                                        label="Enable subnet routing" />
                                                    <v-switch v-model="selected.current.upnp" color="success" inset
                                                        label="Enable UPnP" />
                                                    <v-switch v-model="selected.current.enableDns" color="success" inset
                                                        label="Enable Nettica DNS" />

                                                </v-col>
                                            </div>
                                        </v-expansion-panel-content>
                                    </v-expansion-panel>
                                </v-expansion-panels>


                                <v-card>
                                    <v-card-actions v-if="inEdit">
                                        <v-btn color="success" @click="updateVPN(selected)">
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
                                                    <v-btn color="success" @click="forceFileDownload(selected)">
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
                                                    <v-btn class="px-3" color="error" @click="removeVPN(selected)">
                                                        Delete
                                                        <v-icon right dark>mdi-delete-outline</v-icon>
                                                    </v-btn>
                                                </v-col>
                                            </v-row>
                                        </v-container>
                                    </v-card-actions>
                                </v-card>

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
                                            :rules="[v => !!v || 'device name is required',]" required />

                                        <v-combobox v-model="device.tags" chips hint="Enter a tag, hit tab, hit enter."
                                            label="Tags" multiple dark>
                                            <template v-slot:selection="{ attrs, item, select }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="device.tags.splice(device.tags.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-select return-object v-model="platforms.selected" :items="platforms.items"
                                            item-text="text" item-value="value" label="Platform of this device" single
                                            persistent-hint />

                                        <v-switch v-model="device.enable" color="success" inset
                                            :label="device.enable ? 'Enable device after creation' : 'Disable device after creation'" />
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
                <v-dialog v-if="vpn" v-model="dialogAddVPN" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Add VPN</v-card-title>
                        <v-card-text>
                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-text-field v-model="vpn.name" label="DNS name for this device"
                                            :rules="[v => !!v || 'DNS name is required',]" required />
                                        <v-select return-object v-model="netList.selected" :items="netList.items"
                                            item-text="text" item-value="value" label="Join this network"
                                            :rules="[v => !!v || 'Net is required',]" single persistent-hint required />
                                        <v-text-field v-model="vpn.current.endpoint" label="Public endpoint for clients" />
                                        <v-text-field v-model="vpn.current.listenPort" type="number" label="Listen port" />

                                        <v-combobox v-model="vpn.tags" chips hint="Enter a tag, hit tab, hit enter."
                                            label="Tags" multiple dark>
                                            <template v-slot:selection="{ attrs, item, select, selected }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="vpn.tags.splice(vpn.tags.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-switch v-model="vpn.enable" color="success" inset
                                            :label="vpn.enable ? 'Enable VPN after creation' : 'Disable VPN after creation'" />
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
                <v-dialog v-if="device" v-model="dialogUpdate" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Edit Host</v-card-title>
                        <v-card-text>

                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">

                                        <v-select return-object v-model="netList.selected" :items="netList.items"
                                            item-text="text" item-value="value" label="Join this net"
                                            :rules="[v => !!v || 'Net is required',]" single persistent-hint required />
                                        <v-combobox v-model="device.tags" chips hint="Write tag name and hit enter"
                                            label="Tags" multiple dark>
                                            <template v-slot:selection="{ attrs, item, select }">
                                                <v-chip v-bind="attrs" :input-value="selected" close @click="select"
                                                    @click:close="device.tags.splice(device.tags.indexOf(item), 1)">
                                                    <strong>{{ item }}</strong>&nbsp;
                                                </v-chip>
                                            </template>
                                        </v-combobox>
                                        <v-btn color="success" @click="forceFileDownload(device)">
                                            Download Config
                                            <v-icon right dark>mdi-cloud-download-outline</v-icon>
                                        </v-btn>
                                        <!--                                <v-img :src="'data:image/png;base64, ' + getdeviceQrcode(device.id)"/> -->
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                    </v-card>

                </v-dialog>
                <v-dialog v-if="device" v-model="dialogCopy" max-width="550">
                    <v-card>
                        <v-card-title class="headline">Copy Host to Net</v-card-title>
                        <v-card-text>

                            <v-row>
                                <v-col cols="12">
                                    <v-form ref="form" v-model="valid">
                                        <v-text-field v-model="device.name" label="New name for device"
                                            :rules="[v => !!v || 'device name is required',]" required />

                                        <v-select return-object v-model="netList.selected" :items="netList.items"
                                            item-text="text" item-value="value" label="Copy to this net"
                                            :rules="[v => !!v || 'Net is required',]" single persistent-hint required />
                                    </v-form>
                                </v-col>
                            </v-row>
                        </v-card-text>
                    </v-card>
                    <v-card>
                        <v-card-actions>
                            <v-btn :disabled="!valid" color="success" @click="copy(device)">
                                Submit
                                <v-icon right dark>mdi-check-outline</v-icon>
                            </v-btn>
                            <v-btn color="primary" @click="dialogCopy = false">
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

        notification: {},
        showPrivate: true,
        showPreshared: true,
        showTree: false,
        footerProps: { 'items-per-page-options': [25, 50, 100, -1] },
        listView: true,
        dialogCreate: false,
        dialogAddVPN: false,
        dialogUpdate: false,
        dialogCopy: false,
        dialogServiceHost: false,
        device: null,
        net: null,
        vpn: null,
        vpns: [],
        items: [],
        active: [],
        open: [],
        inEdit: false,
        name: '',
        panel: 1,
        valid: false,
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
                { text: "Android", value: "Android", },
                { text: "Native WireGuard", value: "Native", },
            ],
        },
        search: '',
        headers: [
            { text: 'Name', value: 'name', },
            { text: 'Status', value: 'status', },
            { text: 'Net', value: 'netName', },
            { text: 'IP addresses', value: 'current.address', },
            //        { text: 'ID', value:'id', },
            { text: "Endpoint", value: 'current.endpoint', },
            //        { text: 'Created by', value: 'created', sortable: false, },
            { text: 'Tags', value: 'tags', },
            { text: 'Actions', value: 'action', sortable: false, },

        ],
    }),

    computed: {
        selected() {
            if (!this.active.length) return undefined

            const id = this.active[0]
            console.log("selected id = ", id)

            var vpn = this.vpns.find(vpn => vpn.id === id)
            if (vpn) {
                return vpn
            }

            var device = this.devices.find(device => device.id === id)
            //            if (device) {
            //                this.platforms.selected.value = device.platform
            //}

            return this.items.find(item => item.id === id)
        },
        ...mapGetters({
            getdeviceQrcode: 'device/getdeviceQrcode',
            getdeviceConfig: 'device/getdeviceConfig',
            user: 'auth/user',
            servers: 'server/servers',
            accounts: 'account/accounts',
            devices: 'device/devices',
            nets: 'net/nets',
            deviceQrcodes: 'device/deviceQrcodes',
            getvpnconfig: "vpn/getVPNConfig",

        }),
    },

    mounted() {
        this.readAllAccounts(this.user.email)
        this.readAllDevices()
        this.readAllNetworks()
    },

    watch: {
        devices: function (val) {
            console.log("buildTree = ", this.buildTree())
            this.showTree = true
        },
    },

    methods: {
        ...mapActions('device', {
            errordevice: 'error',
            readAllDevices: 'readAll',
            readQrCode: 'readQrcode',
            readConfig: 'readConfig',
            createdevice: 'create',
            updatedevice: 'update',
            deletedevice: 'delete',
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
            this.readAllNetworks()
            console.log("buildTree = ", this.buildTree())
        },

        async asyncRefresh() {
            await this.readAllAccounts(this.user.email)
            await this.readAllDevices()
            await this.readAllNetworks()
            console.log("buildTree = ", this.buildTree())
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
                    platform: this.devices[i].platform,
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
            this.dialogCreate = true;
        },

        create(device) {
            this.device.platform = this.platforms.selected.value

            this.dialogCreate = false;
            this.createdevice(device)
        },

        startAddVPN(device) {
            this.device = device
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
                }
            }

            this.netList.selected = this.netList.items[selected];
            this.dialogAddVPN = true;
        },

        async createVPN(vpn) {
            this.vpn = vpn
            this.vpn.current.listenPort = parseInt(this.vpn.current.listenPort, 10);
            // append the port to the endpoint if it is not there
            if (this.vpn.current.endpoint != null && this.vpn.current.endpoint != "" && this.vpn.current.endpoint.indexOf(":") == -1) {
                if (this.vpn.current.listenPort == 0) {
                    this.vpn.current.listenPort = 51820
                }
                this.vpn.current.endpoint = this.vpn.current.endpoint + ":" + this.vpn.current.listenPort.toString()
            }
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
            this.vpn.current.listenPort = parseInt(this.vpn.current.listenPort, 10);
            // append the port to the endpoint if it is not there
            if (this.vpn.current.endpoint != null && this.vpn.current.endpoint != "" && this.vpn.current.endpoint.indexOf(":") == -1) {
                if (this.vpn.current.listenPort == 0) {
                    this.vpn.current.listenPort = 51820
                }
                this.vpn.current.endpoint = this.vpn.current.endpoint + ":" + this.vpn.current.listenPort.toString()
            }

            this.vpn.current.persistentKeepalive = parseInt(this.vpn.current.persistentKeepalive, 10);
            this.vpn.current.mtu = parseInt(this.vpn.current.mtu, 10);

            if (this.publicSubnets) {
                this.vpn.current.allowedIPs.push("0.0.0.0/5", "8.0.0.0/7",
                    "11.0.0.0/8", "12.0.0.0/6", "16.0.0.0/4", "32.0.0.0/3", "64.0.0.0/3", "96.0.0.0/6",
                    "101.0.0.0/8", "102.0.0.0/7", "104.0.0.0/5", "112.0.0.0/5", "120.0.0.0/6",
                    "124.0.0.0/7", "126.0.0.0/8", "128.0.0.0/3", "160.0.0.0/5", "168.0.0.0/6",
                    "172.0.0.0/12", "172.32.0.0/11", "172.64.0.0/10",
                    "172.128.0.0/9", "173.0.0.0/8", "174.0.0.0/7", "176.0.0.0/4", "192.0.0.0/9", "192.128.0.0/11", "192.160.0.0/13", "192.169.0.0/16",
                    "192.170.0.0/15", "192.172.0.0/14", "192.176.0.0/12", "192.192.0.0/10", "193.0.0.0/8", "194.0.0.0/7", "196.0.0.0/6", "200.0.0.0/5", "208.0.0.0/4")
            }

            // check allowed IPs
            if (this.vpn.current.allowedIPs.length < 1) {
                this.errordevice('Please provide at least one valid CIDR address for device allowed IPs');
                return
            }
            for (let i = 0; i < this.vpn.current.allowedIPs.length; i++) {
                if (this.$isCidr(this.vpn.current.allowedIPs[i]) === 0) {
                    this.errordevice('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }
            // check address
            if (this.vpn.current.address.length < 1) {
                this.errordevice('Please provide at least one valid CIDR address for device');
                return;
            }
            for (let i = 0; i < this.vpn.current.address.length; i++) {
                if (this.$isCidr(this.vpn.current.address[i]) === 0) {
                    this.errordevice('Invalid CIDR detected, please correct before submitting');
                    return
                }
            }

            this.inEdit = false;
            this.updatevpn(this.vpn)
            this.Refresh()
        },


        removeDevice(device) {
            if (confirm(`Do you really want to delete ${device.name} ?`)) {
                this.deletedevice(device)
            }
        },

        async removeVPN(vpn) {
            if (confirm(`Do you really want to delete ${vpn.name} from ${vpn.netName}?`)) {
                await this.deletevpn(vpn)
                // refresh the page
                this.Refresh()
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
        startCopy(device) {

            this.device = device;
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

            this.dialogCopy = true;
            this.dialogUpdate = false;

        },
        copy(device) {

            this.noEdit = true;
            this.device = device;

            this.device.current.listenPort = parseInt(this.device.current.listenPort, 10);
            this.device.current.persistentKeepalive = parseInt(this.device.current.persistentKeepalive, 10);
            this.device.current.mtu = parseInt(this.device.current.mtu, 10);

            var changed = false;
            if (this.device.netid != this.netList.selected.value) {
                this.device.netName = this.netList.selected.text
                this.device.netid = this.netList.selected.value
                changed = true;
            }
            this.device.netName = this.netList.selected.text
            this.device.platform = this.platforms.selected.value

            if (changed) {
                this.device.id = ""
                this.device.current.endpoint = ""
                this.device.current.listenPort = 0
                this.device.netName = this.netList.selected.text
                this.device.netid = this.netList.selected.value
                this.createdevice(device)

            }

            this.readAllHosts();

            this.dialogCopy = false;
        },

        updateEnable(device) {
            // the switch automatically updates device.enable to the proper value
            this.updatedevice(device)
        },

        copyDeviceConfig(device) {
            var url = "curl \"http://localhost:53280/config/?id=" + device.id + "&apiKey=" + device.apiKey + "&server=https://my.nettica.com\""

            // copy to clipboard
            var dummy = document.createElement("textarea");
            document.body.appendChild(dummy);
            dummy.value = url;
            dummy.select();
            document.execCommand("copy");
            document.body.removeChild(dummy);

            this.notification = {
                show: true,
                color: 'success',
                text: "Copied to clipboard",
                timeout: 2000,
            }

        },

        async updateDevice(device) {

            this.noEdit = true;
            this.device = device;

            device.platform = this.selected.platform.value
            console.log("platform = ", device.platform)

            // all good, submit
            this.dialogUpdate = false;
            this.dialogServiceHost = false;
            this.inEdit = false;
            this.notification = {
                show: true,
                text: "Saved",
                timeout: 2000,
            }

            await this.updatedevice(device)
            await new Promise(r => setTimeout(r, 1000));
            this.Refresh()
        },

        async forceFileDownload(vpn) {
            console.log(vpn)
            await this.readvpnconfig(vpn)
            // sleep for one second
            await new Promise(r => setTimeout(r, 1000));
            let config = this.getvpnconfig(vpn.id)
            if (!config) {
                console.log("failed to get config")
                this.errordevice('Failed to download device config');
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
