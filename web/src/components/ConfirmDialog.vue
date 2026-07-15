<script setup lang="ts">
defineProps<{
  open: boolean
  title: string
  message: string
  confirmText?: string
  danger?: boolean
  busy?: boolean
}>()

defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-40 flex items-center justify-center bg-black/60 px-4">
    <section class="app-card w-full max-w-md rounded-2xl p-6 shadow-2xl">
      <h2 class="app-heading text-lg font-semibold">{{ title }}</h2>
      <p class="app-muted mt-3 text-sm leading-6">{{ message }}</p>
      <div class="mt-6 flex justify-end gap-3">
        <button class="app-button-secondary text-sm" :disabled="busy" @click="$emit('cancel')">
          取消
        </button>
        <button
          class="text-sm disabled:opacity-60"
          :class="danger ? 'app-button-danger' : 'app-button-primary'"
          :disabled="busy"
          @click="$emit('confirm')"
        >
          {{ busy ? '处理中...' : confirmText || '确认' }}
        </button>
      </div>
    </section>
  </div>
</template>
