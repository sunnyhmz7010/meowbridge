<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError, apiClient } from '@/api/client'
import type { PushLogListItem } from '@/api/types'
import AppLayout from '@/components/AppLayout.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import EmptyState from '@/components/EmptyState.vue'
import { showToast } from '@/components/toast'

const router = useRouter()
const logs = ref<PushLogListItem[]>([])
const loading = ref(true)
const error = ref('')
const cleanupOpen = ref(false)
const cleanupBusy = ref(false)

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    logs.value = await apiClient.listPushLogs()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '加载日志失败'
  } finally {
    loading.value = false
  }
}

async function cleanup(): Promise<void> {
  cleanupBusy.value = true
  try {
    const result = await apiClient.cleanupPushLogs()
    showToast(`已清理 ${result.deleted} 条日志`, 'success')
    cleanupOpen.value = false
    await load()
  } catch (err) {
    showToast(err instanceof ApiError ? err.message : '清理日志失败', 'error')
  } finally {
    cleanupBusy.value = false
  }
}

function formatTime(value: string): string {
  return value ? new Date(value).toLocaleString() : '-'
}

onMounted(load)
</script>

<template>
  <AppLayout>
    <div class="flex flex-col justify-between gap-4 border-b pb-6 lg:flex-row lg:items-end" style="border-color: var(--border);">
      <div>
        <p class="app-muted text-sm uppercase tracking-[0.22em]">Delivery Logs</p>
        <h1 class="app-heading mt-2 text-4xl font-black tracking-tight">推送日志</h1>
        <p class="app-muted mt-2 text-sm">查看 Webhook 解析结果、MeoW 响应和失败原因。</p>
      </div>
      <button class="app-button-danger" @click="cleanupOpen = true">
        按保留天数清理
      </button>
    </div>

    <div v-if="!loading && !error" class="mt-6 grid gap-4 md:grid-cols-3">
      <section class="app-card p-5">
        <p class="app-muted text-xs font-bold uppercase tracking-[0.16em]">日志总数</p>
        <p class="mt-2 text-4xl font-black">{{ logs.length }}</p>
      </section>
      <section class="app-card p-5">
        <p class="app-muted text-xs font-bold uppercase tracking-[0.16em]">成功</p>
        <p class="mt-2 text-4xl font-black" style="color: var(--success);">{{ logs.filter((log) => log.success).length }}</p>
      </section>
      <section class="app-card p-5">
        <p class="app-muted text-xs font-bold uppercase tracking-[0.16em]">失败</p>
        <p class="mt-2 text-4xl font-black" style="color: var(--danger);">{{ logs.filter((log) => !log.success).length }}</p>
      </section>
    </div>

    <p v-if="error" class="mt-6 rounded-xl border p-4 text-sm" style="border-color: color-mix(in srgb, var(--danger) 40%, transparent); background: var(--danger-soft); color: var(--danger);">{{ error }}</p>
    <p v-else-if="loading" class="app-muted mt-6 text-sm">加载中...</p>

    <template v-if="!loading && !error">
      <EmptyState v-if="logs.length === 0" class="mt-6" title="暂无推送日志" description="收到 Webhook 请求后，这里会显示推送记录。" />

      <div v-else class="app-card mt-6 overflow-hidden">
        <table class="app-table text-sm">
          <thead>
            <tr>
              <th class="px-4 py-3">时间</th>
              <th class="px-4 py-3">Endpoint</th>
              <th class="px-4 py-3">来源</th>
              <th class="px-4 py-3">标题</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td class="app-muted px-4 py-3">{{ formatTime(log.created_at) }}</td>
              <td class="px-4 py-3">{{ log.endpoint_name }}</td>
              <td class="app-muted px-4 py-3">{{ log.source_type }}</td>
              <td class="px-4 py-3">
                <p class="max-w-sm truncate">{{ log.parsed_title || '-' }}</p>
                <p class="app-muted mt-1 max-w-sm truncate text-xs">{{ log.parsed_msg }}</p>
              </td>
              <td class="px-4 py-3">
                <span class="app-badge" :class="log.success ? 'app-badge-success' : 'app-badge-danger'">
                  {{ log.success ? '成功' : '失败' }} / {{ log.meow_status_code || '-' }}
                </span>
                <p v-if="log.error_message" class="mt-1 max-w-xs truncate text-xs" style="color: var(--danger);">{{ log.error_message }}</p>
              </td>
              <td class="px-4 py-3">
                <button class="app-button-secondary px-3 py-1.5 text-sm" @click="router.push(`/logs/${log.id}`)">
                  详情
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <ConfirmDialog
      :open="cleanupOpen"
      title="清理推送日志"
      message="将按当前 log_retention_days 设置删除过期日志。该操作不可恢复。"
      confirm-text="清理"
      danger
      :busy="cleanupBusy"
      @cancel="cleanupOpen = false"
      @confirm="cleanup"
    />
  </AppLayout>
</template>
