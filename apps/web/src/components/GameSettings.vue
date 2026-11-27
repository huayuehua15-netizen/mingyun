<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  clearAISettings,
  fetchAISettings,
  saveAISettings,
  testAISettings
} from '../api'
import type { AISettingsInput, AISettingsView, AIProviderPreset, AITestResult } from '../api'

const props = defineProps<{
  initial: AISettingsView | null
  initialError?: string
}>()

const emit = defineEmits<{
  (e: 'updated', value: AISettingsView): void
  (e: 'home'): void
}>()

const settings = ref<AISettingsView | null>(props.initial)
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const clearing = ref(false)
const confirmingClear = ref(false)
const error = ref(props.initialError || '')
const localError = ref('')
const savedNotice = ref('')
const testResult = ref<AITestResult | null>(null)

const selectedProvider = ref('')
const baseURL = ref('')
const model = ref('')
const apiKey = ref('')
const temperature = ref(0.5)
const maxTokens = ref(1200)
const customBase = ref(false)
const showBase = ref(false)
const showAdvanced = ref(false)

const presets = computed(() => settings.value?.presets ?? [])
const groups = computed(() =>
  [
    { key: 'cn', label: '国内直连', items: presets.value.filter((p) => p.region === 'cn') },
    { key: 'global', label: '海外接口', items: presets.value.filter((p) => p.region === 'global') },
    { key: 'local', label: '本机与自建', items: presets.value.filter((p) => p.region === 'local') },
    { key: 'custom', label: '自定', items: presets.value.filter((p) => p.region === 'custom') }
  ].filter((group) => group.items.length)
)
const activePreset = computed(() => presets.value.find((p) => p.id === selectedProvider.value))
const modelOptions = computed(() => activePreset.value?.models ?? [])
const writable = computed(() => !!settings.value?.writable)
const fieldDisabled = computed(() => !writable.value || saving.value || clearing.value)
const statusText = computed(() => (settings.value?.enabled ? '在线' : '离线'))
const sourceLabel = computed(() => {
  if (!settings.value) return '—'
  if (settings.value.source === 'file') return '网页配置'
  if (settings.value.source === 'env') return '环境变量'
  return '尚未接入'
})
const statusLine = computed(() => {
  if (!settings.value) return '正在读取模型配置…'
  if (!settings.value.enabled) return '排盘与卦历照常可用；解读先走离线摘要。'
  return `${settings.value.provider_name} · ${settings.value.model}`
})
const localEndpoint = computed(() =>
  /^(https?:\/\/)?(localhost|127\.0\.0\.1|\[::1\]|0\.0\.0\.0|host\.docker\.internal)([:/]|$)/i.test(
    baseURL.value.trim()
  )
)
const keyPlaceholder = computed(() => {
  if (settings.value?.has_key && settings.value.provider_id === selectedProvider.value) {
    return `已保存 ${settings.value.key_hint}，留空保持不变`
  }
  return localEndpoint.value ? '粘贴 API Key；本地端点可留空' : '粘贴 API Key'
})
const canSubmit = computed(() => !!baseURL.value.trim() && !!model.value.trim())
const testTone = computed(() => {
  if (!testResult.value) return ''
  return testResult.value.ok ? 'is-ok' : 'is-bad'
})

function applySettings(next: AISettingsView) {
  settings.value = next
  selectedProvider.value = next.provider_id || ''
  baseURL.value = next.base_url || ''
  model.value = next.model || ''
  temperature.value = next.temperature ?? 0.5
  maxTokens.value = next.max_tokens ?? 1200
  apiKey.value = ''
  customBase.value = next.provider_id === 'custom'
  const preset = next.presets.find((p) => p.id === next.provider_id)
  customBase.value = !preset || (!!next.base_url && next.base_url !== preset.base_url)
  showBase.value = customBase.value
  testResult.value = null
  savedNotice.value = ''
  error.value = ''
  localError.value = ''
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const next = await fetchAISettings()
    applySettings(next)
    emit('updated', next)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '读取模型配置失败'
  } finally {
    loading.value = false
  }
}

function selectProvider(preset: AIProviderPreset) {
  const changed = selectedProvider.value !== preset.id
  if (!changed) return
  selectedProvider.value = preset.id
  customBase.value = !preset.base_url
  showBase.value = !preset.base_url
  baseURL.value = preset.base_url
  model.value = preset.default_model || ''
  apiKey.value = ''
  testResult.value = null
  localError.value = ''
}

function buildInput(): AISettingsInput {
  const input: AISettingsInput = {
    provider_id: selectedProvider.value,
    base_url: baseURL.value.trim(),
    model: model.value.trim(),
    temperature: temperature.value,
    max_tokens: maxTokens.value
  }
  const key = apiKey.value.trim()
  if (key) input.api_key = key
  return input
}

async function save() {
  if (!canSubmit.value) {
    localError.value = '先选服务商，并填写模型名。'
    return
  }
  saving.value = true
  localError.value = ''
  try {
    const next = await saveAISettings(buildInput())
    applySettings(next)
    savedNotice.value = `已封存：${next.provider_name} · ${next.model}。解读立即走这炉火。`
    emit('updated', next)
  } catch (e) {
    localError.value = e instanceof Error ? e.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function test() {
  if (!canSubmit.value) {
    localError.value = '先选服务商，并填写模型名。'
    return
  }
  testing.value = true
  localError.value = ''
  savedNotice.value = ''
  testResult.value = null
  try {
    testResult.value = await testAISettings(buildInput())
  } catch (e) {
    localError.value = e instanceof Error ? e.message : '测试失败'
  } finally {
    testing.value = false
  }
}

async function removeConfig() {
  if (!confirmingClear.value) {
    confirmingClear.value = true
    return
  }
  clearing.value = true
  localError.value = ''
  try {
    const next = await clearAISettings()
    applySettings(next)
    emit('updated', next)
    confirmingClear.value = false
  } catch (e) {
    localError.value = e instanceof Error ? e.message : '撤去失败'
  } finally {
    clearing.value = false
  }
}

watch([baseURL, model, apiKey], () => {
  testResult.value = null
})

onMounted(() => {
  if (props.initial) applySettings(props.initial)
  else void load()
})
</script>

<template>
  <section class="scene-view settings-view">
    <span class="view-watermark" aria-hidden="true">机</span>
    <header class="scene-bar">
      <button class="scene-back" type="button" @click="$emit('home')">← 山门</button>
      <div class="scene-heading">
        <span>墨匣 · 模型接入</span>
        <h1>接一炉自己的火</h1>
      </div>
      <div class="settings-status" :class="{ 'is-online': settings?.enabled }">
        <i></i>
        <b>{{ statusText }}</b>
        <small>{{ sourceLabel }}</small>
      </div>
    </header>

    <div class="scene-body settings-body">
      <p v-if="error" class="settings-error">{{ error }}</p>

      <section v-if="settings" class="settings-current">
        <div class="settings-seal" :class="{ 'is-online': settings.enabled }">
          <b>{{ settings.enabled ? '启' : '止' }}</b>
          <span>模型</span>
        </div>
        <div class="settings-current-copy">
          <span class="settings-kicker">当前炉火</span>
          <h2>{{ settings.enabled ? settings.provider_name : '尚未接入模型' }}</h2>
          <p>{{ statusLine }}</p>
          <p v-if="settings.enabled" class="settings-endpoint-line">{{ settings.base_url }}</p>
        </div>
        <dl class="settings-current-facts">
          <div><dt>来源</dt><dd>{{ sourceLabel }}</dd></div>
          <div><dt>Key</dt><dd>{{ settings.has_key ? settings.key_hint : '未保存' }}</dd></div>
          <div><dt>温度</dt><dd>{{ settings.temperature }}</dd></div>
          <div><dt>输出上限</dt><dd>{{ settings.max_tokens }}</dd></div>
        </dl>
      </section>

      <p v-if="loading" class="settings-loading">正在读取模型配置…</p>

      <template v-if="settings">
        <section class="settings-section">
          <header class="settings-section-head">
            <span>一 · 择炉</span>
            <h2>选一家服务商</h2>
            <p>选中后自动填好接口地址；模型名可以自由修改。</p>
          </header>
          <div v-for="group in groups" :key="group.key" class="provider-group">
            <h3>{{ group.label }}</h3>
            <div class="provider-grid">
              <button
                v-for="preset in group.items"
                :key="preset.id"
                type="button"
                class="provider-card"
                :class="{ 'is-active': selectedProvider === preset.id }"
                :disabled="!writable || saving"
                @click="selectProvider(preset)"
              >
                <b>{{ preset.name }}</b>
                <small>{{ preset.note || preset.id }}</small>
                <em v-if="preset.default_model">{{ preset.default_model }}</em>
              </button>
            </div>
          </div>
        </section>

        <section class="settings-section">
          <header class="settings-section-head">
            <span>二 · 接线</span>
            <h2>Key 与模型名</h2>
            <p>Key 只提交给本机服务端，保存后不再回显。</p>
          </header>
          <div class="settings-form">
            <label class="settings-field">
              <span>
                API Key
                <em v-if="settings.has_key && settings.provider_id === selectedProvider">已存 {{ settings.key_hint }}</em>
                <em v-else-if="settings.has_key">换服务商后不沿用旧 Key</em>
              </span>
              <input
                v-model="apiKey"
                type="password"
                autocomplete="new-password"
                :placeholder="keyPlaceholder"
                :disabled="fieldDisabled"
              />
            </label>

            <label class="settings-field">
              <span>模型名 <em>按服务商控制台填写</em></span>
              <input
                v-model="model"
                list="mingyun-model-options"
                placeholder="例如 deepseek-chat / qwen-plus / glm-5.3"
                :disabled="fieldDisabled"
              />
              <datalist id="mingyun-model-options">
                <option v-for="name in modelOptions" :key="name" :value="name" />
              </datalist>
            </label>

            <div class="settings-endpoint">
              <span>接口地址</span>
              <input
                v-if="customBase || showBase"
                v-model="baseURL"
                placeholder="https://api.example.com/v1"
                :disabled="fieldDisabled"
              />
              <code v-else>{{ baseURL || '—' }}</code>
              <button
                v-if="!customBase"
                type="button"
                class="text-btn"
                :disabled="fieldDisabled"
                @click="showBase = !showBase"
              >
                {{ showBase ? '收起' : '改' }}
              </button>
            </div>

            <button
              type="button"
              class="settings-more"
              :disabled="fieldDisabled"
              @click="showAdvanced = !showAdvanced"
            >
              {{ showAdvanced ? '收起高级参数' : '高级参数' }}
            </button>
            <div v-if="showAdvanced" class="settings-advanced">
              <label class="settings-field settings-field--short">
                <span>温度</span>
                <input v-model.number="temperature" type="number" min="0" max="2" step="0.1" :disabled="fieldDisabled" />
              </label>
              <label class="settings-field settings-field--short">
                <span>输出上限</span>
                <input v-model.number="maxTokens" type="number" min="200" max="8192" step="100" :disabled="fieldDisabled" />
              </label>
            </div>

            <p v-if="!writable" class="settings-locked">
              服务端已锁定模型设置（MINGYUN_MODEL_SETTINGS=off），只能由环境变量配置。
            </p>
            <p v-if="localError" class="settings-error-line">{{ localError }}</p>

            <div class="settings-actions">
              <button class="settings-btn settings-btn--test" type="button" :disabled="testing || !canSubmit" @click="test">
                {{ testing ? '试火中…' : '测试连接' }}
              </button>
              <button class="settings-btn settings-btn--save" type="button" :disabled="saving || !canSubmit || !writable" @click="save">
                {{ saving ? '封存中…' : '保存并启用' }}
              </button>
            </div>

            <div v-if="testResult" class="settings-test" :class="testTone">
              <b>{{ testResult.ok ? '接通' : '未通' }}</b>
              <span>{{ testResult.message }} · {{ testResult.latency_ms }}ms</span>
              <p v-if="testResult.preview">“{{ testResult.preview }}”</p>
            </div>
            <p v-if="savedNotice" class="settings-saved">{{ savedNotice }}</p>
          </div>
        </section>

        <aside class="settings-note">
          <span>三 · 守则</span>
          <p>Key 由服务端持有（0600 文件或环境变量），网页只显示尾号；公开部署请加 HTTPS、反代鉴权与限流。</p>
          <p>网页配置优先于 <code>MINGYUN_AI_*</code> 环境变量；撤去后回落到环境变量。</p>
          <div class="settings-danger" v-if="settings.source === 'file'">
            <button type="button" :disabled="clearing || !writable" @click="removeConfig">
              {{ confirmingClear ? '再点一次确认撤去' : '撤去网页配置' }}
            </button>
            <button v-if="confirmingClear" type="button" class="text-btn" @click="confirmingClear = false">取消</button>
          </div>
        </aside>
      </template>

      <p v-if="!settings && !loading && !error" class="settings-loading">模型设置不可用。</p>
    </div>
  </section>
</template>
