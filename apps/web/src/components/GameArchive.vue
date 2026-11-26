<script setup lang="ts">
import { computed } from 'vue'
import type { ArchiveEntry } from '../types'

const props = defineProps<{ entries: ArchiveEntry[] }>()
defineEmits<{
  (e: 'open', entry: ArchiveEntry): void
  (e: 'cast'): void
  (e: 'home'): void
}>()

const aiCount = computed(() => props.entries.filter((entry) => entry.mode === 'ai').length)

function formatSaved(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString('zh-CN', { hour12: false })
}

function modeLabel(entry: ArchiveEntry): string {
  if (entry.mode === 'ai') return entry.model || 'AI 解读'
  if (entry.mode === 'offline') return '离线摘要'
  return '未解读'
}
</script>

<template>
  <section class="archive-view">
    <span class="view-watermark" aria-hidden="true">藏</span>
    <header class="scene-bar scene-bar--page">
      <button class="scene-back" @click="$emit('home')">← 山门</button>
      <div class="scene-heading">
        <span>卷二</span>
        <h1>藏卦例</h1>
      </div>
      <div class="archive-count"><b>{{ entries.length }}</b><span>卦</span></div>
    </header>

    <div v-if="entries.length" class="archive-body">
      <p class="archive-intro">
        只存本机，不上传服务端。点一卦，可以翻回原盘与当时的四段解读。
      </p>
      <div class="archive-list">
        <button
          v-for="(entry, i) in entries"
          :key="entry.chart_id"
          class="archive-row"
          @click="$emit('open', entry)"
        >
          <span class="archive-no">{{ String(i + 1).padStart(2, '0') }}</span>
          <div class="archive-main">
            <h2>{{ entry.gua_name }}</h2>
            <p>{{ entry.question || '未记所问之事' }}</p>
          </div>
          <div class="archive-meta">
            <span :class="{ 'is-ai': entry.mode === 'ai' }">{{ modeLabel(entry) }}</span>
            <time>{{ formatSaved(entry.saved_at) }}</time>
          </div>
          <i class="archive-arrow">↗</i>
        </button>
      </div>

      <aside class="archive-aside">
        <div class="archive-tally">
          <span>已藏</span>
          <b>{{ entries.length }}</b>
          <em>卦</em>
          <p>其中 {{ aiCount }} 卦有 AI 解读，其余为离线摘要。</p>
        </div>

        <div v-if="entries[0]" class="archive-latest">
          <span class="side-kicker">最近一卦</span>
          <h3>{{ entries[0].gua_name }}</h3>
          <p>{{ entries[0].question || '未记所问之事' }}</p>
          <dl>
            <div>
              <dt>收卦</dt>
              <dd>{{ formatSaved(entries[0].saved_at) }}</dd>
            </div>
            <div>
              <dt>解读</dt>
              <dd>{{ modeLabel(entries[0]) }}</dd>
            </div>
          </dl>
          <button class="archive-latest-btn" @click="$emit('open', entries[0])">
            翻回原盘 <i>→</i>
          </button>
        </div>

        <div class="archive-note">
          <span class="side-kicker">存档说明</span>
          <p>卦例保存在浏览器 localStorage 中，清空站点数据会一并移除；服务端不保存你的问事与解读。</p>
        </div>
      </aside>
    </div>

    <div v-else class="archive-empty">
      <div class="archive-empty-seal">空</div>
      <div>
        <h2>尚无卦例</h2>
        <p>先问一事，摇六次铜钱。卦成之后，它会自动收进这里。</p>
        <button class="seal-btn seal-btn--small" @click="$emit('cast')">去起一卦 <i>→</i></button>
      </div>
    </div>
  </section>
</template>
