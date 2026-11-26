<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ArchiveEntry, ArchiveStatus } from '../types'
import GameArchiveDetail from './GameArchiveDetail.vue'

const props = defineProps<{ entries: ArchiveEntry[] }>()
const emit = defineEmits<{
  (e: 'open', entry: ArchiveEntry): void
  (e: 'cast'): void
  (e: 'home'): void
  (e: 'delete', entry: ArchiveEntry): void
  (e: 'clear'): void
  (e: 'export'): void
}>()

type FilterKey = 'all' | 'ai' | 'offline' | 'pending' | 'error'

const query = ref('')
const filter = ref<FilterKey>('all')
const selectedId = ref('')
const pendingDeleteId = ref('')
const confirmClear = ref(false)

const filters: { key: FilterKey; label: string }[] = [
  { key: 'all', label: '全部' },
  { key: 'ai', label: 'AI 解读' },
  { key: 'offline', label: '离线摘要' },
  { key: 'pending', label: '待解读' },
  { key: 'error', label: '失败' }
]

function statusOf(entry: ArchiveEntry): ArchiveStatus {
  if (entry.status) return entry.status
  if (entry.reading) return 'done'
  if (entry.reading_error) return 'error'
  return 'pending'
}

function statusLabel(entry: ArchiveEntry): string {
  const status = statusOf(entry)
  if (status === 'error') return '解读失败'
  if (status === 'reading') return '解读中'
  if (status === 'pending') return '待解读'
  if (entry.mode === 'offline') return '离线摘要'
  return entry.model || 'AI 解读'
}

function statusClass(entry: ArchiveEntry): string {
  const status = statusOf(entry)
  if (status === 'error') return 'is-error'
  if (status === 'pending' || status === 'reading') return 'is-pending'
  if (entry.mode === 'offline') return 'is-offline'
  return 'is-ai'
}

function formatTime(iso: string): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  })
}

function formatFullTime(iso: string): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', { hour12: false })
}

const counts = computed<Record<FilterKey, number>>(() => {
  const result: Record<FilterKey, number> = { all: props.entries.length, ai: 0, offline: 0, pending: 0, error: 0 }
  for (const entry of props.entries) {
    const status = statusOf(entry)
    if (status === 'error') {
      result.error += 1
    } else if (status === 'pending' || status === 'reading') {
      result.pending += 1
    } else if (entry.mode === 'offline') {
      result.offline += 1
    } else {
      result.ai += 1
    }
  }
  return result
})

const filteredEntries = computed(() => {
  const keyword = query.value.trim().toLowerCase()
  return props.entries.filter((entry) => {
    const status = statusOf(entry)
    if (filter.value === 'ai' && !(status === 'done' && entry.mode === 'ai')) return false
    if (filter.value === 'offline' && !(status === 'done' && entry.mode === 'offline')) return false
    if (filter.value === 'pending' && !(status === 'pending' || status === 'reading')) return false
    if (filter.value === 'error' && status !== 'error') return false
    if (!keyword) return true
    const haystack = [
      entry.question,
      entry.gua_name,
      entry.changed_gua,
      entry.topic_label,
      entry.model,
      entry.reading,
      ...(entry.followups ?? []).map((turn) => turn.content)
    ]
      .join(' ')
      .toLowerCase()
    return haystack.includes(keyword)
  })
})

const selectedEntry = computed(
  () => filteredEntries.value.find((entry) => entry.chart_id === selectedId.value) || null
)

function selectEntry(entry: ArchiveEntry) {
  confirmClear.value = false
  if (selectedId.value === entry.chart_id) {
    selectedId.value = ''
    pendingDeleteId.value = ''
    return
  }
  selectedId.value = entry.chart_id
  pendingDeleteId.value = ''
}

function requestDelete(entry: ArchiveEntry) {
  if (pendingDeleteId.value === entry.chart_id) {
    emit('delete', entry)
    pendingDeleteId.value = ''
    selectedId.value = ''
    return
  }
  pendingDeleteId.value = entry.chart_id
  confirmClear.value = false
}

function requestClear() {
  if (confirmClear.value) {
    emit('clear')
    confirmClear.value = false
    return
  }
  confirmClear.value = true
  pendingDeleteId.value = ''
}
</script>

<template>
  <section class="archive-view">
    <span class="view-watermark" aria-hidden="true">藏</span>
    <header class="scene-bar scene-bar--page">
      <button class="scene-back" @click="$emit('home')">← 山门</button>
      <div class="scene-heading">
        <span>卷二 · 藏</span>
        <h1>卦历</h1>
      </div>
      <div class="archive-count"><b>{{ entries.length }}</b><span>则</span></div>
    </header>

    <div v-if="entries.length" class="archive-body">
      <p class="archive-intro">
        每次起卦自动留档，完整盘面、用神与当时解读都收在这里；点一则展开详情。只存本机，不上传服务端。
      </p>

      <div class="archive-main-column">
        <div class="archive-toolbar">
          <label class="archive-search">
            <span>寻</span>
            <input v-model="query" type="search" placeholder="按问题、卦名、模型或断语搜索" />
            <button v-if="query" type="button" @click="query = ''">清</button>
          </label>
          <div class="archive-filters" aria-label="按解读状态筛选">
            <button
              v-for="item in filters"
              :key="item.key"
              type="button"
              class="archive-filter"
              :class="{ 'is-on': filter === item.key }"
              @click="filter = item.key; confirmClear = false"
            >
              {{ item.label }} <b>{{ counts[item.key] }}</b>
            </button>
          </div>
        </div>

        <div v-if="filteredEntries.length" class="archive-list">
          <div v-for="(entry, i) in filteredEntries" :key="entry.chart_id" class="archive-entry">
            <button
              class="archive-row"
              :class="{ 'is-selected': selectedEntry?.chart_id === entry.chart_id }"
              @click="selectEntry(entry)"
            >
              <span class="archive-no">{{ String(i + 1).padStart(2, '0') }}</span>
              <div class="archive-main">
                <h2>
                  {{ entry.gua_name }}
                  <i v-if="entry.changed_gua">→ {{ entry.changed_gua }}</i>
                </h2>
                <p>{{ entry.question || '未记所问之事' }}</p>
                <em v-if="entry.topic_label">{{ entry.topic_label }}</em>
                <em
                  v-if="entry.followups?.some((turn) => turn.role === 'user')"
                  class="archive-followup-count"
                >
                  含 {{ entry.followups.filter((turn) => turn.role === 'user').length }} 轮追问
                </em>
              </div>
              <div class="archive-meta">
                <span :class="statusClass(entry)">{{ statusLabel(entry) }}</span>
                <time>{{ formatTime(entry.cast_at || entry.chart.created_at || entry.saved_at) }}</time>
              </div>
              <i class="archive-arrow">↗</i>
            </button>
            <GameArchiveDetail
              v-if="selectedEntry?.chart_id === entry.chart_id"
              :entry="selectedEntry"
              :confirming-delete="pendingDeleteId === entry.chart_id"
              @open="$emit('open', $event)"
              @delete="requestDelete"
            />
          </div>
        </div>

        <div v-else class="archive-no-match">
          <span>无</span>
          <p>没有找到匹配的记录。</p>
          <button type="button" class="text-btn" @click="query = ''; filter = 'all'">清掉筛选</button>
        </div>
      </div>

      <aside class="archive-aside">
        <div class="archive-tally">
          <span>共录</span>
          <b>{{ entries.length }}</b>
          <em>则</em>
          <p>{{ counts.ai }} 则有 AI 解读 · {{ counts.offline }} 则离线摘要 · {{ counts.pending }} 则待解读</p>
        </div>

        <div v-if="entries[0]" class="archive-latest">
          <span class="side-kicker">最近一卦</span>
          <h3>{{ entries[0].gua_name }}</h3>
          <p>{{ entries[0].question || '未记所问之事' }}</p>
          <dl>
            <div>
              <dt>起卦</dt>
              <dd>{{ formatFullTime(entries[0].cast_at || entries[0].chart.created_at) }}</dd>
            </div>
            <div>
              <dt>解读</dt>
              <dd>{{ statusLabel(entries[0]) }}</dd>
            </div>
          </dl>
          <button class="archive-latest-btn" @click="$emit('open', entries[0])">
            翻回原盘 <i>→</i>
          </button>
        </div>

        <div class="archive-actions">
          <button type="button" class="archive-tool" @click="$emit('export')">导出 JSON</button>
          <button
            type="button"
            class="archive-tool is-danger"
            :class="{ 'is-confirming': confirmClear }"
            @click="requestClear"
          >
            {{ confirmClear ? '再点一次清空' : '清空记录' }}
          </button>
        </div>

        <div class="archive-note">
          <span class="side-kicker">存档说明</span>
          <p>
            卦历保存在浏览器 localStorage，最多保留 100 则；清空站点数据会一并移除。服务端不保存你的问事与解读。
          </p>
        </div>
      </aside>
    </div>

    <div v-else class="archive-empty">
      <div class="archive-empty-seal">空</div>
      <div>
        <h2>尚无卦历</h2>
        <p>先问一事，摇六次铜钱。卦成之后，盘面与解读会自动收进这里。</p>
        <button class="seal-btn seal-btn--small" @click="$emit('cast')">去起一卦 <i>→</i></button>
      </div>
    </div>
  </section>
</template>
