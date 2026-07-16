<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiError, apiClient, defaultParserConfig } from '@/api/client'
import type { EndpointInput, MsgType, ParserConfig, WebhookPreset, WebhookPreviewResult } from '@/api/types'
import AppLayout from '@/components/AppLayout.vue'
import { showToast } from '@/components/toast'

const route = useRoute()
const router = useRouter()
const endpointID = computed(() => Number(route.params.id || 0))
const isEdit = computed(() => endpointID.value > 0)
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const presets = ref<WebhookPreset[]>([])
const previewPayload = ref('')
const previewResult = ref<WebhookPreviewResult | null>(null)
const previewError = ref('')
const previewing = ref(false)

const parserFields = reactive({
  title: '',
  msg: '',
  url: '',
  img_url: '',
  msg_type: 'markdown' as MsgType,
})

const form = reactive<EndpointInput>({
  name: '',
  meow_nickname: '',
  default_title: '',
  msg_type: 'text',
  html_height: 200,
  default_url: '',
  default_img_url: '',
  parser_config: defaultParserConfig(),
  active: true,
})

const defaultPreviewPayload = JSON.stringify({
  sourcecontrol: 'github',
  service: 'github',
  event_type: 'push',
  hook: {
    url: 'https://github.com/sunnyhmz7010/meowbridge',
  },
  ref: 'refs/heads/main',
}, null, 2)

function linesToFragments(value: string): string[] {
  return value.split('\n').map((line) => line.replace(/\r$/, '')).filter((line) => line !== '')
}

function fragmentsToText(value: string[] | undefined): string {
  return Array.isArray(value) ? value.join('\n') : ''
}

function buildParserConfig(): ParserConfig {
  if (form.parser_config.mode === 'auto') {
    return defaultParserConfig()
  }
  if (form.parser_config.mode === 'preset') {
    return {
      mode: 'preset',
      preset: form.parser_config.preset || 'github_push_minimal',
      field_mapping: {},
      default_values: {},
    }
  }
  return {
    mode: 'custom',
    preset: 'generic',
    field_mapping: {
      title: linesToFragments(parserFields.title),
      msg: linesToFragments(parserFields.msg),
      url: linesToFragments(parserFields.url),
      img_url: linesToFragments(parserFields.img_url),
      msg_type: [parserFields.msg_type],
    },
    default_values: {
      msg_type: parserFields.msg_type,
    },
  }
}

function applyParserConfig(config: ParserConfig): void {
  form.parser_config = config
  parserFields.title = fragmentsToText(config.field_mapping.title)
  parserFields.msg = fragmentsToText(config.field_mapping.msg)
  parserFields.url = fragmentsToText(config.field_mapping.url)
  parserFields.img_url = fragmentsToText(config.field_mapping.img_url)
  parserFields.msg_type = (config.default_values.msg_type as MsgType | undefined) || 'markdown'
}

async function loadPresets(): Promise<void> {
  try {
    presets.value = await apiClient.getWebhookPresets()
  } catch (err) {
    showToast(err instanceof ApiError ? err.message : '加载解析器预设失败', 'error')
  }
}

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
    applyParserConfig(endpoint.parser_config)
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
    form.parser_config = buildParserConfig()
    if (isEdit.value) {
      await apiClient.updateEndpoint(endpointID.value, {
        name: form.name,
        default_title: form.default_title,
        msg_type: form.msg_type,
        html_height: form.html_height,
        default_url: form.default_url,
        default_img_url: form.default_img_url,
        parser_config: form.parser_config,
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

async function previewParser(): Promise<void> {
  previewError.value = ''
  previewResult.value = null
  let payload: unknown
  try {
    payload = JSON.parse(previewPayload.value)
  } catch {
    previewError.value = '测试 payload 不是合法 JSON'
    return
  }

  previewing.value = true
  try {
    previewResult.value = await apiClient.previewWebhook({
      parser_config: buildParserConfig(),
      payload,
    })
  } catch (err) {
    previewError.value = err instanceof ApiError ? err.message : '解析预览失败'
  } finally {
    previewing.value = false
  }
}

onMounted(() => {
  previewPayload.value = defaultPreviewPayload
  void load()
  void loadPresets()
})
</script>

<template>
  <AppLayout>
    <div class="app-page-header flex flex-col justify-between gap-4 lg:flex-row lg:items-end">
      <div>
        <button class="app-button-ghost mb-5 text-sm" @click="router.push('/endpoints')">← 返回 Endpoint</button>
        <p class="app-muted text-sm uppercase tracking-[0.22em]">Endpoint Form</p>
        <h1 class="app-heading mt-2 text-4xl font-black tracking-tight">{{ isEdit ? '编辑 Endpoint' : '新建 Endpoint' }}</h1>
        <p class="app-muted mt-2 text-sm">配置外部服务的 Webhook 入口、默认消息和 payload 解析方式。</p>
      </div>
      <span class="app-badge app-badge-muted w-fit">{{ isEdit ? 'EDIT MODE' : 'CREATE MODE' }}</span>
    </div>

    <p v-if="error" class="app-alert-danger mt-6">{{ error }}</p>
    <p v-if="loading" class="app-muted mt-6 text-sm">加载中...</p>
    <div v-else-if="error" class="mt-6">
      <button class="app-button-secondary" @click="router.push('/endpoints')">返回 Endpoint 列表</button>
    </div>

    <form v-else class="mt-6 grid gap-6" @submit.prevent="submit">
      <section class="app-section-grid">
        <div>
          <p class="app-section-kicker">Identity</p>
          <h2 class="app-heading mt-2 text-xl font-black">入口标识</h2>
          <p class="app-muted mt-2 text-sm leading-6">名称用于后台识别，MeoW nickname 用于指定接收方，创建后保持稳定。</p>
        </div>
        <div class="app-card grid gap-5 p-6">
          <div class="grid gap-5 lg:grid-cols-2">
            <label class="app-muted grid gap-2 text-sm">
              名称
              <input v-model="form.name" class="app-input" required />
            </label>

            <label class="app-muted grid gap-2 text-sm">
              MeoW nickname
              <input
                v-model="form.meow_nickname"
                class="app-input disabled:opacity-60"
                :disabled="isEdit"
                required
              />
              <span v-if="isEdit" class="app-muted text-xs">创建后不可修改。</span>
            </label>
          </div>
        </div>
      </section>

      <section class="app-section-grid">
        <div>
          <p class="app-section-kicker">Message Defaults</p>
          <h2 class="app-heading mt-2 text-xl font-black">默认消息</h2>
          <p class="app-muted mt-2 text-sm leading-6">当 Webhook payload 缺少可解析内容时，使用这里的兜底标题、类型和展示参数。</p>
        </div>
        <div class="app-card grid gap-5 p-6">
          <label class="app-muted grid gap-2 text-sm">
            默认标题
            <input v-model="form.default_title" class="app-input" />
          </label>

          <div class="grid gap-5 md:grid-cols-2">
            <label class="app-muted grid gap-2 text-sm">
              消息类型
              <select v-model="form.msg_type" class="app-input">
                <option value="text">text</option>
                <option value="html">html</option>
                <option value="markdown">markdown</option>
              </select>
            </label>

            <label class="app-muted grid gap-2 text-sm">
              HTML 高度
              <input v-model.number="form.html_height" min="1" type="number" class="app-input" />
            </label>
          </div>
        </div>
      </section>

      <section class="app-section-grid">
        <div>
          <p class="app-section-kicker">Links</p>
          <h2 class="app-heading mt-2 text-xl font-black">跳转与图标</h2>
          <p class="app-muted mt-2 text-sm leading-6">为消息补充默认跳转地址和图标地址，解析器输出为空时使用。</p>
        </div>
        <div class="app-card grid gap-5 p-6">
          <label class="app-muted grid gap-2 text-sm">
            默认跳转 URL
            <input v-model="form.default_url" class="app-input" />
          </label>

          <label class="app-muted grid gap-2 text-sm">
            默认图标 URL
            <input v-model="form.default_img_url" class="app-input" />
          </label>
        </div>
      </section>

      <section class="app-section-grid">
        <div>
          <p class="app-section-kicker">Parser</p>
          <h2 class="app-heading mt-2 text-xl font-black">Webhook 解析</h2>
          <p class="app-muted mt-2 text-sm leading-6">优先用预设解析器处理 GitHub 等常见 payload；字段不固定时再使用自定义映射。</p>
          <div class="app-card-muted mt-4 p-4 text-sm">
            <p class="font-semibold">当前推荐</p>
            <p class="app-muted mt-1 text-xs leading-5">GitHub 推送事件选择“预设解析器 / GitHub 简化 Push”，可把当前只显示 JSON 的 payload 转成可读通知。</p>
          </div>
        </div>

        <div class="app-card grid gap-6 p-6">
          <div class="grid gap-5 lg:grid-cols-2">
            <label class="app-muted grid gap-2 text-sm">
              解析模式
              <select v-model="form.parser_config.mode" class="app-input">
                <option value="auto">自动解析</option>
                <option value="preset">预设解析器</option>
                <option value="custom">自定义字段映射</option>
              </select>
            </label>

            <label v-if="form.parser_config.mode === 'preset'" class="app-muted grid gap-2 text-sm">
              预设解析器
              <select v-model="form.parser_config.preset" class="app-input">
                <option value="">选择预设</option>
                <option v-for="preset in presets" :key="preset.id" :value="preset.id">
                  {{ preset.name }}
                </option>
              </select>
              <span class="app-muted text-xs">推荐当前场景使用 GitHub 简化 Push。</span>
            </label>
          </div>

          <div v-if="form.parser_config.mode === 'custom'" class="app-card-muted grid gap-4 p-5">
            <div>
              <h3 class="app-heading font-semibold">自定义字段映射</h3>
              <p class="app-muted mt-1 text-xs">每行一个 JSONPath 或字面量，按顺序拼接为最终字段。</p>
            </div>
            <label class="app-muted grid gap-2 text-sm">
              标题字段
              <textarea v-model="parserFields.title" class="app-input min-h-20" placeholder="每行一个 JSONPath 或字面量，例如：GitHub:&#10;$.event_type" />
            </label>
            <label class="app-muted grid gap-2 text-sm">
              消息字段
              <textarea v-model="parserFields.msg" class="app-input min-h-28" placeholder="仓库: &#10;$.hook.url&#10;&#10;分支: &#10;$.ref" />
            </label>
            <div class="grid gap-4 md:grid-cols-3">
              <label class="app-muted grid gap-2 text-sm">
                URL 字段
                <input v-model="parserFields.url" class="app-input" placeholder="$.hook.url" />
              </label>
              <label class="app-muted grid gap-2 text-sm">
                图标字段
                <input v-model="parserFields.img_url" class="app-input" placeholder="$.icon" />
              </label>
              <label class="app-muted grid gap-2 text-sm">
                消息类型
                <select v-model="parserFields.msg_type" class="app-input">
                  <option value="text">text</option>
                  <option value="html">html</option>
                  <option value="markdown">markdown</option>
                </select>
              </label>
            </div>
          </div>

          <div class="grid gap-4 lg:grid-cols-2">
            <label class="app-muted grid gap-2 text-sm">
              测试 payload
              <textarea v-model="previewPayload" class="app-input min-h-64 font-mono text-xs" />
            </label>
            <div class="grid gap-3">
              <div class="flex items-center justify-between">
                <span class="app-muted text-sm font-semibold">解析预览</span>
                <button type="button" class="app-button-secondary px-3 py-1.5 text-sm" :disabled="previewing" @click="previewParser">
                  {{ previewing ? '解析中...' : '预览' }}
                </button>
              </div>
              <p v-if="previewError" class="app-alert-danger p-3">{{ previewError }}</p>
              <pre v-if="previewResult" class="app-code-block min-h-64 overflow-auto p-4 text-xs">{{ JSON.stringify(previewResult, null, 2) }}</pre>
              <div v-else class="app-card-muted min-h-64 p-4 text-sm">
                <p class="font-semibold">等待预览</p>
                <p class="app-muted mt-2 leading-6">粘贴实际 Webhook JSON，点击预览确认标题、消息、跳转 URL 和消息类型。</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div class="app-card flex flex-col gap-4 p-5 md:flex-row md:items-center md:justify-between">
        <label class="app-muted flex items-center gap-3 text-sm">
          <input v-model="form.active" type="checkbox" class="h-4 w-4" />
          启用 endpoint
        </label>

        <div class="flex justify-end gap-3">
          <button type="button" class="app-button-secondary" @click="router.push('/endpoints')">取消</button>
          <button class="app-button-primary disabled:opacity-60" :disabled="saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </form>
  </AppLayout>
</template>
