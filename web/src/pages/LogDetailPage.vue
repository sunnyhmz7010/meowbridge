<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError, apiClient } from '@/api/client'
import AppLayout from '@/components/AppLayout.vue'

const route = useRoute()
const router = useRouter()
const log = ref<Awaited<ReturnType<typeof apiClient.getPushLog>> | null>(null)
const loading = ref(true)
const error = ref('')

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    log.value = await apiClient.getPushLog(Number(route.params.id))
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '加载日志详情失败'
  } finally {
    loading.value = false
  }
}

function formatBlock(value: string): string {
  if (!value) {
    return '-'
  }
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

onMounted(load)
</script>

<template>
  <AppLayout>
    <div class="app-page-header">
      <button class="app-button-ghost mb-5 text-sm" @click="router.push('/logs')">← 返回日志</button>
      <p class="app-muted text-sm uppercase tracking-[0.22em]">Log Detail</p>
      <h1 class="app-heading mt-2 text-4xl font-black tracking-tight">日志详情</h1>
      <p class="app-muted mt-2 text-sm">查看单次 Webhook 的解析结果、原始请求和 MeoW 响应。</p>
    </div>

    <p class="app-alert-warning mt-6">日志 payload 可能包含敏感信息，仅在可信环境查看。</p>

    <p v-if="error" class="app-alert-danger mt-6">{{ error }}</p>
    <p v-else-if="loading" class="app-muted mt-6 text-sm">加载中...</p>

    <section v-else-if="log" class="mt-6 grid gap-6">
      <div class="app-card p-6">
        <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-start">
          <div>
            <p class="app-section-kicker">Summary</p>
            <h2 class="app-heading mt-2 text-xl font-black">概要</h2>
          </div>
          <span class="app-badge w-fit" :class="log.success ? 'app-badge-success' : 'app-badge-danger'">
            {{ log.success ? '成功' : '失败' }}
          </span>
        </div>
        <dl class="mt-5 grid gap-4 text-sm md:grid-cols-2 xl:grid-cols-3">
          <div class="app-card-muted p-4"><dt class="app-muted text-xs">Endpoint</dt><dd class="mt-1 font-semibold">{{ log.endpoint_name }}</dd></div>
          <div class="app-card-muted p-4"><dt class="app-muted text-xs">Source Type</dt><dd class="mt-1 font-semibold">{{ log.source_type }}</dd></div>
          <div class="app-card-muted p-4"><dt class="app-muted text-xs">请求方法</dt><dd class="mt-1 font-semibold">{{ log.request_method || '-' }}</dd></div>
          <div class="app-card-muted p-4"><dt class="app-muted text-xs">MeoW 状态码</dt><dd class="mt-1 font-semibold">{{ log.meow_status_code || '-' }}</dd></div>
          <div class="app-card-muted p-4"><dt class="app-muted text-xs">时间</dt><dd class="mt-1 font-semibold">{{ log.created_at ? new Date(log.created_at).toLocaleString() : '-' }}</dd></div>
          <div class="app-card-muted p-4"><dt class="app-muted text-xs">错误</dt><dd class="mt-1 font-semibold">{{ log.error_message || '-' }}</dd></div>
        </dl>
      </div>

      <div class="app-card p-6">
        <p class="app-section-kicker">Parsed Result</p>
        <h2 class="app-heading mt-2 text-xl font-black">解析结果</h2>
        <dl class="mt-5 grid gap-4 text-sm">
          <div class="grid gap-1">
            <dt class="app-muted text-xs">标题</dt>
            <dd class="font-semibold">{{ log.parsed_title || '-' }}</dd>
          </div>
          <div class="grid gap-1">
            <dt class="app-muted text-xs">消息类型</dt>
            <dd>{{ log.parsed_msg_type || '-' }}</dd>
          </div>
          <div class="grid gap-1">
            <dt class="app-muted text-xs">消息</dt>
            <dd class="app-card-muted whitespace-pre-wrap p-4">{{ log.parsed_msg || '-' }}</dd>
          </div>
        </dl>
      </div>

      <div class="grid gap-6 lg:grid-cols-2">
        <section class="app-card p-6">
          <p class="app-section-kicker">Request</p>
          <h2 class="app-heading mt-2 text-lg font-black">Headers</h2>
          <pre class="app-code-block mt-4 overflow-auto p-4 text-xs">{{ formatBlock(log.request_headers) }}</pre>
        </section>
        <section class="app-card p-6">
          <p class="app-section-kicker">Request</p>
          <h2 class="app-heading mt-2 text-lg font-black">Query</h2>
          <pre class="app-code-block mt-4 overflow-auto p-4 text-xs">{{ formatBlock(log.request_query) }}</pre>
        </section>
        <section class="app-card p-6">
          <p class="app-section-kicker">Request</p>
          <h2 class="app-heading mt-2 text-lg font-black">Payload</h2>
          <pre class="app-code-block mt-4 max-h-96 overflow-auto p-4 text-xs">{{ formatBlock(log.request_payload) }}</pre>
        </section>
        <section class="app-card p-6">
          <p class="app-section-kicker">Response</p>
          <h2 class="app-heading mt-2 text-lg font-black">MeoW 响应</h2>
          <pre class="app-code-block mt-4 max-h-96 overflow-auto p-4 text-xs">{{ formatBlock(log.meow_response_body) }}</pre>
        </section>
      </div>
    </section>
  </AppLayout>
</template>
