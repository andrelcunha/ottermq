<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar>

        <q-toolbar-title> Ottermq </q-toolbar-title>

        <div>User: {{username}}</div>
        <q-btn flat label="Logout" @click="logout" />
      </q-toolbar>
    </q-header>

    <q-tabs
      v-model="tab"
      class="bg-primary text-primary"
      align="justify"
      dense
      active-color="primary"
      indicator-color="primary"
    >
      <q-tab name="overview" label="Overview" @click="navigate('overview')"/>
      <q-tab name="connections" label="Connections" @click="navigate('connections')"/>
      <q-tab name="exchanges" label="Exchanges" @click="navigate('exchanges')"/>
      <q-tab name="queues" label="Queues" @click="navigate('queues')"/>
    </q-tabs>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const tab = ref('overview')
const username = ref('guest')
const router = useRouter()

function logout() {
  // clear auth and redirect -- Mocked for now
  router.push('/login')
}

const navigate = (route) => {
  router.push(route)
}

</script>
