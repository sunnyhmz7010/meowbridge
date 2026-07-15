<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiError, apiClient } from '@/api/client'
import type { EndpointView } from '@/api/types'
import AppLayout from '@/components/AppLayout.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import EmptyState from '@/components/EmptyState.vue'
import { showToast } from '@/components/toast'

const router = useRouter()
const endpoints = ref<EndpointView[]>([])
const loading = ref(true)
const error = ref('')
const copyFallbackURL = ref('')
const actionBusy = ref(false)
const confirmState = ref<{
  open: boolean
  title: string
  message: string
  confirmText: string
  danger: boolean
  run: () => Promise<void>
}>({
  open: false,
  title: '',
  message: '',
  confirmText: '确认',
  danger: false,
  run: async () => {},
})

function webhookURL(endpoint: EndpointView): string {
  return `${window.location.origin}/webhook/${endpoint.token}`
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    endpoints.value = await apiClient.listEndpoints()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '加载 endpoint 失败'
  } finally {
    loading.value = false
  }
}

function ask(options: Omit<typeof confirmState.value, 'open'>): void {
  confirmState.value = { ...options, open: true }
}

async function runConfirmed(): Promise<void> {
  actionBusy.value = true
  try {
    await confirmState.value.run()
    confirmState.value.open = false
  } catch (err) {
    showToast(err instanceof ApiError ? err.message : '操作失败，请稍后重试', 'error')
  } finally {
    actionBusy.value = false
  }
}

async function copyWebhook(endpoint: EndpointView): Promise<void> {
  const url = webhookURL(endpoint)
  copyFallbackURL.value = ''
  try {
    if (!navigator.clipboard) {
      throw new Error('clipboard unavailable')
    }
    await navigator.clipboard.writeText(url)
    showToast('Webhook URL 已复制', 'success')
  } catch {
    copyFallbackURL.value = url
    showToast('无法自动复制，请手动复制页面中显示的 URL', 'error')
  }
}

function toggleActive(endpoint: EndpointView): void {
  ask({
    title: endpoint.active ? '停用 endpoint' : '启用 endpoint',
    message: endpoint.active
      ? `停用后，${endpoint.name} 的 Webhook 请求会被拒绝。`
      : `启用后，${endpoint.name} 将重新接收 Webhook 请求。`,
    confirmText: endpoint.active ? '停用' : '启用',
    danger: endpoint.active,
    run: async () => {
      await apiClient.setEndpointActive(endpoint.id, !endpoint.active)
      showToast('状态已更新', 'success')
      await load()
    },
  })
}

function resetToken(endpoint: EndpointView): void {
  ask({
    title: '重置 token',
    message: `重置后，${endpoint.name} 的旧 Webhook URL 会立即失效。`,
    confirmText: '重置',
    danger: true,
    run: async () => {
      await apiClient.resetEndpointToken(endpoint.id)
      copyFallbackURL.value = ''
      showToast('Token 已重置，请复制新的 Webhook URL', 'success')
      await load()
    },
  })
}

function deleteEndpoint(endpoint: EndpointView): void {
  ask({
    title: '删除 endpoint',
    message: `确定删除 ${endpoint.name} 吗？该操作不可恢复。`,
    confirmText: '删除',
    danger: true,
    run: async () => {
      await apiClient.deleteEndpoint(endpoint.id)
      copyFallbackURL.value = ''
      showToast('Endpoint 已删除', 'success')
      await load()
    },
  })
}

onMounted(load)
</script>

<template>
  <AppLayout>
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold">Endpoint</h1>
        <p class="mt-1 text-sm text-slate-400">管理外部服务可直接填写的标准 Webhook 入口。</p>
      </div>
      <button class="rounded-xl bg-cyan-500 px-4 py-2 font-medium text-slate-950 hover:bg-cyan-400" @click="router.push('/endpoints/new')">
        新建 Endpoint
      </button>
    </div>

    <p v-if="error" class="mt-6 rounded-xl border border-red-500/40 bg-red-950 p-4 text-sm text-red-100">{{ error }}</p>
    <p v-else-if="loading" class="mt-6 text-sm text-slate-400">加载中...</p>
    <div v-if="copyFallbackURL" class="mt-6 rounded-xl border border-amber-500/40 bg-amber-950 p-4 text-sm text-amber-100">
      <p>请手动复制 Webhook URL：</p>
      <code class="mt-2 block break-all rounded-lg bg-slate-950 p-3 text-amber-50">{{ copyFallbackURL }}</code>
    </div>

    <template v-if="!loading && !error">
      <EmptyState
        v-if="endpoints.length === 0"
        class="mt-6"
        title="还没有 endpoint"
        description="创建第一个 endpoint 后，即可复制 Webhook URL 到外部服务。"
      >
        <button class="rounded-xl bg-cyan-500 px-4 py-2 font-medium text-slate-950 hover:bg-cyan-400" @click="router.push('/endpoints/new')">
          新建 Endpoint
        </button>
      </EmptyState>

      <div v-else class="mt-6 overflow-hidden rounded-2xl border border-slate-800">
        <table class="w-full border-collapse text-left text-sm">
          <thead class="bg-slate-900 text-slate-300">
            <tr>
              <th class="px-4 py-3">名称</th>
              <th class="px-4 py-3">MeoW nickname</th>
              <th class="px-4 py-3">消息类型</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800 bg-slate-950">
            <tr v-for="endpoint in endpoints" :key="endpoint.id">
              <td class="px-4 py-3">
                <p class="font-medium">{{ endpoint.name }}</p>
                <p class="mt-1 max-w-xs truncate text-xs text-slate-500">{{ endpoint.default_title || '无默认标题' }}</p>
              </td>
              <td class="px-4 py-3 text-slate-300">{{ endpoint.meow_nickname }}</td>
              <td class="px-4 py-3 text-slate-300">{{ endpoint.msg_type }}</td>
              <td class="px-4 py-3">
                <span class="rounded-full px-2 py-1 text-xs" :class="endpoint.active ? 'bg-emerald-950 text-emerald-200' : 'bg-slate-800 text-slate-300'">
                  {{ endpoint.active ? '启用' : '停用' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-2">
                  <button class="rounded-lg border border-slate-700 px-3 py-1.5 hover:bg-slate-800" @click="copyWebhook(endpoint)">复制 URL</button>
                  <button class="rounded-lg border border-slate-700 px-3 py-1.5 hover:bg-slate-800" @click="router.push(`/endpoints/${endpoint.id}`)">编辑</button>
                  <button class="rounded-lg border border-slate-700 px-3 py-1.5 hover:bg-slate-800" @click="toggleActive(endpoint)">
                    {{ endpoint.active ? '停用' : '启用' }}
                  </button>
                  <button class="rounded-lg border border-amber-700 px-3 py-1.5 text-amber-200 hover:bg-amber-950" @click="resetToken(endpoint)">重置 token</button>
                  <button class="rounded-lg border border-red-700 px-3 py-1.5 text-red-200 hover:bg-red-950" @click="deleteEndpoint(endpoint)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <ConfirmDialog
      :open="confirmState.open"
      :title="confirmState.title"
      :message="confirmState.message"
      :confirm-text="confirmState.confirmText"
      :danger="confirmState.danger"
      :busy="actionBusy"
      @cancel="confirmState.open = false"
      @confirm="runConfirmed"
    />
  </AppLayout>
</template>
