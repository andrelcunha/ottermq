<template>
  <q-page padding>
    <div class="container">

      <div class="text-h6 q-mb-md">Connections</div>

        <q-table
          :rows="rows"
          :columns="columns"
          row-key="name"
          flat 
          bordered
          :loading="store.loading"
        >
          <template #body-cell-state="props">
            <q-td :props="props">
              <span 
              class="small-square q-mx-xs"
              :class="props.row.state === 'running'? 'small-square--green' : 'small-square--red'"
              />
              {{ props.row.state || '-' }}
            </q-td>
          </template>

          <template #body-cell-ssl="props">
            <q-td :props="props">{{ props.row.ssl ? '●' : '○' }}</q-td>
          </template>

          <template #body-cell-connectedAt="props">
            <q-td :props="props">
              <div class="show-time">{{ props.row.time }}</div>
              <div class="show-date">{{ props.row.date }}</div>
            </q-td>
          </template>
        </q-table>
      </div>
  </q-page>
</template>

<script setup>
import { onMounted, onBeforeUnmount, computed } from 'vue'
import { useConnectionsStore } from 'src/stores/connections'

const store = useConnectionsStore()
let timer = null

const columns = [
  { name: 'vhost', label: 'VHost', field: 'vhost' },
  { name: 'name', label: 'Name', field: 'name' },
  { name: 'user_name', label: 'User', field: 'user_name' },
  { name: 'state', label: 'State', field: 'state' },
  { name: 'ssl', label: 'SSL', field: 'ssl', align: 'center' },
  { name: 'protocol', label: 'Protocol', field: 'protocol', align: 'right' },
  { name: 'channels', label: 'Channels', field: 'channels', align: 'right' },
  { name: 'heartbeat', label: 'Last Heartbeat', field: 'heartbeat', align: 'right' },
  { name: 'connectedAt', label: 'Connected At', field: 'connectedAt', align: 'right' }
]

const rows = computed(() => store.items.map(c => {
  const d = new Date(c.connected_at)
  return {
    ...c,
    heartbeat: heartbeatSeconds(c.last_heartbeat) + 's',
    time: formatTime(d),
    date: formatDate(d)
  }
}))

function heartbeatSeconds(last) {
  const lastDate = new Date(last).getTime()
  const now = Date.now()
  return Math.floor((now - lastDate) / 1000)
}

function formatTime(date) {
  const h = String(date.getHours()).padStart(2,'0')
  const m = String(date.getMinutes()).padStart(2,'0')
  const s = String(date.getSeconds()).padStart(2,'0')
  return `${h}:${m}:${s}`
}
function formatDate(date) {
  const y = date.getFullYear()
  const m = String(date.getMonth()+1).padStart(2,'0')
  const d = String(date.getDate()).padStart(2,'0')
  return `${y}-${m}-${d}`
}

onMounted(async () => {
  await store.fetch()
  timer = setInterval(store.fetch, 15000)
})

onBeforeUnmount(() => { if (timer) clearInterval(timer) })
</script>

<style scoped lang="scss">
.small-square { 
    height: .7em; 
    width: .7em; 
    // background: #fff; 
    border: 50%; 
    padding: 2px; 
    margin-right: 4px; 
    display: inline-block; 
}
.small-square--green { background: #0f0}
.small-square--red { background: #c10015}
.show-time { font-size: 1em; margin-bottom: 2px; }
.show-date { font-size: .8em; color: #999; }
</style>
