<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError, apiClient } from '@/api/client'
import { authStore } from '@/stores/auth'

const router = useRouter()
const password = ref('')
const error = ref('')
const loading = ref(false)
const checkingSetup = ref(true)
const needsSetup = ref(false)
const title = computed(() => needsSetup.value ? '初始化管理员' : '登录管理后台')
const description = computed(() => needsSetup.value
  ? '首次打开时设置管理员密码，设置后会自动进入后台。'
  : '输入管理员密码登录后台。')
const buttonText = computed(() => {
  if (loading.value) {
    return needsSetup.value ? '初始化中...' : '登录中...'
  }
  return needsSetup.value ? '设置密码并进入后台' : '登录'
})

async function submit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    if (needsSetup.value) {
      await authStore.setup(password.value)
    } else {
      await authStore.login(password.value)
    }
    await router.push('/endpoints')
  } catch (err) {
    error.value = err instanceof ApiError && err.status === 401
      ? '密码错误或凭证无效'
      : err instanceof ApiError ? err.message : (needsSetup.value ? '初始化失败' : '登录失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  checkingSetup.value = true
  try {
    const status = await apiClient.getSetupStatus()
    needsSetup.value = status.needs_setup
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '加载初始化状态失败'
  } finally {
    checkingSetup.value = false
  }
})
</script>

<template>
  <main class="app-shell grid min-h-screen place-items-center px-4 py-10">
    <section class="grid w-full max-w-5xl overflow-hidden rounded-2xl border lg:grid-cols-[1.08fr_0.92fr]" style="border-color: var(--border); box-shadow: var(--shadow); background: var(--panel);">
      <div class="hidden p-10 lg:block" style="background: linear-gradient(160deg, var(--primary) 0%, #312e81 55%, #0f172a 100%); color: white;">
        <p class="text-sm font-bold uppercase tracking-[0.3em]" style="color: #c3c0ff;">meowbridge</p>
        <h1 class="mt-5 text-4xl font-black tracking-tight">轻量 Webhook 桥接控制台</h1>
        <p class="mt-4 max-w-md leading-7" style="color: #dbe4ff;">
          用独立 token 隔离外部服务入口，统一解析并转发到固定 MeoW 推送端。
        </p>
        <div class="mt-10 grid gap-3 text-sm">
          <div class="rounded-xl border border-white/15 bg-white/10 p-4">标准 Webhook URL 可直接填写</div>
          <div class="rounded-xl border border-white/15 bg-white/10 p-4">SQLite 单文件持久化，部署简单</div>
          <div class="rounded-xl border border-white/15 bg-white/10 p-4">推送日志可追踪解析和上游响应</div>
        </div>
      </div>

      <form class="p-8 sm:p-10" @submit.prevent="submit">
        <p class="text-sm font-bold uppercase tracking-[0.3em]" style="color: var(--primary);">Admin</p>
        <h2 class="app-heading mt-3 text-3xl font-black">{{ title }}</h2>
        <p class="app-muted mt-2 text-sm">{{ checkingSetup ? '正在检查初始化状态...' : description }}</p>
        <label class="app-muted mt-8 block text-sm font-semibold">
          管理员密码
          <input
            v-model="password"
            class="app-input mt-2"
            type="password"
            :autocomplete="needsSetup ? 'new-password' : 'current-password'"
            required
          />
        </label>
        <p v-if="error" class="mt-4 rounded-xl border px-3 py-2 text-sm" style="border-color: color-mix(in srgb, var(--danger) 40%, transparent); background: var(--danger-soft); color: var(--danger);">
          {{ error }}
        </p>
        <button
          class="app-button-primary mt-6 w-full py-3 disabled:opacity-60"
          :disabled="loading || checkingSetup"
        >
          {{ checkingSetup ? '检查中...' : buttonText }}
        </button>
      </form>
    </section>
  </main>
</template>
