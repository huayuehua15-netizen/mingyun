<script setup lang="ts">
import { computed } from 'vue'
import type { ArchiveEntry, ArchiveStatus, LinePayload } from '../types'
import { isLiuYaoPayload } from '../types'

const props = defineProps<{
  entry: ArchiveEntry
  confirmingDelete: boolean
}>()

defineEmits<{
  (e: 'open', entry: ArchiveEntry): void
  (e: 'delete', entry: ArchiveEntry): void
}>()

function statusOf(item: ArchiveEntry): ArchiveStatus {
  if (item.status) return item.status
  if (item.reading) return 'done'
  if (item.reading_error) return 'error'
  return 'pending'
}

function statusLabel(item: ArchiveEntry): string {
  const status = statusOf(item)
  if (status === 'error') return '解读失败'
  if (status === 'reading') return '解读中'
  if (status === 'pending') return '待解读'
  if (item.mode === 'offline') return '离线摘要'
  return item.model || 'AI 解读'
}

function statusClass(item: ArchiveEntry): string {
  const status = statusOf(item)
  if (status === 'error') return 'is-error'
  if (status === 'pending' || status === 'reading') return 'is-pending'
  if (item.mode === 'offline') return 'is-offline'
  return 'is-ai'
}

function formatFullTime(iso: string): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('zh-CN', { hour12: false })
}

function yaoLabel(position: number): string {
  return ['初', '二', '三', '四', '五', '上'][position - 1] + '爻'
}

function lineMarks(line: LinePayload): string {
  const marks = [...(line.marks || [])]
  if (line.shi) marks.unshift('世')
  if (line.ying) marks.unshift('应')
  return marks.filter(Boolean).join(' · ') || '—'
}

function lineChange(line: LinePayload): string {
  if (!line.change) return '—'
  return `${line.change.gan}${line.change.zhi}${line.change.wu_xing} · ${line.change.liu_qin}`
}

function useShenText(item: ArchiveEntry): string {
  const payload = item.chart.payload
  if (!isLiuYaoPayload(payload)) return '非六爻盘'
  const plan = payload.use_shen
  if (!plan) return '未定用神'
  const primary = plan.primary?.length ? plan.primary.join('、') : '—'
  return `${plan.topic_label} · ${plan.rule_id} · 主用神 ${primary}`
}

const liuyao = computed(() => {
  const payload = props.entry.chart.payload
  return isLiuYaoPayload(payload) ? payload : null
})

const title = computed(() => {
  if (props.entry.title) return props.entry.title
  if (props.entry.gua_name) return props.entry.gua_name
  const payload = props.entry.chart.payload
  if (isLiuYaoPayload(payload)) return payload.gua_name
  if ('ke_ti' in payload) return payload.ke_ti
  if ('final_palace' in payload) return `小六壬 · ${payload.final_palace.name}`
  if ('ben_gua' in payload) return `梅花 · ${payload.ben_gua.name}`
  if ('zhi_fu' in payload) return `奇门 · ${payload.dun}${payload.ju}局`
  return `${payload.ming_gong}宫 · ${payload.wu_xing_ju}`
})

const subtitle = computed(() => {
  if (props.entry.subtitle) return props.entry.subtitle
  if (props.entry.changed_gua) return '之 ' + props.entry.changed_gua
  if (props.entry.method === 'daliuren') return '大六壬课'
  if (props.entry.method === 'ziwei') return '紫微命盘'
  if (props.entry.method === 'xiaoliuren') return '小六壬课'
  if (props.entry.method === 'meihua') return '梅花易数卦'
  if (props.entry.method === 'qimen') return '奇门遁甲局'
  return '六爻卦'
})

const methodFacts = computed(() =>
  props.entry.chart.facts.filter((f) => f.weight === 'core' || f.weight === 'normal').slice(0, 16)
)

interface ReadingSections {
  overview: string
  verdict: string
  timing: string
  advice: string
}

function parseReading(text: string): ReadingSections {
  const out: ReadingSections = { overview: '', verdict: '', timing: '', advice: '' }
  if (!text.trim()) return out

  const map: Record<string, keyof ReadingSections> = {
    盘面要点: 'overview',
    断语: 'verdict',
    应期与走向: 'timing',
    应期: 'timing',
    建议: 'advice'
  }
  const raw: Record<keyof ReadingSections, string[]> = { overview: [], verdict: [], timing: [], advice: [] }
  let current: keyof ReadingSections | '' = ''
  let found = false

  for (const line of text.replace(/\r\n/g, '\n').split('\n')) {
    const heading = line.match(
      /^\s{0,3}#{1,3}\s*(?:\*\*)?\s*(盘面要点|断语|应期与走向|应期|建议)\s*(?:\*\*)?\s*[:：]?\s*$/
    )
    if (heading) {
      current = map[heading[1]]
      found = true
      continue
    }
    if (current) raw[current].push(line)
    else raw.overview.push(line)
  }

  const clean = (lines: string[]) =>
    lines
      .join('\n')
      .replace(/\*\*/g, '')
      .replace(/^\s*#+\s*/gm, '')
      .trim()

  if (!found) {
    out.verdict = clean(raw.overview)
    return out
  }
  out.overview = clean(raw.overview)
  out.verdict = clean(raw.verdict)
  out.timing = clean(raw.timing)
  out.advice = clean(raw.advice)
  return out
}

const lines = computed(() => {
  const payload = liuyao.value
  if (!payload) return []
  return [...payload.lines].sort((a, b) => b.position - a.position)
})
const reading = computed(() => parseReading(props.entry.reading))
const followupTurns = computed(() => props.entry.followups ?? [])
const followupRounds = computed(
  () =>
    followupTurns.value.filter(
      (turn) => turn.role === 'assistant' && turn.status === 'done'
    ).length
)
</script>

<template>
  <section class="archive-detail" aria-label="历史详情">
    <header class="archive-detail-head">
      <div class="archive-detail-copy">
        <div class="archive-detail-kicker">
          <span :class="statusClass(entry)">{{ statusLabel(entry) }}</span>
          <em>{{ entry.topic_label || subtitle }}</em>
        </div>
        <h2>
          {{ title }}
          <i>{{ subtitle }}</i>
        </h2>
        <p>{{ entry.question || '未记所问之事' }}</p>
      </div>
      <div class="archive-detail-actions">
        <button class="archive-open" type="button" @click="$emit('open', entry)">
          翻回原盘 <i>→</i>
        </button>
        <button
          class="archive-delete"
          :class="{ 'is-confirming': confirmingDelete }"
          type="button"
          @click="$emit('delete', entry)"
        >
          {{ confirmingDelete ? '再点一次确认' : '删除' }}
        </button>
      </div>
    </header>

    <dl v-if="liuyao" class="archive-detail-facts">
      <div>
        <dt>卦体</dt>
        <dd>
          {{ liuyao.upper }}上{{ liuyao.lower }}下 ·
          {{ liuyao.palace }}
        </dd>
      </div>
      <div>
        <dt>世应</dt>
        <dd>世 {{ yaoLabel(liuyao.shi_yao) }} · 应 {{ yaoLabel(liuyao.ying_yao) }}</dd>
      </div>
      <div>
        <dt>月建 / 日辰</dt>
        <dd>{{ liuyao.month_gz }} / {{ liuyao.day_gz }}</dd>
      </div>
      <div>
        <dt>旬空</dt>
        <dd>{{ liuyao.xun_kong.join('、') }}</dd>
      </div>
      <div class="archive-fact-wide">
        <dt>用神</dt>
        <dd>{{ useShenText(entry) }}</dd>
      </div>
      <div>
        <dt>六亲</dt>
        <dd>{{ liuyao.lines.map((line) => line.liu_qin).join(' · ') }}</dd>
      </div>
      <div>
        <dt>起卦时间</dt>
        <dd>{{ formatFullTime(entry.cast_at || entry.chart.created_at) }}</dd>
      </div>
      <div>
        <dt>解读时间</dt>
        <dd>{{ entry.read_at ? formatFullTime(entry.read_at) : '尚未解读' }}</dd>
      </div>
    </dl>

    <dl v-else class="archive-detail-facts">
      <div v-for="f in methodFacts" :key="f.key" :class="{ 'archive-fact-wide': f.weight === 'core' }">
        <dt>{{ f.label }}</dt>
        <dd>{{ f.value }}</dd>
      </div>
      <div>
        <dt>起盘时间</dt>
        <dd>{{ formatFullTime(entry.cast_at || entry.chart.created_at) }}</dd>
      </div>
      <div>
        <dt>解读时间</dt>
        <dd>{{ entry.read_at ? formatFullTime(entry.read_at) : '尚未解读' }}</dd>
      </div>
    </dl>

    <div v-if="liuyao" class="archive-lines">
      <header class="archive-lines-head">
        <h3>六爻盘</h3>
        <em>自上而下 · 变爻单列</em>
      </header>
      <div class="archive-line-table">
        <div class="archive-line-head">
          <span>爻位</span><span>六亲</span><span>纳甲</span><span>阴阳</span><span>六神</span
          ><span>标记</span><span>变爻</span>
        </div>
        <div v-for="line in lines" :key="line.position" class="archive-line-row">
          <span class="archive-line-pos">{{ yaoLabel(line.position) }}</span>
          <b>{{ line.liu_qin }}</b>
          <span>{{ line.gan }}{{ line.zhi }} · {{ line.wu_xing }}</span>
          <span>{{ line.yin_yang }}{{ line.moving ? '动' : '静' }}</span>
          <span>{{ line.liu_shen }}</span>
          <em>{{ lineMarks(line) }}</em>
          <span class="archive-line-change">{{ lineChange(line) }}</span>
        </div>
      </div>
    </div>

    <div class="archive-detail-reading">
      <header class="archive-detail-reading-head">
        <h3>当时解读</h3>
        <em v-if="entry.mode === 'ai'">{{ entry.model || 'AI' }} · 已校验</em>
        <em v-else-if="entry.mode === 'offline'">离线摘要 · 只列事实</em>
        <em v-else>未生成解读</em>
      </header>

      <div v-if="entry.reading" class="archive-reading-sections">
        <section v-if="reading.overview">
          <span>盘面要点</span>
          <p>{{ reading.overview }}</p>
        </section>
        <section v-if="reading.verdict">
          <span>断语</span>
          <p>{{ reading.verdict }}</p>
        </section>
        <section v-if="reading.timing">
          <span>应期与走向</span>
          <p>{{ reading.timing }}</p>
        </section>
        <section v-if="reading.advice">
          <span>建议</span>
          <p>{{ reading.advice }}</p>
        </section>
      </div>
      <div v-else-if="entry.status === 'error' || entry.reading_error" class="archive-reading-empty">
        <p>这条记录在解读阶段失败：{{ entry.reading_error || '模型未返回结果' }}。</p>
        <p>盘面与起卦信息已保留，可翻回原盘后重试。</p>
      </div>
      <div v-else class="archive-reading-empty">
        <p>起卦时已自动留档，但当时没有生成解读。翻回原盘后可以补一次解读。</p>
      </div>
    </div>

    <div v-if="followupTurns.length" class="archive-detail-followups">
      <header class="archive-detail-reading-head">
        <h3>同盘追问</h3>
        <em>{{ followupRounds }} 轮 · 同一张盘</em>
      </header>
      <div class="archive-followup-list">
        <article
          v-for="turn in followupTurns"
          :key="turn.id"
          class="archive-followup-turn"
          :class="[`is-${turn.role}`, { 'is-error': turn.status === 'error' }]"
        >
          <header>
            <strong>{{ turn.role === 'user' ? '你问' : turn.model || '命云' }}</strong>
            <time>{{ formatFullTime(turn.created_at) }}</time>
          </header>
          <p v-if="turn.content">{{ turn.content }}</p>
          <em v-if="turn.error" class="archive-followup-error">{{ turn.error }}</em>
        </article>
      </div>
    </div>
  </section>
</template>
