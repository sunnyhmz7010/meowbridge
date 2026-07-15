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
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold">推送日志</h1>
        <p class="mt-1 text-sm text-slate-400">查看 Webhook 解析结果、MeoW 响应和失败原因。</p>
      </div>
      <button class="rounded-xl border border-red-700 px-4 py-2 text-red-200 hover:bg-red-950" @click="cleanupOpen = true">
        按保留天数清理
      </button>
    </div>

    <p v-if="error" class="mt-6 rounded-xl border border-red-500/40 bg-red-950 p-4 text-sm text-red-100">{{ error }}</p>
    <p v-else-if="loading" class="mt-6 text-sm text-slate-400">加载中...</p>

    <template v-if="!loading && !error">
      <EmptyState v-if="logs.length === 0" class="mt-6" title="暂无推送日志" description="收到 Webhook 请求后，这里会显示推送记录。" />

      <div v-else class="mt-6 overflow-hidden rounded-2xl border border-slate-800">
        <table class="w-full border-collapse text-left text-sm">
          <thead class="bg-slate-900 text-slate-300">
            <tr>
              <th class="px-4 py-3">时间</th>
              <th class="px-4 py-3">Endpoint</th>
              <th class="px-4 py-3">来源</th>
              <th class="px-4 py-3">标题</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800 bg-slate-950">
            <tr v-for="log in logs" :key="log.id">
              <td class="px-4 py-3 text-slate-400">{{ formatTime(log.created_at) }}</td>
              <td class="px-4 py-3">{{ log.endpoint_name }}</td>
              <td class="px-4 py-3 text-slate-300">{{ log.source_type }}</td>
              <td class="px-4 py-3">
                <p class="max-w-sm truncate">{{ log.parsed_title || '-' }}</p>
                <p class="mt-1 max-w-sm truncate text-xs text-slate-500">{{ log.parsed_msg }}</p>
              </td>
              <td class="px-4 py-3">
                <span class="rounded-full px-2 py-1 text-xs" :class="log.success ? 'bg-emerald-950 text-emerald-200' : 'bg-red-950 text-red-200'">
                  {{ log.success ? '成功' : '失败' }} / {{ log.meow_status_code || '-' }}
                </span>
                <p v-if="log.error_message" class="mt-1 max-w-xs truncate text-xs text-red-300">{{ log.error_message }}</p>
              </td>
              <td class="px-4 py-3">
                <button class="rounded-lg border border-slate-700 px-3 py-1.5 hover:bg-slate-800" @click="router.push(`/logs/${log.id}`)">
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
