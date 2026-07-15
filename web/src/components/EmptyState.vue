<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <section class="rounded-2xl border border-dashed border-slate-700 bg-slate-900/60 p-10 text-center">
    <h2 class="text-lg font-semibold">{{ title }}</h2>
    <p class="mt-2 text-sm text-slate-400">{{ description }}</p>
    <div class="mt-5">
      <slot />
    </div>
  </section>
</template>
