<script setup lang="ts">
import { useRouter } from 'vue-router'
import { authStore } from '@/stores/auth'
import { themeStore, type ThemePreference } from '@/stores/theme'

const router = useRouter()
const navItems = [
  { to: '/endpoints', label: 'Endpoint', icon: '↗' },
  { to: '/logs', label: '日志', icon: '≡' },
  { to: '/settings', label: '设置', icon: '⚙' },
]

function setTheme(event: Event): void {
  themeStore.setPreference((event.target as HTMLSelectElement).value as ThemePreference)
}

function logout(): void {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="app-shell lg:grid lg:grid-cols-[17rem_1fr]">
    <aside class="app-sidebar sticky top-0 z-20 border-b px-4 py-4 lg:h-screen lg:border-b-0 lg:border-r lg:p-5">
      <div class="flex items-center justify-between lg:block">
        <RouterLink to="/endpoints" class="flex items-center gap-3">
          <span class="flex h-10 w-10 items-center justify-center rounded-2xl" style="background: var(--primary-soft); color: var(--primary);">m</span>
          <span>
            <span class="block text-lg font-semibold tracking-tight">meowbridge</span>
            <span class="app-muted hidden text-xs lg:block">Webhook to MeoW</span>
          </span>
        </RouterLink>
        <button class="app-button-ghost px-3 py-2 text-sm lg:hidden" @click="logout">退出</button>
      </div>

      <nav class="mt-5 flex gap-2 overflow-x-auto text-sm lg:flex-col lg:overflow-visible">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          class="app-nav-link shrink-0"
          :to="item.to"
        >
          <span>{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="mt-6 hidden lg:block">
        <p class="app-muted text-xs uppercase tracking-[0.2em]">运行模式</p>
        <div class="app-card-muted mt-3 rounded-2xl p-4 text-sm">
          <p class="font-medium">单实例 SQLite</p>
          <p class="app-muted mt-1 text-xs">配置保持扁平，适合 Docker 直接运行。</p>
        </div>
      </div>
    </aside>

    <div class="min-w-0">
      <header class="px-4 pt-4 lg:px-8 lg:pt-6">
        <div class="app-topbar flex items-center justify-between rounded-2xl px-4 py-3">
          <div>
            <p class="app-muted text-xs uppercase tracking-[0.22em]">Admin Console</p>
            <p class="text-sm font-medium">管理 Webhook 入口与推送日志</p>
          </div>
          <div class="flex items-center gap-2">
            <select class="app-input w-32 py-2 text-sm" :value="themeStore.preference.value" @change="setTheme">
              <option value="system">跟随系统</option>
              <option value="light">亮色</option>
              <option value="dark">暗色</option>
            </select>
            <button class="app-button-secondary hidden text-sm lg:inline-flex" @click="logout">
              退出
            </button>
          </div>
        </div>
      </header>

      <main class="mx-auto max-w-7xl px-4 py-6 lg:px-8 lg:py-8">
        <slot />
      </main>
    </div>
  </div>
</template>
