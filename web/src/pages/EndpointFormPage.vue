<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError, apiClient } from '@/api/client'
import type { EndpointInput } from '@/api/types'
import AppLayout from '@/components/AppLayout.vue'
import { showToast } from '@/components/toast'

const route = useRoute()
const router = useRouter()
const endpointID = computed(() => Number(route.params.id || 0))
const isEdit = computed(() => endpointID.value > 0)
const loading = ref(false)
const saving = ref(false)
const error = ref('')

const form = reactive<EndpointInput>({
  name: '',
  meow_nickname: '',
  default_title: '',
  msg_type: 'text',
  html_height: 200,
  default_url: '',
  default_img_url: '',
  active: true,
})

async function load(): Promise<void> {
  if (!isEdit.value) {
    return
  }
  loading.value = true
  error.value = ''
  try {
    const endpoint = await apiClient.getEndpoint(endpointID.value)
    form.name = endpoint.name
    form.meow_nickname = endpoint.meow_nickname
    form.default_title = endpoint.default_title
    form.msg_type = endpoint.msg_type
    form.html_height = endpoint.html_height
    form.default_url = endpoint.default_url
    form.default_img_url = endpoint.default_img_url
    form.active = endpoint.active
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '加载 endpoint 失败'
  } finally {
    loading.value = false
  }
}

async function submit(): Promise<void> {
  error.value = ''
  saving.value = true
  try {
    if (isEdit.value) {
      await apiClient.updateEndpoint(endpointID.value, {
        name: form.name,
        default_title: form.default_title,
        msg_type: form.msg_type,
        html_height: form.html_height,
        default_url: form.default_url,
        default_img_url: form.default_img_url,
        active: form.active,
      })
      showToast('Endpoint 已更新', 'success')
    } else {
      await apiClient.createEndpoint(form)
      showToast('Endpoint 已创建', 'success')
    }
    await router.push('/endpoints')
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : '保存 endpoint 失败'
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <AppLayout>
    <button class="mb-6 text-sm text-slate-400 hover:text-slate-100" @click="router.push('/endpoints')">← 返回 Endpoint</button>
    <h1 class="text-2xl font-semibold">{{ isEdit ? '编辑 Endpoint' : '新建 Endpoint' }}</h1>
    <p class="mt-1 text-sm text-slate-400">Endpoint 会生成标准 Webhook URL，外部服务可直接填写。</p>

    <p v-if="error" class="mt-6 rounded-xl border border-red-500/40 bg-red-950 p-4 text-sm text-red-100">{{ error }}</p>
    <p v-if="loading" class="mt-6 text-sm text-slate-400">加载中...</p>
    <div v-else-if="error" class="mt-6">
      <button class="rounded-xl border border-slate-700 px-4 py-2" @click="router.push('/endpoints')">返回 Endpoint 列表</button>
    </div>

    <form v-else class="mt-6 grid gap-5 rounded-2xl border border-slate-800 bg-slate-900 p-6" @submit.prevent="submit">
      <label class="grid gap-2 text-sm text-slate-300">
        名称
        <input v-model="form.name" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" required />
      </label>

      <label class="grid gap-2 text-sm text-slate-300">
        MeoW nickname
        <input
          v-model="form.meow_nickname"
          class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 disabled:opacity-60"
          :disabled="isEdit"
          required
        />
        <span v-if="isEdit" class="text-xs text-slate-500">创建后不可修改。</span>
      </label>

      <label class="grid gap-2 text-sm text-slate-300">
        默认标题
        <input v-model="form.default_title" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" />
      </label>

      <div class="grid gap-5 md:grid-cols-2">
        <label class="grid gap-2 text-sm text-slate-300">
          消息类型
          <select v-model="form.msg_type" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100">
            <option value="text">text</option>
            <option value="html">html</option>
            <option value="markdown">markdown</option>
          </select>
        </label>

        <label class="grid gap-2 text-sm text-slate-300">
          HTML 高度
          <input v-model.number="form.html_height" min="1" type="number" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" />
        </label>
      </div>

      <label class="grid gap-2 text-sm text-slate-300">
        默认跳转 URL
        <input v-model="form.default_url" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" />
      </label>

      <label class="grid gap-2 text-sm text-slate-300">
        默认图标 URL
        <input v-model="form.default_img_url" class="rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100" />
      </label>

      <label class="flex items-center gap-3 text-sm text-slate-300">
        <input v-model="form.active" type="checkbox" class="h-4 w-4" />
        启用 endpoint
      </label>

      <div class="flex justify-end gap-3">
        <button type="button" class="rounded-xl border border-slate-700 px-4 py-2" @click="router.push('/endpoints')">取消</button>
        <button class="rounded-xl bg-cyan-500 px-4 py-2 font-medium text-slate-950 hover:bg-cyan-400 disabled:opacity-60" :disabled="saving">
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </form>
  </AppLayout>
</template>
