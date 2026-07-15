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
    <section class="w-full max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
      <h2 class="text-lg font-semibold">{{ title }}</h2>
      <p class="mt-3 text-sm leading-6 text-slate-300">{{ message }}</p>
      <div class="mt-6 flex justify-end gap-3">
        <button class="rounded-lg border border-slate-700 px-4 py-2 text-sm" :disabled="busy" @click="$emit('cancel')">
          取消
        </button>
        <button
          class="rounded-lg px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
          :class="danger ? 'bg-red-600 hover:bg-red-500' : 'bg-cyan-600 hover:bg-cyan-500'"
          :disabled="busy"
          @click="$emit('confirm')"
        >
          {{ busy ? '处理中...' : confirmText || '确认' }}
        </button>
      </div>
    </section>
  </div>
</template>
