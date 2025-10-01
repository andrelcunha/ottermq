<template>
  <div>
    <div class="row items-center q-gutter-sm cursor-pointer" @click="open = !open">
      <span class="triangle">{{ open ? '▾' : '▸' }}</span>
      <slot name="title" />
    </div>
    <div v-show="open" class="q-mt-sm">
      <slot />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
const open = ref(true)
</script>

<style scoped lang="scss">
.triangle { margin-right: 8px; color: #999; }
</style>
