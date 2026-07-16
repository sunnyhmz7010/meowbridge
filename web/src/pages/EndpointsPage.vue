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

function parserLabel(endpoint: EndpointView): string {
  const config = endpoint.parser_config
  if (config?.mode === 'preset') {
    return config.preset || '预设解析器'
  }
  if (config?.mode === 'custom') {
    return '自定义映射'
  }
  return '自动解析'
}

function parserMode(endpoint: EndpointView): string {
  return endpoint.parser_config?.mode || 'auto'
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
    <div class="flex flex-col justify-between gap-4 border-b pb-6 lg:flex-row lg:items-end" style="border-color: var(--border);">
      <div>
        <p class="app-muted text-sm uppercase tracking-[0.22em]">Webhook Endpoints</p>
        <h1 class="app-heading mt-2 text-4xl font-black tracking-tight">Endpoint</h1>
        <p class="app-muted mt-2 text-sm">管理外部服务可直接填写的标准 Webhook 入口。</p>
      </div>
      <button class="app-button-primary" @click="router.push('/endpoints/new')">
        + 新建 Endpoint
      </button>
    </div>

    <div v-if="!loading && !error" class="mt-6 grid gap-4 md:grid-cols-3">
      <section class="app-card p-5">
        <p class="app-muted text-xs font-bold uppercase tracking-[0.16em]">Endpoint 总数</p>
        <p class="mt-2 text-4xl font-black">{{ endpoints.length }}</p>
      </section>
      <section class="app-card p-5">
        <p class="app-muted text-xs font-bold uppercase tracking-[0.16em]">启用中</p>
        <p class="mt-2 text-4xl font-black" style="color: var(--success);">{{ endpoints.filter((endpoint) => endpoint.active).length }}</p>
      </section>
      <section class="app-card p-5">
        <p class="app-muted text-xs font-bold uppercase tracking-[0.16em]">已停用</p>
        <p class="mt-2 text-4xl font-black">{{ endpoints.filter((endpoint) => !endpoint.active).length }}</p>
      </section>
    </div>

    <p v-if="error" class="mt-6 rounded-xl border p-4 text-sm" style="border-color: color-mix(in srgb, var(--danger) 40%, transparent); background: var(--danger-soft); color: var(--danger);">{{ error }}</p>
    <p v-else-if="loading" class="app-muted mt-6 text-sm">加载中...</p>
    <div v-if="copyFallbackURL" class="mt-6 rounded-xl border p-4 text-sm" style="border-color: color-mix(in srgb, var(--warning) 40%, transparent); background: var(--warning-soft); color: var(--warning);">
      <p>请手动复制 Webhook URL：</p>
      <code class="app-code-block mt-2 block break-all p-3">{{ copyFallbackURL }}</code>
    </div>

    <template v-if="!loading && !error">
      <EmptyState
        v-if="endpoints.length === 0"
        class="mt-6"
        title="还没有 endpoint"
        description="创建第一个 endpoint 后，即可复制 Webhook URL 到外部服务。"
      >
        <button class="app-button-primary" @click="router.push('/endpoints/new')">
          新建 Endpoint
        </button>
      </EmptyState>

      <div v-else class="app-card mt-6 overflow-hidden">
        <table class="app-table text-sm">
          <thead>
            <tr>
              <th class="px-4 py-3">名称</th>
              <th class="px-4 py-3">MeoW nickname</th>
              <th class="px-4 py-3">消息类型</th>
              <th class="px-4 py-3">解析器</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="endpoint in endpoints" :key="endpoint.id">
              <td class="px-4 py-3">
                <p class="font-semibold">{{ endpoint.name }}</p>
                <p class="app-muted mt-1 max-w-xs truncate text-xs">{{ endpoint.default_title || '无默认标题' }}</p>
              </td>
              <td class="app-muted px-4 py-3">{{ endpoint.meow_nickname }}</td>
              <td class="app-muted px-4 py-3">{{ endpoint.msg_type }}</td>
              <td class="px-4 py-3">
                <span class="app-badge" :class="parserMode(endpoint) === 'auto' ? 'app-badge-muted' : 'app-badge-success'">
                  {{ parserLabel(endpoint) }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span class="app-badge" :class="endpoint.active ? 'app-badge-success' : 'app-badge-muted'">
                  {{ endpoint.active ? '启用' : '停用' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-2">
                  <button class="app-button-secondary px-3 py-1.5 text-sm" @click="copyWebhook(endpoint)">复制 URL</button>
                  <button class="app-button-secondary px-3 py-1.5 text-sm" @click="router.push(`/endpoints/${endpoint.id}`)">编辑</button>
                  <button class="app-button-secondary px-3 py-1.5 text-sm" @click="toggleActive(endpoint)">
                    {{ endpoint.active ? '停用' : '启用' }}
                  </button>
                  <button class="app-button-warning px-3 py-1.5 text-sm" @click="resetToken(endpoint)">重置 token</button>
                  <button class="app-button-danger px-3 py-1.5 text-sm" @click="deleteEndpoint(endpoint)">删除</button>
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
