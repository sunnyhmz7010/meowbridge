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
    <button class="app-button-ghost mb-6 text-sm" @click="router.push('/logs')">← 返回日志</button>
    <p class="app-muted text-sm uppercase tracking-[0.22em]">Log Detail</p>
    <h1 class="app-heading mt-2 text-3xl font-semibold tracking-tight">日志详情</h1>
    <p class="mt-2 text-sm" style="color: var(--warning);">日志 payload 可能包含敏感信息，仅在可信环境查看。</p>

    <p v-if="error" class="mt-6 rounded-xl border p-4 text-sm" style="border-color: color-mix(in srgb, var(--danger) 40%, transparent); background: var(--danger-soft); color: var(--danger);">{{ error }}</p>
    <p v-else-if="loading" class="app-muted mt-6 text-sm">加载中...</p>

    <section v-else-if="log" class="mt-6 grid gap-6">
      <div class="app-card rounded-3xl p-6">
        <h2 class="app-heading text-lg font-semibold">概要</h2>
        <dl class="mt-4 grid gap-3 text-sm md:grid-cols-2">
          <div><dt class="app-muted">Endpoint</dt><dd>{{ log.endpoint_name }}</dd></div>
          <div><dt class="app-muted">Source Type</dt><dd>{{ log.source_type }}</dd></div>
          <div><dt class="app-muted">请求方法</dt><dd>{{ log.request_method || '-' }}</dd></div>
          <div><dt class="app-muted">MeoW 状态码</dt><dd>{{ log.meow_status_code || '-' }}</dd></div>
          <div><dt class="app-muted">结果</dt><dd>{{ log.success ? '成功' : '失败' }}</dd></div>
          <div><dt class="app-muted">错误</dt><dd>{{ log.error_message || '-' }}</dd></div>
          <div><dt class="app-muted">时间</dt><dd>{{ log.created_at ? new Date(log.created_at).toLocaleString() : '-' }}</dd></div>
        </dl>
      </div>

      <div class="app-card rounded-3xl p-6">
        <h2 class="app-heading text-lg font-semibold">解析结果</h2>
        <dl class="mt-4 grid gap-3 text-sm">
          <div><dt class="app-muted">标题</dt><dd>{{ log.parsed_title || '-' }}</dd></div>
          <div><dt class="app-muted">消息类型</dt><dd>{{ log.parsed_msg_type || '-' }}</dd></div>
          <div><dt class="app-muted">消息</dt><dd class="whitespace-pre-wrap">{{ log.parsed_msg || '-' }}</dd></div>
        </dl>
      </div>

      <div class="grid gap-6 lg:grid-cols-2">
        <section class="app-card rounded-3xl p-6">
          <h2 class="app-heading text-lg font-semibold">Headers</h2>
          <pre class="app-code-block mt-4 overflow-auto p-4 text-xs">{{ formatBlock(log.request_headers) }}</pre>
        </section>
        <section class="app-card rounded-3xl p-6">
          <h2 class="app-heading text-lg font-semibold">Query</h2>
          <pre class="app-code-block mt-4 overflow-auto p-4 text-xs">{{ formatBlock(log.request_query) }}</pre>
        </section>
        <section class="app-card rounded-3xl p-6">
          <h2 class="app-heading text-lg font-semibold">Payload</h2>
          <pre class="app-code-block mt-4 max-h-96 overflow-auto p-4 text-xs">{{ formatBlock(log.request_payload) }}</pre>
        </section>
        <section class="app-card rounded-3xl p-6">
          <h2 class="app-heading text-lg font-semibold">MeoW 响应</h2>
          <pre class="app-code-block mt-4 max-h-96 overflow-auto p-4 text-xs">{{ formatBlock(log.meow_response_body) }}</pre>
        </section>
      </div>
    </section>
  </AppLayout>
</template>
