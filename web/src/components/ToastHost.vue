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
        'border-emerald-400/40 bg-emerald-950 text-emerald-100': toast.tone === 'success',
        'border-red-400/40 bg-red-950 text-red-100': toast.tone === 'error',
        'border-cyan-400/40 bg-cyan-950 text-cyan-100': toast.tone === 'info',
      }"
    >
      {{ toast.message }}
    </div>
  </div>
</template>
