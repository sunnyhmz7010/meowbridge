<script setup lang="ts">
import { useRouter } from 'vue-router'
import { authStore } from '@/stores/auth'
import ToastHost from './ToastHost.vue'

const router = useRouter()

function logout(): void {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100">
    <ToastHost />
    <header class="border-b border-slate-800 bg-slate-900/80 backdrop-blur">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-4">
        <RouterLink to="/endpoints" class="text-lg font-semibold tracking-tight">meowbridge</RouterLink>
        <nav class="flex items-center gap-3 text-sm text-slate-300">
          <RouterLink class="rounded-lg px-3 py-2 hover:bg-slate-800" to="/endpoints">Endpoint</RouterLink>
          <RouterLink class="rounded-lg px-3 py-2 hover:bg-slate-800" to="/logs">日志</RouterLink>
          <RouterLink class="rounded-lg px-3 py-2 hover:bg-slate-800" to="/settings">设置</RouterLink>
          <button class="rounded-lg border border-slate-700 px-3 py-2 hover:bg-slate-800" @click="logout">
            退出
          </button>
        </nav>
      </div>
    </header>
    <main class="mx-auto max-w-6xl px-4 py-8">
      <slot />
    </main>
  </div>
</template>
