<template>
    <q-page padding class="row justify-center items-center">
        <q-card style="width: 360px; max-width: 95vw">
        <q-card-section>
            <div class="text-h6">Sign in</div>
        </q-card-section>

        <q-separator />

        <q-card-section>
            <q-form @submit.prevent="submit" class="column q-gutter-sm">
            <q-input v-model="username" label="Username" dense outlined />
            <q-input v-model="password" type="password" label="Password" dense outlined />
            <q-btn label="Login" color="primary" type="submit" :loading="auth.loading" />
            <div v-if="auth.error" class="text-negative text-caption q-mt-xs">{{ auth.error }}</div>
            </q-form>
        </q-card-section>
        </q-card>
    </q-page>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from 'src/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('guest')
const password = ref('guest')

async function submit () {
  await auth.login(username.value, password.value)
  const redirect = route.query.redirect || '/queues'
  router.replace(redirect)
}
</script>
