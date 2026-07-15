<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <section class="app-card rounded-3xl border-dashed p-10 text-center">
    <h2 class="app-heading text-lg font-semibold">{{ title }}</h2>
    <p class="app-muted mt-2 text-sm">{{ description }}</p>
    <div class="mt-5">
      <slot />
    </div>
  </section>
</template>
