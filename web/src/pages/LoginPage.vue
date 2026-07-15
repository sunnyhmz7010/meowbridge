<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError } from '@/api/client'
import { authStore } from '@/stores/auth'

const router = useRouter()
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(password.value)
    await router.push('/endpoints')
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-slate-950 px-4 text-slate-100">
    <form class="w-full max-w-sm rounded-2xl border border-slate-800 bg-slate-900 p-8 shadow-xl" @submit.prevent="submit">
      <p class="text-sm uppercase tracking-[0.3em] text-cyan-300">meowbridge</p>
      <h1 class="mt-3 text-2xl font-semibold">登录管理后台</h1>
      <label class="mt-8 block text-sm text-slate-300">
        管理员密码
        <input
          v-model="password"
          class="mt-2 w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 outline-none focus:border-cyan-400"
          type="password"
          autocomplete="current-password"
          required
        />
      </label>
      <p v-if="error" class="mt-4 rounded-lg border border-red-500/40 bg-red-950 px-3 py-2 text-sm text-red-100">
        {{ error }}
      </p>
      <button
        class="mt-6 w-full rounded-xl bg-cyan-500 px-4 py-3 font-medium text-slate-950 hover:bg-cyan-400 disabled:opacity-60"
        :disabled="loading"
      >
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>
  </main>
</template>
