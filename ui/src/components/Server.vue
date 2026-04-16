<template>
  <v-container v-if="server">
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>Net Configuration</v-card-title>
          <div class="d-flex flex-no-wrap justify-space-between">
            <v-col cols="12">
              <v-text-field v-model="server.id" label="Net ID" disabled />
              <v-text-field v-model="server.netName" label="Net Name" disabled />
              <v-combobox
                v-model="server.address"
                chips
                hint="Write IPv4 or IPv6 CIDR and hit enter. A 100.x.x.0/24 address is recommended."
                label="Host interface address pool"
                multiple
              >
                <template #selection="{ item }">
                  <v-chip
                    closable
                    @click:close="server.address.splice(server.address.indexOf(item.value), 1)"
                  >
                    <strong>{{ item.value }}</strong>&nbsp;
                  </v-chip>
                </template>
              </v-combobox>
            </v-col>
          </div>
          <div class="d-flex flex-no-wrap justify-space-between">
            <v-col cols="12">
              <v-text-field
                v-model="server.listenPort"
                type="number"
                :rules="[v => !!v || 'Listen port is required']"
                label="Default Listen port"
                required
              />
              <v-combobox
                v-model="server.dns"
                chips
                hint="Write IPv4 or IPv6 address and hit enter"
                label="DNS servers for clients"
                multiple
              >
                <template #selection="{ item }">
                  <v-chip
                    closable
                    @click:close="server.dns.splice(server.dns.indexOf(item.value), 1)"
                  >
                    <strong>{{ item.value }}</strong>&nbsp;
                  </v-chip>
                </template>
              </v-combobox>
              <v-combobox
                v-model="server.allowedips"
                chips
                hint="Write IPv4 or IPv6 address and hit enter"
                label="Default Allowed IPs for clients"
                multiple
              >
                <template #selection="{ item }">
                  <v-chip
                    closable
                    @click:close="server.allowedips.splice(server.allowedips.indexOf(item.value), 1)"
                  >
                    <strong>{{ item.value }}</strong>&nbsp;
                  </v-chip>
                </template>
              </v-combobox>
              <v-text-field
                v-model="server.mtu"
                type="number"
                label="Define global MTU"
                hint="Leave at 0 and let wg-quick take care of MTU"
              />
              <v-text-field
                v-model="server.persistentKeepalive"
                type="number"
                label="Persistent keepalive"
                hint="Leave at 0 if you dont want to specify persistent keepalive"
              />
            </v-col>
          </div>
        </v-card>
      </v-col>
    </v-row>
    <v-divider />
    <v-btn class="ma-2" color="success" @click="forceFileDownload">
      Download server configuration
      <v-icon end>mdi-cloud-download-outline</v-icon>
    </v-btn>
    <v-spacer></v-spacer>
    <v-btn class="ma-2" color="warning" @click="updateServer">
      Update server configuration
      <v-icon end>mdi-update</v-icon>
    </v-btn>
    <v-divider />
  </v-container>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useServerStore } from '@/stores/server'
import { isCidr } from '@/plugins/cidr'

const serverStore = useServerStore()

const server = computed(() => serverStore.servers?.[0] ?? null)

onMounted(() => {
  serverStore.read()
})

function updateServer() {
  if (!server.value) return
  server.value.listenPort = parseInt(server.value.listenPort, 10)
  server.value.persistentKeepalive = parseInt(server.value.persistentKeepalive, 10)
  server.value.mtu = parseInt(server.value.mtu, 10)

  if (server.value.address.length < 1) {
    serverStore.error = 'Please provide at least one valid CIDR address for server interface'
    return
  }
  for (const addr of server.value.address) {
    if (isCidr(addr) === 0) {
      serverStore.error = `Invalid CIDR detected, please correct ${addr} before submitting`
      return
    }
  }
  for (const dns of server.value.dns) {
    if (isCidr(dns + '/32') === 0) {
      serverStore.error = `Invalid IP detected, please correct ${dns} before submitting`
      return
    }
  }
  for (const ip of server.value.allowedips) {
    if (isCidr(ip) === 0) {
      serverStore.error = 'Invalid CIDR detected, please correct before submitting'
      return
    }
  }
  serverStore.update(server.value)
}

function forceFileDownload() {
  const url = window.URL.createObjectURL(new Blob([serverStore.config]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', 'wg0.conf')
  document.body.appendChild(link)
  link.click()
}
</script>
