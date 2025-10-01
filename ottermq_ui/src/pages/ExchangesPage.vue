<template>
  <q-page padding>
    <div class="container">

      <div class="text-h6 q-mb-md">Exchanges</div>

        <q-form class="row q-gutter-sm q-mb-md" @submit.prevent="create">
          <q-input v-model="newExchange" label="Exchange name" dense outlined />
          <q-select v-model="type" :options="types" label="Type" dense outlined style="width: 160px" />
          <q-btn label="Add" color="primary" type="submit" />
        </q-form>

        <q-table
          :rows="store.items"
          :columns="columns"
          row-key="name"
          flat bordered
          :loading="store.loading"
          @row-click="(_, row) => select(row.name)"
        >
          <template #body-cell-actions="props">
            <q-td :props="props">
              <q-btn size="sm" color="negative" label="Delete" @click.stop="del(props.row.name)" />
            </q-td>
          </template>
        </q-table>

        <div v-if="store.selected" class="q-mt-lg">
          <div class="text-subtitle1 q-mb-sm">Bindings for <b>{{ store.selected }}</b></div>

          <q-form class="row q-gutter-sm q-mb-md" @submit.prevent="addBinding">
            <q-input v-model="routingKey" label="Routing key" dense outlined />
            <q-input v-model="queueName" label="Queue" dense outlined />
            <q-btn label="Bind" color="primary" type="submit" />
          </q-form>

          <q-table
            :rows="store.bindings"
            :columns="bindingCols"
            row-key="routingKey-queue"
            flat bordered
          >
            <template #body-cell-actions="props">
              <q-td :props="props">
                <q-btn size="sm" label="Unbind" @click="unbind(props.row)" />
              </q-td>
            </template>
          </q-table>

          <q-separator class="q-my-md" />

          <div class="text-subtitle1 q-mb-sm">Publish Message</div>
          <q-form class="column q-gutter-sm" @submit.prevent="publish">
            <q-input v-model="pubRoutingKey" label="Routing key" dense outlined />
            <q-input v-model="message" type="textarea" autogrow label="Message" dense outlined />
            <q-btn label="Publish" color="primary" type="submit" />
          </q-form>
        </div>
    </div>
  </q-page>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useExchangesStore } from 'src/stores/exchanges'
import { Notify } from 'quasar'

const store = useExchangesStore()

const columns = [
  { name: 'vhost', label: 'VHost', field: 'vhost' },
  { name: 'name', label: 'Name', field: 'name' },
  { name: 'type', label: 'Type', field: 'type' },
  { name: 'actions', label: 'Actions', field: 'actions' }
]

const bindingCols = [
  { name: 'routingKey', label: 'Routing Key', field: 'routingKey' },
  { name: 'queue', label: 'Queue', field: 'queue' },
  { name: 'actions', label: 'Actions', field: 'actions' }
]

const newExchange = ref('')
const type = ref('direct')
const types = ['direct', 'fanout', 'topic', 'headers']

const routingKey = ref('')
const queueName = ref('')

const pubRoutingKey = ref('')
const message = ref('')

function select(name) {
  store.select(name)
  store.fetchBindings(name)
}

async function create() {
  if (!newExchange.value) return
  await store.addExchange(newExchange.value, type.value)
  newExchange.value = ''
}

async function del(name) {
  await store.deleteExchange(name)
}

async function addBinding() {
  if (!store.selected) return
  await store.addBinding(store.selected, routingKey.value, queueName.value)
  routingKey.value = ''; queueName.value = ''
}

async function unbind(row) {
  await store.deleteBinding(store.selected, row.routingKey, row.queue)
}

async function publish() {
  if (!store.selected) return
  await store.publish(store.selected, pubRoutingKey.value, message.value)
  Notify.create({ type: 'positive', message: 'Message published!' })
  pubRoutingKey.value = ''; message.value = ''
}

onMounted(store.fetch)
</script>
