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
  <div v-if="open" class="fixed inset-0 z-40 flex items-center justify-center bg-black/65 px-4 backdrop-blur-sm">
    <section class="app-card w-full max-w-md p-6 shadow-2xl">
      <p class="app-section-kicker">Confirm</p>
      <h2 class="app-heading mt-2 text-xl font-black">{{ title }}</h2>
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
