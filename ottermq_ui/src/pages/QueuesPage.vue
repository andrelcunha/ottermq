<template>
  <q-page padding>
    <div class="container">

      <div class="q-mt-md">
        <div class="text-h6 q-mb-md">Queues</div>

        <q-form class="row q-gutter-sm q-mb-md" @submit.prevent="create">
            <q-input v-model="newQueue" label="Queue name" dense outlined />
            <q-btn label="Add" color="primary" type="submit" />
          </q-form>

          <q-table
            :rows="rows"
            :columns="columns"
            row-key="name"
            flat bordered
            :loading="store.loading"
            @row-click="(_, row) => select(row)"
          >
            <template #body-cell-status="props">
              <q-td :props="props">
                <span class="small-green-square" /> running
              </q-td>
            </template>

            <template #body-cell-actions="props">
              <q-td :props="props">
                <q-btn size="sm" color="negative" label="Delete" @click.stop="del(props.row.name)" />
              </q-td>
            </template>
          </q-table>

          <div v-if="store.selected" class="q-mt-md">
            <q-form class="row q-gutter-sm" @submit.prevent="consume">
              <q-input v-model="selectedName" label="Selected queue" dense outlined readonly />
              <q-btn label="Get message" color="primary" type="submit" />
            </q-form>

            <q-card v-if="store.lastMessage" class="q-mt-md">
              <q-card-section>
                <div class="text-subtitle2">Message</div>
                <pre class="q-mt-sm">{{ store.lastMessage }}</pre>
              </q-card-section>
            </q-card>
          </div>
      </div>
    </div>
  </q-page>
</template>

<script setup>
import { onMounted, ref, computed, watch } from 'vue'
import { useQueuesStore } from 'src/stores/queues'

const store = useQueuesStore()
const newQueue = ref('')
const selectedName = ref('')

const columns = [
  { name: 'vhost', label: 'VHost', field: 'vhost' },
  { name: 'name', label: 'Name', field: 'name', format: (v) => String(v) },
  { name: 'status', label: 'Status', field: 'status' },
  { name: 'messages', label: 'Messages', field: 'messages', align: 'right' },
  { name: 'x', label: 'X', field: 'x', align: 'right' },
  { name: 'y', label: 'Y', field: 'y', align: 'right' },
  { name: 'z', label: 'Z', field: 'z', align: 'right' },
  { name: 'actions', label: 'Actions', field: 'actions' }
]

const rows = computed(() =>
  store.items.map(q => ({
    ...q,
    x: 0, y: 0, z: 0
  }))
)

function select(row) {
  store.select(row.name)
}

watch(() => store.selected, v => { selectedName.value = v || '' })

async function create() {
  if (!newQueue.value) return
  await store.addQueue(newQueue.value)
  newQueue.value = ''
}

async function del(name) {
  await store.deleteQueue(name)
}

async function consume() {
  if (!selectedName.value) return
  await store.consume(selectedName.value)
}

onMounted(store.fetch)
</script>

<style scoped lang="scss">
.small-green-square {
  height: 5px; width: 5px; background: #0f0; border: 1px solid lightslategray; padding: 2px; margin-right: 4px; display: inline-block;
}
</style>
