<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError, apiClient } from '@/api/client'
import AppLayout from '@/components/AppLayout.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { showToast } from '@/components/toast'
import { authStore } from '@/stores/auth'

const router = useRouter()
const loading = ref(true)
const savingSettings = ref(false)
const changingPassword = ref(false)
const passwordConfirmOpen = ref(false)
const error = ref('')

const settings = reactive({
  log_retention_days: '14',
})

const original = reactive({
  log_retention_days: '14',
})

const passwordForm = reactive({
  old_password: '',
  new_password: '',
})

function applySettings(values: Record<string, string>): void {
  settings.log_retention_days = values.log_retention_days ?? '14'
  original.log_retention_days = settings.log_retention_days
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    applySettings(await apiClient.getSettings())
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '加载设置失败'
  } finally {
    loading.value = false
  }
}

async function saveSettings(): Promise<void> {
  const changes: Record<string, string> = {}
  if (settings.log_retention_days !== original.log_retention_days) {
    changes.log_retention_days = String(settings.log_retention_days)
  }
  if (Object.keys(changes).length === 0) {
    showToast('没有需要保存的设置', 'info')
    return
  }

  savingSettings.value = true
  try {
    await apiClient.updateSettings(changes)
    showToast('设置已保存', 'success')
  } catch (err) {
    showToast(err instanceof ApiError ? err.message : '保存设置失败', 'error')
    savingSettings.value = false
    return
  }

  try {
    applySettings(await apiClient.getSettings())
  } catch {
    showToast('设置已保存，但刷新设置失败，请稍后手动刷新页面', 'error')
  } finally {
    savingSettings.value = false
  }
}

async function changePassword(): Promise<void> {
  changingPassword.value = true
  try {
    await apiClient.changePassword(passwordForm.old_password, passwordForm.new_password)
    showToast('密码已修改，请重新登录', 'success')
    authStore.logout()
    await router.push('/login')
  } catch (err) {
    showToast(err instanceof ApiError ? err.message : '修改密码失败', 'error')
  } finally {
    changingPassword.value = false
    passwordConfirmOpen.value = false
  }
}

onMounted(load)
</script>

<template>
  <AppLayout>
    <h1 class="text-2xl font-semibold">设置</h1>
    <p class="mt-1 text-sm text-slate-400">更新日志保留和管理员密码。</p>

    <p v-if="error" class="mt-6 rounded-xl border border-red-500/40 bg-red-950 p-4 text-sm text-red-100">{{ error }}</p>
    <p v-else-if="loading" class="mt-6 text-sm text-slate-400">加载中...</p>

    <div v-else class="mt-6 grid gap-6">
      <form class="rounded-2xl border border-slate-800 bg-slate-900 p-6" @submit.prevent="saveSettings">
        <h2 class="text-lg font-semibold">服务设置</h2>
        <div class="mt-5 grid gap-5">
          <label class="grid gap-2 text-sm text-slate-300">
            日志保留天数
            <input v-model="settings.log_retention_days" min="1" type="number" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" required />
          </label>
        </div>
        <div class="mt-6 flex justify-end">
          <button class="rounded-xl bg-cyan-500 px-4 py-2 font-medium text-slate-950 hover:bg-cyan-400 disabled:opacity-60" :disabled="savingSettings">
            {{ savingSettings ? '保存中...' : '保存设置' }}
          </button>
        </div>
      </form>

      <form class="rounded-2xl border border-slate-800 bg-slate-900 p-6" @submit.prevent="passwordConfirmOpen = true">
        <h2 class="text-lg font-semibold">修改密码</h2>
        <div class="mt-5 grid gap-5">
          <label class="grid gap-2 text-sm text-slate-300">
            当前密码
            <input v-model="passwordForm.old_password" type="password" autocomplete="current-password" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" required />
          </label>
          <label class="grid gap-2 text-sm text-slate-300">
            新密码
            <input v-model="passwordForm.new_password" type="password" autocomplete="new-password" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" required />
          </label>
        </div>
        <div class="mt-6 flex justify-end">
          <button class="rounded-xl border border-amber-700 px-4 py-2 text-amber-200 hover:bg-amber-950">
            修改密码
          </button>
        </div>
      </form>
    </div>

    <ConfirmDialog
      :open="passwordConfirmOpen"
      title="修改管理员密码"
      message="修改成功后当前登录状态会被清除，需要重新登录。"
      confirm-text="修改"
      danger
      :busy="changingPassword"
      @cancel="passwordConfirmOpen = false"
      @confirm="changePassword"
    />
  </AppLayout>
</template>
