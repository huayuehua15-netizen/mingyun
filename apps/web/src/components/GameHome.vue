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
    volume: '卷一',
    roman: '壹',
    glyph: '六',
    name: '六爻',
    desc: '铜钱 · 纳甲 · 动变',
    state: '已开卷',
    open: true,
    steps: ['问事', '摇卦', '解卦'],
    action: '入卷'
  },
  {
    key: 'daliuren',
    volume: '卷二',
    roman: '贰',
    glyph: '壬',
    name: '大六壬',
    desc: '课体 · 三传 · 十二天将',
    state: '待启',
    open: false,
    steps: ['起课', '传课', '断课'],
    action: '未启'
  },
  {
    key: 'ziwei',
    volume: '卷三',
    roman: '叁',
    glyph: '紫',
    name: '紫微斗数',
    desc: '星曜 · 十二宫 · 四化',
    state: '待启',
    open: false,
    steps: ['安星', '布宫', '读盘'],
    action: '未启'
  }
]

const openCount = methods.filter((method) => method.open).length

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

    <section class="home-route">
      <header class="section-head">
        <span>行卷</span>
        <i></i>
        <em>一卷六爻，二卷大六壬，三卷紫微斗数</em>
        <b class="route-progress">{{ openCount }}<span>/ {{ methods.length }}</span></b>
      </header>

      <div class="route-track">
        <div class="route-line" aria-hidden="true"></div>
        <button
          v-for="m in methods"
          :key="m.key"
          class="route-stop"
          :class="{ 'is-open': m.open, 'is-locked': !m.open }"
          :disabled="!m.open"
          @click="enterMethod(m.open)"
        >
          <span class="route-node"><i>{{ m.roman }}</i></span>
          <div class="route-card">
            <header class="route-card-head">
              <span class="route-volume">{{ m.volume }}</span>
              <em>{{ m.state }}</em>
            </header>
            <div class="route-title">
              <span class="route-glyph">{{ m.glyph }}</span>
              <div>
                <h2>{{ m.name }}</h2>
                <p>{{ m.desc }}</p>
              </div>
            </div>
            <ul class="route-steps">
              <li v-for="(step, i) in m.steps" :key="step">
                <span>{{ String(i + 1).padStart(2, '0') }}</span>{{ step }}
              </li>
            </ul>
            <footer class="route-card-foot">
              <span>{{ m.action }}</span>
              <i>{{ m.open ? '→' : '·' }}</i>
            </footer>
          </div>
        </button>
      </div>

      <div class="route-foot">
        <button class="text-btn" @click="$emit('archive')">翻看卦例 {{ archiveCount }}</button>
        <button class="text-btn" @click="$emit('codex')">展开三术</button>
      </div>
    </section>
  </section>
</template>
