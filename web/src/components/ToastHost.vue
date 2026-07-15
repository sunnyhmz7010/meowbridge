<script setup lang="ts">
import { toasts } from './toast'
</script>

<template>
  <div class="fixed right-4 top-4 z-50 flex w-80 flex-col gap-2">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="rounded-xl border px-4 py-3 text-sm shadow-lg"
      :class="{
        'app-badge-success': toast.tone === 'success',
        'app-badge-danger': toast.tone === 'error',
        'app-badge-muted': toast.tone === 'info',
      }"
    >
      {{ toast.message }}
    </div>
  </div>
</template>
