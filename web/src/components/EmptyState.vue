<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <section class="app-card border-dashed p-10 text-center">
    <p class="app-section-kicker">Empty State</p>
    <h2 class="app-heading mt-2 text-xl font-black">{{ title }}</h2>
    <p class="app-muted mt-2 text-sm">{{ description }}</p>
    <div class="mt-5">
      <slot />
    </div>
  </section>
</template>
