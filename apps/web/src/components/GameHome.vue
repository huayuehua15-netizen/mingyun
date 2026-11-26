<script setup lang="ts">
import { computed } from 'vue'
import type { ArchiveEntry, CalendarContext } from '../types'
import InkScene from './InkScene.vue'

const props = defineProps<{
  calendar: CalendarContext | null
  archiveCount: number
  hasChart: boolean
  lastEntry: ArchiveEntry | null
}>()

const emit = defineEmits<{
  (e: 'start'): void
  (e: 'archive'): void
  (e: 'codex'): void
}>()

const dateLine = computed(() => {
  const c = props.calendar
  if (!c) return '历法载入中'
  return `${c.year_gz}年 ${c.month_gz}月 ${c.day_gz}日`
})

const termLine = computed(() => {
  const c = props.calendar
  if (!c) return '—'
  return `${c.prev_term.name}后 · ${c.month_general.zhi}将${c.month_general.name}`
})

const methods = [
  {
    key: 'liuyao',
    glyph: '六',
    name: '六爻',
    desc: '铜钱 · 纳甲 · 动变',
    state: '已开放',
    open: true
  },
  {
    key: 'daliuren',
    glyph: '壬',
    name: '大六壬',
    desc: '课体 · 三传 · 十二天将',
    state: '筹备',
    open: false
  },
  {
    key: 'ziwei',
    glyph: '紫',
    name: '紫微斗数',
    desc: '星曜 · 十二宫 · 四化',
    state: '筹备',
    open: false
  }
]

function enterMethod(open: boolean) {
  if (open) emit('start')
}
</script>

<template>
  <section class="home-view">
    <div class="home-scene">
      <InkScene />
      <div class="home-copy">
        <div class="chapter-line">
          <span>序</span>
          <i></i>
          <b>山门</b>
        </div>
        <h1><span>问一事</span><span>起一卦</span></h1>
        <p class="home-lede">
          铜钱六摇，装卦断事。<br />
          盘面由确定性引擎计算，解读交给你自配的大模型。
        </p>
        <div class="home-actions">
          <button class="seal-btn" @click="$emit('start')">
            {{ hasChart ? '继续解卦' : '入山问卦' }}
            <i>→</i>
          </button>
          <button class="text-btn" @click="$emit('codex')">先读三术</button>
        </div>
        <div class="home-stamp">
          <span>心诚</span>
          <span>则灵</span>
          <i>命云</i>
        </div>
      </div>

      <aside class="home-aside">
        <div class="quest-board">
          <header>
            <span class="quest-kicker">今日</span>
            <b>{{ dateLine }}</b>
          </header>
          <dl class="quest-grid">
            <div><dt>节气</dt><dd>{{ termLine }}</dd></div>
            <div><dt>旬空</dt><dd>{{ calendar?.xun_kong.join('、') || '—' }}</dd></div>
            <div><dt>时辰</dt><dd>{{ calendar?.hour_gz || '—' }}时</dd></div>
            <div><dt>真太阳时</dt><dd>{{ calendar?.true_solar_used ? '已校正' : '未校正' }}</dd></div>
          </dl>
          <button class="quest-link" @click="$emit('archive')">
            <span>翻看卦例</span>
            <b>{{ archiveCount }}</b>
            <i>→</i>
          </button>
        </div>

        <div class="side-note">
          <span class="side-kicker">支线</span>
          <p v-if="lastEntry">
            最近一卦：<b>{{ lastEntry.gua_name }}</b><br />
            {{ lastEntry.question || '未记所问' }}
          </p>
          <p v-else>每成一卦，会自动收入「卦例」，只存在这台机器上。</p>
        </div>
      </aside>
    </div>

    <section class="home-methods">
      <header class="section-head">
        <span>三术</span>
        <i></i>
        <em>目前只开放六爻</em>
      </header>
      <div class="method-grid">
        <button
          v-for="m in methods"
          :key="m.key"
          class="method-card"
          :class="{ 'is-open': m.open, 'is-locked': !m.open }"
          :disabled="!m.open"
          @click="enterMethod(m.open)"
        >
          <span class="method-glyph">{{ m.glyph }}</span>
          <span class="method-text">
            <b>{{ m.name }}</b>
            <em>{{ m.desc }}</em>
          </span>
          <span class="method-state">{{ m.state }}</span>
        </button>
      </div>
    </section>
  </section>
</template>
