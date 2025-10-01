<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar>

        <q-toolbar-title> OtterMQ </q-toolbar-title>

        <div>User: {{username}}</div>
        <q-btn flat label="Logout" @click="logout" />
      </q-toolbar>
    </q-header>



    <q-page-container>
      <q-tabs
        v-model="tab"
        active-color="white"
        indicator-color="white"
        inline-label
        align="left"
        class="bg-primary"
      >
        <q-route-tab to="/overview" name="overview" label="Overview" />
        <q-route-tab to="/connections" name="connections" label="Connections" />
        <q-route-tab to="/exchanges" name="exchanges" label="Exchanges" />
        <q-route-tab to="/queues" name="queues" label="Queues" />
      </q-tabs>

      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const username = ref('guest')
const tab = ref(route.path.split('/')[1] || 'overview')
watch(() => route.path, (p) => {
  tab.value = (p.split('/')[1] || 'overview')
})
function logout() {
  // clear auth and redirect -- Mocked for now
  router.push('/login')
}
</script>

<style scoped lang="scss">
header .container {
  max-width: 1200px;
  margin: 0 auto;
}

.small-square { 
    height: .7em; 
    width: .7em; 
    border: 1px solid lightslategray;
    padding: 2px; 
    margin-right: 4px; 
    display: inline-block; 
}
.small-square--green { background: #0f0}
.small-square--red { background: #c10015}
</style>
