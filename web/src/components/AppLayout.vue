<script setup lang="ts">
import { useRouter } from 'vue-router'
import { authStore } from '@/stores/auth'
import { themeStore, type ThemePreference } from '@/stores/theme'

const router = useRouter()
const navItems = [
  { to: '/endpoints', label: 'Endpoints', icon: '⌘' },
  { to: '/logs', label: '投递日志', icon: '▤' },
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
  <div class="app-shell lg:grid lg:grid-cols-[16.25rem_1fr]">
    <aside class="app-sidebar sticky top-0 z-20 border-b px-4 py-4 lg:h-screen lg:border-b-0 lg:border-r lg:p-6">
      <div class="flex items-center justify-between lg:block">
        <RouterLink to="/endpoints" class="flex items-center gap-3">
          <span class="flex h-11 w-11 items-center justify-center rounded-xl text-lg font-black" style="background: var(--primary); color: var(--primary-contrast);">m</span>
          <span>
            <span class="block text-xl font-black tracking-tight" style="color: var(--primary);">meowbridge</span>
            <span class="app-muted hidden text-xs font-medium lg:block">Webhook 到 MeoW</span>
          </span>
        </RouterLink>
        <button class="app-button-ghost px-3 py-2 text-sm lg:hidden" @click="logout">退出</button>
      </div>

      <nav class="mt-6 flex gap-2 overflow-x-auto text-sm lg:flex-col lg:overflow-visible">
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

      <div class="mt-8 hidden lg:grid lg:min-h-[calc(100vh-18rem)] lg:content-between">
        <div>
          <p class="app-muted text-xs uppercase tracking-[0.22em]">Runtime</p>
          <div class="app-card-muted mt-3 rounded-xl p-4 text-sm">
            <p class="font-semibold">单实例 SQLite</p>
            <p class="app-muted mt-1 text-xs leading-5">一行 Docker 运行，首次打开页面初始化管理员。</p>
          </div>
        </div>
        <button class="app-button-ghost w-full justify-start text-sm" @click="logout">
          <span class="mr-2">↪</span>
          退出
        </button>
      </div>
    </aside>

    <div class="min-w-0">
      <header class="sticky top-0 z-10 border-b px-4 py-3 backdrop-blur lg:px-8" style="border-color: var(--border); background: color-mix(in srgb, var(--bg) 88%, transparent);">
        <div class="mx-auto flex max-w-[1440px] items-center justify-between">
          <div>
            <p class="app-muted text-xs uppercase tracking-[0.22em]">Admin Console</p>
            <p class="text-sm font-semibold">管理 Webhook 入口、解析器和投递日志</p>
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

      <main class="mx-auto max-w-[1440px] px-4 py-6 lg:px-8 lg:py-8">
        <slot />
      </main>
    </div>
  </div>
</template>
