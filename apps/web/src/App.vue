<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { castChart, castLiuYao, fetchCalendar, streamFollowup, streamInterpret } from './api'
import type { CastMethod, CastRequest, InterpretEvent } from './api'
import type { ArchiveEntry, CalendarContext, ChartPayload, ChartView, FollowupMessage, ViewName } from './types'
import {
  isDaLiuRenPayload,
  isLiuYaoPayload,
  isMeiHuaPayload,
  isQiMenPayload,
  isXiaoLiuRenPayload,
  isZiWeiPayload
} from './types'
import LiuYaoBoard from './components/LiuYaoBoard.vue'
import DaLiuRenBoard from './components/DaLiuRenBoard.vue'
import ZiWeiBoard from './components/ZiWeiBoard.vue'
import XiaoLiuRenBoard from './components/XiaoLiuRenBoard.vue'
import MeiHuaBoard from './components/MeiHuaBoard.vue'
import QiMenBoard from './components/QiMenBoard.vue'
import GameHome from './components/GameHome.vue'
import GameArchive from './components/GameArchive.vue'
import GameCodex from './components/GameCodex.vue'
import GameNav from './components/GameNav.vue'

const VALID_VIEWS: ViewName[] = ['home', 'cast', 'archive', 'codex']

function resolveInitialView(): ViewName {
  const fromHash = window.location.hash.replace(/^#\/?/, '') as ViewName
  if (VALID_VIEWS.includes(fromHash)) return fromHash
  const params = new URLSearchParams(window.location.search)
  if (params.get('cast') === '1') return 'cast'
  return 'home'
}

const view = ref<ViewName>(resolveInitialView())
const method = ref<CastMethod>('liuyao')
const chartUnit = computed(
  () =>
    ({
      liuyao: '卦',
      daliuren: '课',
      ziwei: '盘',
      xiaoliuren: '课',
      meihua: '卦',
      qimen: '局'
    })[method.value]
)
const methodLabel = computed(
  () =>
    ({
      liuyao: '六爻',
      daliuren: '大六壬',
      ziwei: '紫微斗数',
      xiaoliuren: '小六壬',
      meihua: '梅花易数',
      qimen: '奇门遁甲'
    })[method.value]
)
const methodMeta = computed(
  () =>
    ({
      liuyao: { volume: '卷一', glyph: '卦', steps: ['问事', '摇卦', '解卦'] },
      daliuren: { volume: '卷二', glyph: '课', steps: ['定时', '成课', '断课'] },
      ziwei: { volume: '卷三', glyph: '盘', steps: ['生辰', '安星', '读盘'] },
      xiaoliuren: { volume: '卷四', glyph: '课', steps: ['问事', '起课', '断课'] },
      meihua: { volume: '卷五', glyph: '卦', steps: ['起卦', '成卦', '断卦'] },
      qimen: { volume: '卷六', glyph: '局', steps: ['定时', '起局', '断局'] }
    })[method.value]
)

watch(
  [view, methodLabel],
  ([current, label]) => {
    const viewName =
      current === 'home'
        ? '山门'
        : current === 'cast'
          ? label
          : current === 'archive'
            ? '卦历'
            : '六术'
    document.title = `命云 MingYun · ${viewName}`
  },
  { immediate: true }
)

function stepDone(index: number): boolean {
  if (index === 2) return !!chart.value
  if (method.value === 'liuyao') return index === 0 ? questionReady.value : tosses.value.length > 0
  if (method.value === 'daliuren') return index === 0 ? !!daliurenAt.value && questionReady.value : !!chart.value
  if (method.value === 'xiaoliuren' || method.value === 'qimen') {
    return index === 0 ? !!quickAt.value && questionReady.value : !!chart.value
  }
  if (method.value === 'meihua') {
    return index === 0 ? questionReady.value && (meihuaMode.value === 'time' || meihuaNumbersReady.value) : !!chart.value
  }
  return index === 0 ? birthReady.value : !!chart.value
}
const archive = ref<ArchiveEntry[]>([])
const ARCHIVE_KEY = 'mingyun.archive.v1'
const ARCHIVE_LIMIT = 100

const question = ref('')
const mode = ref<'coin' | 'manual'>('coin')
const manualLines = ref<string[]>(['少阳', '少阳', '少阳', '少阳', '少阳', '少阳'])
const lineOptions = ['少阳', '少阴', '老阳', '老阴']

const methodTabs: { key: CastMethod; name: string; note: string }[] = [
  { key: 'liuyao', name: '六爻', note: '问事' },
  { key: 'daliuren', name: '大六壬', note: '起课' },
  { key: 'ziwei', name: '紫微', note: '看命' },
  { key: 'xiaoliuren', name: '小六壬', note: '近事' },
  { key: 'meihua', name: '梅花', note: '起卦' },
  { key: 'qimen', name: '奇门', note: '近期' }
]

const daliurenAt = ref('')
const quickAt = ref('')
const meihuaMode = ref<'time' | 'numbers'>('time')
const meihuaA = ref('3')
const meihuaB = ref('4')
const meihuaNumbersReady = computed(() => {
  const a = Number(meihuaA.value)
  const b = Number(meihuaB.value)
  return Number.isInteger(a) && Number.isInteger(b) && a > 0 && b > 0 && a <= 999999 && b <= 999999
})
const birth = ref({
  gender: 'male' as 'male' | 'female',
  calendar_type: 'solar' as 'solar' | 'lunar',
  birth_date: '1990-05-15',
  birth_time: '10:30',
  is_leap_month: false,
  leap_month_policy: 'as_this_month' as 'as_this_month' | 'as_next_month'
})

const chart = ref<ChartView | null>(null)
const calendar = ref<CalendarContext | null>(null)
const error = ref('')

const liuyaoPayload = computed(() => {
  const payload = chart.value?.payload
  return payload && isLiuYaoPayload(payload) ? payload : null
})
const daLiuRenPayload = computed(() => {
  const payload = chart.value?.payload
  return payload && isDaLiuRenPayload(payload) ? payload : null
})
const ziWeiPayload = computed(() => {
  const payload = chart.value?.payload
  return payload && isZiWeiPayload(payload) ? payload : null
})
const xiaoLiuRenPayload = computed(() => {
  const payload = chart.value?.payload
  return payload && isXiaoLiuRenPayload(payload) ? payload : null
})
const meiHuaPayload = computed(() => {
  const payload = chart.value?.payload
  return payload && isMeiHuaPayload(payload) ? payload : null
})
const qiMenPayload = computed(() => {
  const payload = chart.value?.payload
  return payload && isQiMenPayload(payload) ? payload : null
})

const castTitle = computed(() => (chart.value ? '盘已定' : `案上起${chartUnit.value}`))

const emptyCopy = computed(() => {
  switch (method.value) {
    case 'daliuren':
      return {
        seal: '未课',
        hint: '定此刻 · 起一课',
        steps: [
          ['定时', '月将加时，课随刻转'],
          ['成课', '四课三传，十二天将'],
          ['断课', '盘面即事实，AI 只负责说人话']
        ]
      }
    case 'ziwei':
      return {
        seal: '未盘',
        hint: '填生辰 · 安一盘',
        steps: [
          ['生辰', '公历农历皆可，闰月有口径'],
          ['安星', '命身十二宫，主辅四化'],
          ['读盘', '盘面即事实，AI 只负责说人话']
        ]
      }
    case 'xiaoliuren':
      return {
        seal: '未课',
        hint: '定此刻 · 起一课',
        steps: [
          ['问事', '近期一件具体事，越具体越好'],
          ['起课', '农历月日时，三宫落六神'],
          ['断课', '月看背景，日看事体，时看动向']
        ]
      }
    case 'meihua':
      return {
        seal: '未卦',
        hint: '定方式 · 起一卦',
        steps: [
          ['问事', '时间起卦或报数起卦皆可'],
          ['成卦', '本卦互卦变卦，体用生克'],
          ['断卦', '盘面即事实，AI 只负责说人话']
        ]
      }
    case 'qimen':
      return {
        seal: '未局',
        hint: '定此刻 · 起一局',
        steps: [
          ['定时', '节气定阴阳遁，符头定三元'],
          ['起局', '九宫三奇六仪，值符值使'],
          ['断局', '星门神与用神宫，先看事实']
        ]
      }
    default:
      return {
        seal: '未占',
        hint: '心诚则灵 · 先问一事',
        steps: [
          ['问事', '问题越具体，盘越好读'],
          ['摇卦', '六次铜钱，自下而上'],
          ['断卦', '盘面即事实，AI 只负责说人话']
        ]
      }
  }
})

// 一爻一摇：tosses 里每一项是三枚铜钱（1=背，0=字）。
const tosses = ref<number[][]>([])
const shaking = ref(false)
const shakingIndex = ref(0)
const spinCoins = ref<number[]>([0, 0, 0])
const landed = ref(false)
const submitting = ref(false)

const boardEl = ref<HTMLElement | null>(null)
const questionEl = ref<HTMLInputElement | null>(null)
const previewMode = ref(false)
let replayAt = ''

type ReadingStatus = 'idle' | 'loading' | 'streaming' | 'done' | 'error'
interface ReadingState {
  status: ReadingStatus
  mode: 'ai' | 'offline' | ''
  model: string
  provider: string
  text: string
  error: string
}

const reading = ref<ReadingState>({
  status: 'idle',
  mode: '',
  model: '',
  provider: '',
  text: '',
  error: ''
})
let readingAbort: AbortController | null = null
let readingSeq = 0

const MAX_FOLLOWUPS = 6
const followups = ref<FollowupMessage[]>([])
const followupInput = ref('')
const followupError = ref('')
const followupBusy = ref(false)
const followupThreadEl = ref<HTMLElement | null>(null)
let followupAbort: AbortController | null = null
let followupSeq = 0

const followupUsed = computed(() =>
  followups.value.reduce((total, turn, index) => {
    if (turn.role !== 'user') return total
    const answer = followups.value[index + 1]
    // 失败的一问一答不占轮次，允许就地重试。
    if (answer?.role === 'assistant' && answer.status === 'error') return total
    return total + 1
  }, 0)
)

const followupUnlocked = computed(
  () =>
    !!chart.value &&
    reading.value.status === 'done' &&
    reading.value.mode === 'ai' &&
    !!reading.value.text.trim()
)

const followupReady = computed(
  () => followupUnlocked.value && followupUsed.value < MAX_FOLLOWUPS
)

const followupLockedText = computed(() => {
  if (!chart.value) return `先起一${chartUnit.value}，再沿这张盘追问。`
  if (followupUsed.value >= MAX_FOLLOWUPS) {
    return `这张盘已追问 ${MAX_FOLLOWUPS} 轮，建议另起一${chartUnit.value}。`
  }
  if (readingBusy.value) return '首轮解读完成后，可以沿这张盘继续追问。'
  if (reading.value.status === 'error') return '首轮解读失败；修复后可继续追问。'
  if (reading.value.mode === 'offline') return '当前是离线摘要，未配置模型；配置后可沿原盘追问。'
  return '首轮四段先落定，再开始追问。'
})

const GUAS = ['初爻', '二爻', '三爻', '四爻', '五爻', '上爻']
const CN_NUMS = ['一', '二', '三', '四', '五', '六']
const ORDER = [5, 4, 3, 2, 1, 0]

interface TossResult {
  coins: number[]
  backs: number
  kind: string
  moving: boolean
}

const strip = computed<(TossResult | null)[]>(() => {
  const out: (TossResult | null)[] = []
  for (let i = 0; i < 6; i++) out.push(evaluate(tosses.value[i]))
  return out
})

const spinBacks = computed(() => spinCoins.value.reduce((a, b) => a + b, 0))
const spinKind = computed(() => kindOf(spinBacks.value))
const spinMoving = computed(() => spinKind.value.startsWith('老'))
const boardRevealed = 6

const coreFacts = computed(() => chart.value?.facts ?? [])

const currentArchiveEntry = computed(() => {
  const current = chart.value
  if (!current) return null
  return archive.value.find((entry) => entry.chart_id === current.chart_id) ?? null
})

const currentCaseOrdinal = computed(() => {
  const current = chart.value
  if (!current) return 0
  const index = archive.value.findIndex((entry) => entry.chart_id === current.chart_id)
  return index >= 0 ? archive.value.length - index : 0
})

interface ReadingSections {
  overview: string
  verdict: string
  timing: string
  advice: string
}

const readingBusy = computed(
  () => reading.value.status === 'loading' || reading.value.status === 'streaming'
)

const readingStatusText = computed(() => {
  const r = reading.value
  if (r.status === 'loading') return '正在连接解读…'
  if (r.status === 'streaming') return r.mode === 'ai' ? `${r.model} · 流式解读` : '离线摘要 · 生成中'
  if (r.status === 'done') return r.mode === 'ai' ? `${r.model} · 解读完成` : '未配置模型 · 离线摘要'
  if (r.status === 'error') return '解读失败 · 盘面仍可查看'
  return '只列盘面事实'
})

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
  const raw: Record<keyof ReadingSections, string[]> = {
    overview: [],
    verdict: [],
    timing: [],
    advice: []
  }
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

const readingSections = computed(() => parseReading(reading.value.text))

const localBacks = computed(() => tosses.value.map((coins) => coins.reduce((a, b) => a + b, 0)))

const questionReady = computed(() => question.value.trim().length > 0)
const needsQuestion = computed(() => method.value !== 'ziwei' && !questionReady.value)
const birthReady = computed(
  () => !!birth.value.birth_date.trim() && !!birth.value.birth_time.trim() && !!birth.value.gender
)
const methodReady = computed(() => {
  if (method.value === 'ziwei') return birthReady.value
  if (method.value === 'daliuren') return !!daliurenAt.value && questionReady.value
  if (method.value === 'xiaoliuren' || method.value === 'qimen') {
    return !!quickAt.value && questionReady.value
  }
  if (method.value === 'meihua') {
    return questionReady.value && (meihuaMode.value === 'time' || meihuaNumbersReady.value)
  }
  return true
})

const castLabel = computed(() => {
  if (submitting.value) {
    if (method.value === 'liuyao') return '成卦中'
    if (method.value === 'daliuren' || method.value === 'xiaoliuren') return '起课中'
    if (method.value === 'qimen') return '起局中'
    return '排盘中'
  }
  if (shaking.value) return '摇卦中'
  if (chart.value) {
    if (method.value === 'ziwei') return '再 排 盘'
    if (method.value === 'daliuren' || method.value === 'xiaoliuren') return '再 起 课'
    if (method.value === 'qimen') return '再 起 局'
    return '再 起 卦'
  }
  if (method.value === 'ziwei') return '排 命 盘'
  if (method.value === 'daliuren' || method.value === 'xiaoliuren') {
    return questionReady.value ? '起 课' : '先 问 一 事'
  }
  if (method.value === 'qimen') return questionReady.value ? '起 局' : '先 问 一 事'
  if (method.value === 'meihua') {
    if (!questionReady.value) return '先 问 一 事'
    if (meihuaMode.value === 'numbers' && !meihuaNumbersReady.value) return '补 两 个 数'
    return '起 卦'
  }
  if (!questionReady.value) return '先 问 一 事'
  if (mode.value === 'manual') return '起 卦'
  if (tosses.value.length >= 6) return '重试成卦'
  return `摇 ${GUAS[tosses.value.length]}`
})

const castDisabled = computed(() => shaking.value || submitting.value || !methodReady.value)

function syncFromHash() {
  const next = window.location.hash.replace(/^#\/?/, '') as ViewName
  if (VALID_VIEWS.includes(next)) view.value = next
}

function navigate(next: ViewName) {
  view.value = next
  if (window.location.hash !== `#/${next}`) window.location.hash = `#/${next}`
  if (next === 'cast') nextTick(() => questionEl.value?.focus())
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function normalizeFollowupList(input: unknown): FollowupMessage[] {
  if (!Array.isArray(input)) return []
  return input
    .map((item, index): FollowupMessage | null => {
      if (!item || typeof item !== 'object') return null
      const raw = item as Partial<FollowupMessage>
      if (raw.role !== 'user' && raw.role !== 'assistant') return null
      const content = typeof raw.content === 'string' ? raw.content : ''
      const error = typeof raw.error === 'string' ? raw.error : undefined
      if (!content.trim() && !error) return null
      const status: FollowupMessage['status'] =
        raw.status === 'streaming' || raw.status === 'done' || raw.status === 'error'
          ? raw.status === 'streaming'
            ? content.trim()
              ? 'done'
              : 'error'
            : raw.status
          : error
            ? 'error'
            : 'done'
      return {
        id: typeof raw.id === 'string' && raw.id ? raw.id : `followup-${index}-${Date.now()}`,
        role: raw.role,
        content,
        created_at:
          typeof raw.created_at === 'string' && raw.created_at
            ? raw.created_at
            : new Date().toISOString(),
        model: typeof raw.model === 'string' ? raw.model : undefined,
        status,
        error
      }
    })
    .filter((item): item is FollowupMessage => !!item)
    .slice(-MAX_FOLLOWUPS * 2)
}

function summaryFields(payload: ChartPayload) {
  if (isLiuYaoPayload(payload)) {
    return {
      title: payload.gua_name,
      subtitle: payload.changed_gua ? `之 ${payload.changed_gua}` : '六爻安静',
      guaName: payload.gua_name,
      changedGua: payload.changed_gua || '',
      topic: payload.use_shen?.topic_label || ''
    }
  }
  if (isDaLiuRenPayload(payload)) {
    return {
      title: `大六壬 · ${payload.ke_ti}`,
      subtitle: `${payload.month_general.zhi}将 · ${payload.gate}课 · ${payload.san_chuan.map((c) => c.branch).join('→')}`,
      guaName: '',
      changedGua: '',
      topic: ''
    }
  }
  if (isXiaoLiuRenPayload(payload)) {
    return {
      title: `小六壬 · ${payload.final_palace.name}`,
      subtitle: `月${payload.month_palace.name} · 日${payload.day_palace.name} · 时${payload.final_palace.name}`,
      guaName: '',
      changedGua: '',
      topic: ''
    }
  }
  if (isMeiHuaPayload(payload)) {
    return {
      title: `梅花 · ${payload.ben_gua.name}`,
      subtitle: `之 ${payload.bian_gua.name} · 第${payload.moving_line}爻 · ${payload.relation}`,
      guaName: payload.ben_gua.name,
      changedGua: payload.bian_gua.name,
      topic: ''
    }
  }
  if (isQiMenPayload(payload)) {
    return {
      title: `奇门 · ${payload.dun}${payload.ju}局`,
      subtitle: `值符${payload.zhi_fu} · 值使${payload.zhi_shi} · ${payload.solar_term}${payload.yuan}`,
      guaName: '',
      changedGua: '',
      topic: ''
    }
  }
  return {
    title: `${payload.ming_gong}宫 · ${payload.wu_xing_ju}`,
    subtitle: `紫微在${payload.ziwei_pos} · ${payload.sihua_year.map((s) => s.star + '化' + s.hua).join('、')}`,
    guaName: '',
    changedGua: '',
    topic: ''
  }
}

function normalizeArchiveEntry(item: unknown): ArchiveEntry | null {
  if (!item || typeof item !== 'object') return null
  const raw = item as Partial<ArchiveEntry>
  if (!raw.chart_id || !raw.chart || !raw.chart.payload) return null

  const chart = raw.chart
  const reading = typeof raw.reading === 'string' ? raw.reading : ''
  const mode = raw.mode === 'ai' || raw.mode === 'offline' ? raw.mode : ''
  const savedAt = raw.saved_at || chart.created_at || new Date().toISOString()
  const status =
    raw.status === 'pending' || raw.status === 'reading' || raw.status === 'done' || raw.status === 'error'
      ? raw.status
      : reading
        ? 'done'
        : raw.reading_error
          ? 'error'
          : 'pending'
  const fields = summaryFields(chart.payload)

  return {
    chart_id: raw.chart_id,
    question: raw.question || chart.question || '',
    gua_name: raw.gua_name || fields.guaName,
    changed_gua: raw.changed_gua || fields.changedGua,
    title: raw.title || fields.title,
    subtitle: raw.subtitle || fields.subtitle,
    method: raw.method || chart.method,
    saved_at: savedAt,
    cast_at: raw.cast_at || chart.created_at || savedAt,
    read_at: raw.read_at || (reading ? savedAt : undefined),
    status,
    topic_label: raw.topic_label || fields.topic,
    mode,
    model: raw.model || '',
    reading,
    reading_error: raw.reading_error,
    followups: normalizeFollowupList(raw.followups),
    chart
  }
}

function loadArchive() {
  try {
    const raw = window.localStorage.getItem(ARCHIVE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return
    archive.value = parsed
      .map((item) => normalizeArchiveEntry(item))
      .filter((item): item is ArchiveEntry => !!item)
      .slice(0, ARCHIVE_LIMIT)
  } catch {
    // 本地存档损坏时静默丢弃，不影响起卦。
  }
}

function persistArchive() {
  try {
    window.localStorage.setItem(ARCHIVE_KEY, JSON.stringify(archive.value))
  } catch {
    // 隐私模式下 localStorage 可能不可写；不阻断当前解读。
  }
}

function upsertArchiveEntry(
  chartData: ChartView,
  patch: Partial<
    Pick<
      ArchiveEntry,
      'question' | 'status' | 'mode' | 'model' | 'reading' | 'reading_error' | 'read_at' | 'followups'
    >
  >
): ArchiveEntry {
  const existing = archive.value.find((item) => item.chart_id === chartData.chart_id)
  const now = new Date().toISOString()
  const fields = summaryFields(chartData.payload)
  const entry: ArchiveEntry = {
    chart_id: chartData.chart_id,
    question: patch.question ?? chartData.question ?? question.value.trim(),
    gua_name: fields.guaName,
    changed_gua: fields.changedGua,
    title: fields.title,
    subtitle: fields.subtitle,
    method: chartData.method,
    saved_at: existing?.saved_at || now,
    cast_at: existing?.cast_at || chartData.created_at || now,
    read_at: patch.read_at ?? existing?.read_at,
    status: patch.status ?? existing?.status ?? 'pending',
    topic_label: fields.topic || existing?.topic_label,
    mode: patch.mode ?? existing?.mode ?? '',
    model: patch.model ?? existing?.model ?? '',
    reading: patch.reading ?? existing?.reading ?? '',
    reading_error: patch.reading_error ?? existing?.reading_error,
    followups: patch.followups ?? existing?.followups ?? [],
    chart: chartData
  }
  archive.value = [entry, ...archive.value.filter((item) => item.chart_id !== entry.chart_id)].slice(
    0,
    ARCHIVE_LIMIT
  )
  persistArchive()
  return entry
}

function recordCurrentChart(chartData: ChartView) {
  const existing = archive.value.find((item) => item.chart_id === chartData.chart_id)
  if (existing) return
  upsertArchiveEntry(chartData, {
    status: 'pending',
    mode: '',
    model: '',
    reading: '',
    reading_error: ''
  })
}

function saveCurrentReading() {
  const current = chart.value
  const text = reading.value.text.trim()
  if (!current || !text) return
  upsertArchiveEntry(current, {
    status: 'done',
    mode: reading.value.mode,
    model: reading.value.model,
    reading: text,
    read_at: new Date().toISOString(),
    reading_error: ''
  })
}

function markReadingError() {
  const current = chart.value
  if (!current) return
  upsertArchiveEntry(current, {
    status: 'error',
    reading_error: reading.value.error || '解读失败',
    read_at: new Date().toISOString()
  })
}

function deleteArchiveEntry(entry: ArchiveEntry) {
  archive.value = archive.value.filter((item) => item.chart_id !== entry.chart_id)
  persistArchive()
}

function clearArchive() {
  if (!archive.value.length) return
  archive.value = []
  persistArchive()
}

function exportArchive() {
  if (!archive.value.length) return
  const blob = new Blob([JSON.stringify(archive.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `mingyun-history-${new Date().toISOString().slice(0, 10)}.json`
  a.click()
  window.setTimeout(() => URL.revokeObjectURL(url), 1000)
}

function openArchive(entry: ArchiveEntry) {
  if (!entry.chart) return
  readingAbort?.abort()
  readingAbort = null
  readingSeq += 1
  followupAbort?.abort()
  followupAbort = null
  followupSeq += 1
  followups.value = normalizeFollowupList(entry.followups)
  followupInput.value = ''
  followupError.value = ''
  followupBusy.value = false
  chart.value = entry.chart
  method.value = (entry.chart.method as CastMethod) || 'liuyao'
  question.value = entry.question || entry.chart.question || ''
  if (entry.chart.calendar) calendar.value = entry.chart.calendar
  reading.value = {
    status: entry.status === 'error' ? 'error' : entry.reading ? 'done' : 'idle',
    mode: entry.mode,
    model: entry.model,
    provider: '',
    text: entry.reading,
    error: entry.reading_error || ''
  }
  view.value = 'cast'
  window.history.replaceState(null, '', '#/cast')
  nextTick(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  })
}

onMounted(async () => {
  loadArchive()
  syncFromHash()
  window.addEventListener('hashchange', syncFromHash)

  const params = new URLSearchParams(window.location.search)
  const presetMethod = params.get('method')
  if (
    presetMethod === 'daliuren' ||
    presetMethod === 'ziwei' ||
    presetMethod === 'liuyao' ||
    presetMethod === 'xiaoliuren' ||
    presetMethod === 'meihua' ||
    presetMethod === 'qimen'
  ) {
    method.value = presetMethod
  }
  daliurenAt.value = localDateTimeInput()
  quickAt.value = localDateTimeInput()
  const presetAt = params.get('at')
  if (presetAt) replayAt = presetAt
  previewMode.value = params.get('cast') === '1'
  const preset = params.get('q')
  if (preset) question.value = preset

  const presetLines = params.get('lines')
  if (presetLines) {
    const lines = presetLines.split(/[,，\s]+/).filter((l) => lineOptions.includes(l)).slice(0, 6)
    if (lines.length === 6) {
      mode.value = 'manual'
      manualLines.value = lines
    }
  }

  const presetBacks = parseBacks(params.get('backs'))
  if (presetBacks && !presetLines) tosses.value = presetBacks

  try {
    calendar.value = await fetchCalendar(replayAt || undefined)
  } catch {
    // 盘头信息拿不到时静默降级，不打断起卦。
  }

  if (params.get('cast') === '1') {
    view.value = 'cast'
    window.history.replaceState(null, '', '#/cast')
    await autoCast()
  }
})

onUnmounted(() => {
  window.removeEventListener('hashchange', syncFromHash)
})

function parseBacks(raw: string | null): number[][] | null {
  if (!raw) return null
  const nums = raw.split(',').map((s) => Number(s.trim()))
  if (nums.length < 1 || nums.length > 6) return null
  if (nums.some((n) => !Number.isInteger(n) || n < 0 || n > 3)) return null
  return nums.map((b) => [1, 1, 1].map((_, j) => (j < b ? 1 : 0)))
}

function evaluate(coins: number[] | undefined): TossResult | null {
  if (!coins || coins.length !== 3) return null
  const backs = coins.reduce((a, b) => a + b, 0)
  return { coins, backs, kind: kindOf(backs), moving: kindOf(backs).startsWith('老') }
}

function kindOf(backs: number): string {
  return backs === 0 ? '老阴' : backs === 1 ? '少阳' : backs === 2 ? '少阴' : '老阳'
}

function randCoin(): number {
  const buf = new Uint32Array(1)
  crypto.getRandomValues(buf)
  return buf[0] & 1
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

function isYang(i: number): boolean {
  const t = strip.value[i]
  return !!t && t.backs % 2 === 1
}

function localDateTimeInput(date = new Date()): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function resetBoard() {
  resetReading()
  chart.value = null
  tosses.value = []
  error.value = ''
}

function startMethod(next: CastMethod) {
  method.value = next
  if (next === 'daliuren' && !daliurenAt.value) daliurenAt.value = localDateTimeInput()
  if ((next === 'xiaoliuren' || next === 'qimen') && !quickAt.value) quickAt.value = localDateTimeInput()
  if (chart.value && chart.value.method !== next) resetBoard()
  navigate('cast')
  nextTick(() => questionEl.value?.focus())
}

function onHomeStart(next?: CastMethod) {
  if (next) {
    startMethod(next)
    return
  }
  if (chart.value) {
    method.value = (chart.value.method as CastMethod) || 'liuyao'
    navigate('cast')
    return
  }
  startMethod('liuyao')
}

function switchMethod(next: CastMethod) {
  if (next === method.value) return
  method.value = next
  if (next === 'daliuren' && !daliurenAt.value) daliurenAt.value = localDateTimeInput()
  if ((next === 'xiaoliuren' || next === 'qimen') && !quickAt.value) quickAt.value = localDateTimeInput()
  resetBoard()
  error.value = ''
}

function setMode(next: 'coin' | 'manual') {
  if (shaking.value || submitting.value || mode.value === next) return
  resetReading()
  mode.value = next
  chart.value = null
  tosses.value = []
  error.value = ''
}

function setMeiHuaMode(next: 'time' | 'numbers') {
  if (submitting.value || meihuaMode.value === next) return
  resetBoard()
  meihuaMode.value = next
}

function reset() {
  resetReading()
  question.value = ''
  chart.value = null
  tosses.value = []
  error.value = ''
}

function startAnother() {
  reset()
  nextTick(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' })
    questionEl.value?.focus()
  })
}

function resetFollowups() {
  followupAbort?.abort()
  followupAbort = null
  followupSeq += 1
  followups.value = []
  followupInput.value = ''
  followupError.value = ''
  followupBusy.value = false
}

function resetReading() {
  readingAbort?.abort()
  readingAbort = null
  readingSeq += 1
  resetFollowups()
  reading.value = {
    status: 'idle',
    mode: '',
    model: '',
    provider: '',
    text: '',
    error: ''
  }
}

async function startReading() {
  const current = chart.value
  if (!current) return

  readingAbort?.abort()
  const seq = ++readingSeq
  const ctrl = new AbortController()
  readingAbort = ctrl
  reading.value = {
    status: 'loading',
    mode: '',
    model: '',
    provider: '',
    text: '',
    error: ''
  }

  try {
    await streamInterpret(
      current.chart_id,
      current.question || question.value.trim() || '',
      (ev: InterpretEvent) => {
        if (seq !== readingSeq) return
        const r = reading.value
        if (ev.event === 'meta') {
          r.mode = ev.data.mode
          r.model = ev.data.model
          r.provider = ev.data.provider
          r.status = 'streaming'
        } else if (ev.event === 'delta') {
          r.text += ev.data.text
          r.status = 'streaming'
        } else if (ev.event === 'done') {
          r.status = 'done'
        } else if (ev.event === 'error') {
          r.status = 'error'
          r.error = ev.data.message || '解读服务暂时不可用'
        }
      },
      ctrl.signal
    )
    if (seq === readingSeq && reading.value.status === 'streaming') {
      reading.value.status = 'done'
    }
    if (seq === readingSeq && reading.value.status === 'done') {
      saveCurrentReading()
    } else if (seq === readingSeq && reading.value.status === 'error') {
      markReadingError()
    }
  } catch (e) {
    if (ctrl.signal.aborted || seq !== readingSeq) return
    reading.value.status = 'error'
    reading.value.error = e instanceof Error ? e.message : '解读失败'
    markReadingError()
  } finally {
    if (readingAbort === ctrl) readingAbort = null
  }
}

function lineState(v: string): string {
  return v.startsWith('老') ? '动' : '静'
}

function buildFollowupHistory(): { role: 'user' | 'assistant'; content: string }[] {
  const opening: { role: 'user' | 'assistant'; content: string }[] = []
  if (reading.value.text.trim()) {
    opening.push({ role: 'assistant', content: reading.value.text.trim() })
  }
  const turns: { role: 'user' | 'assistant'; content: string }[] = []
  for (let i = 0; i < followups.value.length; i += 1) {
    const question = followups.value[i]
    const answer = followups.value[i + 1]
    if (question.role !== 'user' || !question.content.trim()) continue
    if (!answer || answer.role !== 'assistant' || answer.status !== 'done' || !answer.content.trim()) {
      continue
    }
    turns.push({ role: 'user', content: question.content.trim() })
    turns.push({ role: 'assistant', content: answer.content.trim() })
    i += 1
  }
  // 首轮解读是同一张盘的基准，不能因为追问轮次多而被截掉。
  // 6 轮上限内最多 1 条首轮 + 10 条追问，服务端最多接收 12 条。
  return [...opening, ...turns.slice(-(12 - opening.length))]
}

function saveFollowups() {
  const current = chart.value
  if (!current) return
  upsertArchiveEntry(current, {
    followups: followups.value.map((turn) => ({
      id: turn.id,
      role: turn.role,
      content: turn.content,
      created_at: turn.created_at,
      model: turn.model,
      status: turn.status === 'streaming' ? 'done' : turn.status,
      error: turn.error
    }))
  })
}

function scrollFollowupThread() {
  nextTick(() => {
    followupThreadEl.value?.lastElementChild?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  })
}

function onFollowupKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return
  event.preventDefault()
  void sendFollowup()
}

async function sendFollowup() {
  if (!followupReady.value || followupBusy.value) return
  const current = chart.value
  if (!current) return
  const message = followupInput.value.trim()
  if (!message) return

  followupAbort?.abort()
  const seq = ++followupSeq
  const ctrl = new AbortController()
  followupAbort = ctrl
  followupBusy.value = true
  followupError.value = ''

  const history = buildFollowupHistory()
  const stamp = new Date().toISOString()
  const userTurn: FollowupMessage = {
    id: `u-${Date.now()}-${seq}`,
    role: 'user',
    content: message,
    created_at: stamp,
    status: 'done'
  }
  const assistantTurn: FollowupMessage = {
    id: `a-${Date.now()}-${seq}`,
    role: 'assistant',
    content: '',
    created_at: new Date().toISOString(),
    model: reading.value.model,
    status: 'streaming'
  }
  followups.value.push(userTurn, assistantTurn)
  followupInput.value = ''
  scrollFollowupThread()

  try {
    await streamFollowup(
      current,
      message,
      history,
      (ev: InterpretEvent) => {
        if (seq !== followupSeq) return
        const turn = followups.value.find((item) => item.id === assistantTurn.id)
        if (!turn) return
        if (ev.event === 'meta') {
          turn.model = ev.data.model || turn.model
          turn.status = 'streaming'
        } else if (ev.event === 'delta') {
          turn.content += ev.data.text
          turn.status = 'streaming'
        } else if (ev.event === 'done') {
          turn.status = 'done'
        } else if (ev.event === 'error') {
          turn.status = 'error'
          turn.error = ev.data.message || '追问服务暂时不可用'
          followupError.value = turn.error
        }
      },
      ctrl.signal
    )
    if (seq !== followupSeq) return
    const turn = followups.value.find((item) => item.id === assistantTurn.id)
    if (turn && turn.status === 'streaming') turn.status = 'done'
    saveFollowups()
  } catch (e) {
    if (ctrl.signal.aborted || seq !== followupSeq) return
    const turn = followups.value.find((item) => item.id === assistantTurn.id)
    const messageText = e instanceof Error ? e.message : '追问失败'
    if (turn) {
      turn.status = 'error'
      turn.error = messageText
    }
    followupError.value = messageText
    saveFollowups()
  } finally {
    if (followupAbort === ctrl) followupAbort = null
    if (seq === followupSeq) followupBusy.value = false
  }
}

function scrollToBoard() {
  if (previewMode.value) return
  boardEl.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function onCast() {
  if (shaking.value || submitting.value) return
  if (method.value === 'daliuren') {
    await submitDaLiuRen()
    return
  }
  if (method.value === 'ziwei') {
    await submitZiWei()
    return
  }
  if (method.value === 'xiaoliuren') {
    if (chart.value) {
      reset()
      return
    }
    await submitXiaoLiuRen()
    return
  }
  if (method.value === 'meihua') {
    if (chart.value) {
      reset()
      return
    }
    await submitMeiHua()
    return
  }
  if (method.value === 'qimen') {
    if (chart.value) {
      reset()
      return
    }
    await submitQiMen()
    return
  }
  if (!chart.value && !questionReady.value) {
    questionEl.value?.focus()
    return
  }
  if (mode.value === 'manual') {
    if (chart.value) {
      reset()
      return
    }
    await submitManual()
    return
  }
  if (chart.value) {
    reset()
    return
  }
  if (tosses.value.length >= 6) {
    await submitCoinChart()
    return
  }
  await shakeOne()
}

async function shakeOne() {
  if (tosses.value.length >= 6) return
  error.value = ''
  shakingIndex.value = tosses.value.length
  shaking.value = true
  landed.value = false

  const finalCoins = [randCoin(), randCoin(), randCoin()]
  spinCoins.value = [randCoin(), randCoin(), randCoin()]

  const flicker = window.setInterval(() => {
    spinCoins.value = [randCoin(), randCoin(), randCoin()]
  }, 80)

  await sleep(720)
  window.clearInterval(flicker)
  spinCoins.value = finalCoins
  await sleep(200)
  landed.value = true
  await sleep(280)

  tosses.value.push(finalCoins)
  shaking.value = false
  landed.value = false

  if (tosses.value.length === 6) await submitCoinChart()
}

async function submitCoinChart() {
  submitting.value = true
  error.value = ''
  try {
    const data = await castLiuYao({
      mode: 'coin',
      question: question.value.trim(),
      coins: tosses.value.map((c) => [...c]),
      ...(replayAt ? { at: replayAt } : {})
    })
    chart.value = data
    if (data.calendar) calendar.value = data.calendar
    recordCurrentChart(data)
    await nextTick()
    window.setTimeout(scrollToBoard, 260)
    void startReading()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '排盘失败'
  } finally {
    submitting.value = false
  }
}

async function submitManual() {
  submitting.value = true
  error.value = ''
  try {
    const data = await castLiuYao({
      mode: 'manual',
      question: question.value.trim(),
      lines: manualLines.value,
      ...(replayAt ? { at: replayAt } : {})
    })
    chart.value = data
    if (data.calendar) calendar.value = data.calendar
    recordCurrentChart(data)
    await nextTick()
    window.setTimeout(scrollToBoard, 200)
    void startReading()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '排盘失败'
  } finally {
    submitting.value = false
  }
}

async function autoCast() {
  if (method.value !== 'liuyao') return
  if (mode.value === 'manual') {
    await submitManual()
    return
  }
  if (tosses.value.length !== 6) {
    tosses.value = Array.from({ length: 6 }, () => [randCoin(), randCoin(), randCoin()])
  }
  await nextTick()
  await submitCoinChart()
}

async function submitMethodChart(next: CastMethod, req: CastRequest) {
  submitting.value = true
  error.value = ''
  try {
    const data = await castChart(next, req)
    chart.value = data
    if (data.calendar) calendar.value = data.calendar
    recordCurrentChart(data)
    await nextTick()
    window.setTimeout(scrollToBoard, 260)
    void startReading()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '排盘失败'
  } finally {
    submitting.value = false
  }
}

async function submitDaLiuRen() {
  if (!daliurenAt.value) {
    error.value = '请先确认起课时间'
    return
  }
  if (!questionReady.value) {
    error.value = '大六壬需要先写明所问之事'
    questionEl.value?.focus()
    return
  }
  await submitMethodChart('daliuren', {
    question: question.value.trim(),
    at: new Date(daliurenAt.value).toISOString()
  })
}

async function submitXiaoLiuRen() {
  if (!questionReady.value) {
    error.value = '小六壬需要先写明所问之事'
    questionEl.value?.focus()
    return
  }
  if (!quickAt.value) {
    error.value = '请先确认起课时间'
    return
  }
  await submitMethodChart('xiaoliuren', {
    question: question.value.trim(),
    at: new Date(quickAt.value).toISOString()
  })
}

async function submitMeiHua() {
  if (!questionReady.value) {
    error.value = '梅花易数需要先写明所问之事'
    questionEl.value?.focus()
    return
  }
  if (meihuaMode.value === 'numbers' && !meihuaNumbersReady.value) {
    error.value = '请填写两个正整数再报数起卦'
    return
  }
  const req: CastRequest = {
    question: question.value.trim(),
    mode: meihuaMode.value
  }
  if (meihuaMode.value === 'numbers') {
    req.numbers = [Number(meihuaA.value), Number(meihuaB.value)]
  } else {
    req.at = new Date(quickAt.value || localDateTimeInput()).toISOString()
  }
  await submitMethodChart('meihua', req)
}

async function submitQiMen() {
  if (!questionReady.value) {
    error.value = '奇门遁甲需要先写明所问之事'
    questionEl.value?.focus()
    return
  }
  if (!quickAt.value) {
    error.value = '请先确认起局时间'
    return
  }
  await submitMethodChart('qimen', {
    question: question.value.trim(),
    at: new Date(quickAt.value).toISOString()
  })
}

async function submitZiWei() {
  if (!birthReady.value) {
    error.value = '请补全出生日期、时辰与性别'
    return
  }
  await submitMethodChart('ziwei', {
    question: question.value.trim(),
    gender: birth.value.gender,
    calendar_type: birth.value.calendar_type,
    birth_date: birth.value.birth_date,
    birth_time: birth.value.birth_time,
    is_leap_month: birth.value.is_leap_month,
    leap_month_policy: birth.value.leap_month_policy
  })
}

function formatTime(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleString('zh-CN', { hour12: false })
}
</script>

<template>
  <div class="game-shell">
    <header class="game-topbar">
      <button class="brand" type="button" @click="navigate('home')">
        <svg class="brand-mark" viewBox="0 0 32 32" aria-hidden="true">
          <rect width="32" height="32" fill="#9e3327" />
          <text
            x="16"
            y="23"
            text-anchor="middle"
            font-family="'Songti SC','STSong','SimSun',serif"
            font-size="18"
            fill="#f8f3e8"
          >命</text>
        </svg>
        命云
        <span>MINGYUN</span>
      </button>
      <div class="game-hud">
        <span class="hud-cell"><em>历史</em><b>{{ archive.length }}</b></span>
        <span v-if="calendar" class="hud-cell"><em>日</em><b>{{ calendar.day_gz }}</b></span>
        <span class="hud-cell hud-cell--quiet"><em>六术</em><b>{{ methodLabel }}</b></span>
        <a href="https://github.com/huayuehua15-netizen/mingyun" target="_blank" rel="noopener">GitHub ↗</a>
      </div>
    </header>

    <GameHome
        v-if="view === 'home'"
        key="home"
        :calendar="calendar"
        :archive-count="archive.length"
        :has-chart="!!chart"
        :last-entry="archive[0] || null"
        @start="onHomeStart"
        @archive="navigate('archive')"
        @codex="navigate('codex')"
      />

      <section v-else-if="view === 'cast'" key="cast" class="scene-view cast-view">
        <span class="view-watermark" aria-hidden="true">{{ methodMeta.glyph }}</span>
        <header class="scene-bar">
          <button class="scene-back" @click="navigate('home')">← 山门</button>
          <div class="scene-heading">
            <span>{{ methodMeta.volume }} · {{ methodLabel }}</span>
            <h1>{{ castTitle }}</h1>
          </div>
          <div class="scene-steps" :aria-label="`${methodLabel}起盘进度`">
            <template v-for="(step, i) in methodMeta.steps" :key="step">
              <span :class="{ 'is-done': stepDone(i) }">{{ step }}</span>
              <i v-if="i < 2"></i>
            </template>
          </div>
        </header>

        <div class="scene-body">
          <section id="cast" class="cast-panel">
      <div class="method-tabs" role="tablist" aria-label="选择术数">
        <button
          v-for="item in methodTabs"
          :key="item.key"
          class="method-tab"
          :class="{ 'is-active': method === item.key }"
          role="tab"
          :aria-selected="method === item.key"
          @click="switchMethod(item.key)"
        >
          <b>{{ item.name }}</b>
          <span>{{ item.note }}</span>
        </button>
      </div>

      <div class="cast-row" :class="{ 'is-simple': method !== 'liuyao' }">
        <div class="ask">
          <label for="question">
            所问之事
            <span>{{ method === 'ziwei' ? '可留空 · 只看命盘' : 'QUESTION' }}</span>
          </label>
          <input
            id="question"
            ref="questionEl"
            v-model="question"
            maxlength="200"
            :placeholder="
              method === 'ziwei'
                ? '例如：今年的事业与感情如何？'
                : method === 'daliuren'
                  ? '例如：这桩合作能否谈成？'
                  : method === 'meihua'
                    ? '例如：这次比赛能否进决赛？'
                    : method === 'qimen'
                      ? '例如：这次考试近期走势如何？'
                      : method === 'xiaoliuren'
                        ? '例如：这次面试结果如何？'
                        : '例如：这次考试能否通过？'
            "
            @keydown.enter.prevent="onCast"
          />
        </div>
        <div v-if="method === 'liuyao'" class="mode-switch">
          <button class="mode-btn" :class="{ 'is-active': mode === 'coin' }" @click="setMode('coin')">
            铜钱摇卦
          </button>
          <button class="mode-btn" :class="{ 'is-active': mode === 'manual' }" @click="setMode('manual')">
            手动录入
          </button>
        </div>
        <button
          class="cast-btn"
          :class="{ 'is-need-question': needsQuestion, 'is-tossing': shaking }"
          :disabled="castDisabled"
          @click="onCast"
        >
          {{ castLabel }}
        </button>
      </div>

      <div v-if="method === 'daliuren'" class="method-form">
        <div class="method-form-copy">
          <span>起课时间</span>
          <p>月将加时，天地盘随时而定。默认取此刻，也可回到想问的那一分钟。</p>
        </div>
        <div class="method-fields">
          <label>
            <span>时刻</span>
            <input v-model="daliurenAt" type="datetime-local" />
          </label>
          <button class="text-btn" type="button" @click="daliurenAt = localDateTimeInput()">用此刻</button>
        </div>
      </div>

      <div v-if="method === 'xiaoliuren' || method === 'qimen'" class="method-form">
        <div class="method-form-copy">
          <span>{{ method === 'xiaoliuren' ? '起课时间' : '起局时间' }}</span>
          <p v-if="method === 'xiaoliuren'">
            小六壬以农历月、日、时落六宫。月看背景，日看事体，时看当下动向，适合近期一件具体事。
          </p>
          <p v-else>
            时家奇门以节气定阴阳遁与局数，再布九宫。此处只排基础转盘，不套格局与十干克应。
          </p>
        </div>
        <div class="method-fields">
          <label>
            <span>时刻</span>
            <input v-model="quickAt" type="datetime-local" />
          </label>
          <button class="text-btn" type="button" @click="quickAt = localDateTimeInput()">用此刻</button>
        </div>
      </div>

      <div v-if="method === 'meihua'" class="method-form">
        <div class="method-form-copy">
          <span>起卦方式</span>
          <p>时间起卦取年支、农历月日时；报数起卦取两组数。动爻只取一处，体用按动爻所在卦为用。</p>
        </div>
        <div class="mode-switch">
          <button class="mode-btn" :class="{ 'is-active': meihuaMode === 'time' }" @click="setMeiHuaMode('time')">
            时间起卦
          </button>
          <button class="mode-btn" :class="{ 'is-active': meihuaMode === 'numbers' }" @click="setMeiHuaMode('numbers')">
            报数起卦
          </button>
        </div>
        <div v-if="meihuaMode === 'time'" class="method-fields">
          <label>
            <span>时刻</span>
            <input v-model="quickAt" type="datetime-local" />
          </label>
          <button class="text-btn" type="button" @click="quickAt = localDateTimeInput()">用此刻</button>
        </div>
        <div v-else class="method-fields">
          <label>
            <span>上卦数</span>
            <input v-model="meihuaA" type="number" min="1" max="999999" inputmode="numeric" />
          </label>
          <label>
            <span>下卦数</span>
            <input v-model="meihuaB" type="number" min="1" max="999999" inputmode="numeric" />
          </label>
        </div>
      </div>

      <div v-if="method === 'ziwei'" class="method-form">
        <div class="method-form-copy">
          <span>出生信息</span>
          <p>紫微以生辰安命身、定五行局、布十二宫。公历农历都可，闰月按所选口径处理。</p>
        </div>
        <div class="method-fields is-grid">
          <label>
            <span>性别</span>
            <select v-model="birth.gender">
              <option value="male">男 · 乾造</option>
              <option value="female">女 · 坤造</option>
            </select>
          </label>
          <label>
            <span>历法</span>
            <select v-model="birth.calendar_type">
              <option value="solar">公历</option>
              <option value="lunar">农历</option>
            </select>
          </label>
          <label>
            <span>出生日期</span>
            <input v-model="birth.birth_date" type="date" />
          </label>
          <label>
            <span>出生时辰</span>
            <input v-model="birth.birth_time" type="time" />
          </label>
          <label class="method-check">
            <input v-model="birth.is_leap_month" type="checkbox" />
            <span>农历闰月</span>
          </label>
          <label>
            <span>闰月口径</span>
            <select v-model="birth.leap_month_policy">
              <option value="as_this_month">按本月</option>
              <option value="as_next_month">按下月</option>
            </select>
          </label>
        </div>
        <p class="method-hint">默认东经 113.39°、真太阳时开启；子时按 23:00 换日。</p>
      </div>

      <div v-if="method === 'liuyao' && mode === 'coin'" class="coin-strip">
        <div
          v-if="shaking"
          class="toss-overlay"
          :class="{ 'is-landed': landed }"
          role="status"
          aria-live="polite"
        >
          <div class="toss-caption">
            <span class="toss-num">{{ CN_NUMS[shakingIndex] }}</span>
            <div>
              <b>摇{{ GUAS[shakingIndex] }}</b>
              <em>三枚铜钱 · 自下而上</em>
            </div>
          </div>
          <div class="toss-coins">
            <span
              v-for="(c, j) in spinCoins"
              :key="j"
              class="toss-slot"
              :style="{ '--d': `${j * 60}ms` }"
            >
              <i class="coin-shadow"></i>
              <span class="toss-coin" :class="{ 'is-back': c === 1 }"></span>
            </span>
          </div>
          <div class="toss-result" :class="{ 'is-show': landed }">
            <template v-if="landed">
              <strong>背{{ spinBacks }}</strong>
              <span>{{ spinKind }} · {{ spinMoving ? '动' : '静' }}</span>
            </template>
            <template v-else>铜钱翻转中</template>
          </div>
        </div>

        <div
          v-for="(t, i) in strip"
          :key="i"
          class="cast-cell"
          :class="{ 
            'is-fresh': !!t && i === tosses.length - 1,
            'is-next': !chart && i === tosses.length && !shaking
          }"
        >
          <div class="cell-head">
            <span>{{ GUAS[i] }}</span>
            <em>0{{ i + 1 }}</em>
          </div>
          <div class="coins">
            <span
              v-for="j in 3"
              :key="j"
              class="coin"
              :class="{ 'is-back': !!t && t.coins[j - 1] === 1 }"
            ></span>
          </div>
          <div class="cell-foot" :class="{ 'is-dim': !t }">
            <template v-if="t">背{{ t.backs }} · {{ t.kind }} · {{ t.moving ? '动' : '静' }}</template>
            <template v-else>待摇</template>
          </div>
        </div>
      </div>

      <div v-else-if="method === 'liuyao'" class="coin-strip is-manual">
        <div v-for="(_, i) in manualLines" :key="i" class="cast-cell">
          <div class="cell-head">
            <span>{{ GUAS[i] }}</span>
            <em>0{{ i + 1 }}</em>
          </div>
          <div class="select-wrap">
            <select v-model="manualLines[i]">
              <option v-for="opt in lineOptions" :key="opt" :value="opt">{{ opt }}</option>
            </select>
          </div>
          <div class="cell-foot">{{ manualLines[i] }} · {{ lineState(manualLines[i]) }}</div>
        </div>
      </div>

      <p v-if="error" class="error-line">{{ error }}</p>
    </section>

    <section ref="boardEl" class="board-panel">
      <template v-if="chart">
        <header v-if="liuyaoPayload" class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              本卦 · {{ liuyaoPayload.palace }} · {{ liuyaoPayload.rank }} · 宫五行 {{ liuyaoPayload.palace_wuxing }}
            </div>
            <h2>{{ liuyaoPayload.gua_name }}</h2>
          </div>
          <div v-if="liuyaoPayload.changed_gua" class="gua-change">
            <span>之</span>
            <strong>{{ liuyaoPayload.changed_gua }}</strong>
            <em>{{ liuyaoPayload.changed_palace }}</em>
          </div>
          <div v-else class="gua-change is-quiet">
            <span>六爻安静</span>
            <em>无变卦</em>
          </div>
        </header>

        <header v-else-if="daLiuRenPayload" class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              大六壬 · {{ daLiuRenPayload.month_general.zhi }}将 · {{ daLiuRenPayload.hour_zhi }}时
            </div>
            <h2>{{ daLiuRenPayload.ke_ti }}</h2>
          </div>
          <div class="gua-change">
            <span>三传</span>
            <strong>{{ daLiuRenPayload.san_chuan.map((c) => c.branch).join(' → ') }}</strong>
            <em>{{ daLiuRenPayload.gate }}课</em>
          </div>
        </header>

        <header v-else-if="ziWeiPayload" class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              紫微斗数 · {{ ziWeiPayload.lunar_date }}
            </div>
            <h2>{{ ziWeiPayload.ming_gong }}宫命</h2>
          </div>
          <div class="gua-change">
            <span>身宫</span>
            <strong>{{ ziWeiPayload.shen_gong }}宫</strong>
            <em>{{ ziWeiPayload.wu_xing_ju }} · 紫微在{{ ziWeiPayload.ziwei_pos }}</em>
          </div>
        </header>

        <header v-else-if="xiaoLiuRenPayload" class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              小六壬 · 农历{{ xiaoLiuRenPayload.lunar_leap ? '闰' : '' }}{{ xiaoLiuRenPayload.lunar_month }}月{{
                xiaoLiuRenPayload.lunar_day
              }}日 · {{ xiaoLiuRenPayload.hour_zhi }}时
            </div>
            <h2>{{ xiaoLiuRenPayload.final_palace.name }}</h2>
          </div>
          <div class="gua-change">
            <span>三宫</span>
            <strong>{{ xiaoLiuRenPayload.month_palace.name }} → {{ xiaoLiuRenPayload.day_palace.name }} → {{ xiaoLiuRenPayload.final_palace.name }}</strong>
            <em>{{ xiaoLiuRenPayload.final_palace.element }} · {{ xiaoLiuRenPayload.final_palace.deity }}</em>
          </div>
        </header>

        <header v-else-if="meiHuaPayload" class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              梅花易数 · {{ meiHuaPayload.mode === 'numbers' ? '报数' : '时间' }}起卦 · 第{{ meiHuaPayload.moving_line }}爻动
            </div>
            <h2>{{ meiHuaPayload.ben_gua.name }}</h2>
          </div>
          <div class="gua-change">
            <span>之</span>
            <strong>{{ meiHuaPayload.bian_gua.name }}</strong>
            <em>{{ meiHuaPayload.relation }}</em>
          </div>
        </header>

        <header v-else-if="qiMenPayload" class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              奇门遁甲 · {{ qiMenPayload.dun }}{{ qiMenPayload.ju }}局 · {{ qiMenPayload.solar_term }}{{ qiMenPayload.yuan }}
            </div>
            <h2>{{ qiMenPayload.zhi_fu }}值符</h2>
          </div>
          <div class="gua-change">
            <span>值使</span>
            <strong>{{ qiMenPayload.zhi_shi }}</strong>
            <em>符 {{ qiMenPayload.zhi_fu_palace }}宫 · 使 {{ qiMenPayload.zhi_shi_palace }}宫</em>
          </div>
        </header>

        <div class="board-meta fade-in">
          <template v-if="liuyaoPayload">
            <span>月建 <b>{{ liuyaoPayload.month_gz }}</b></span>
            <span>日辰 <b>{{ liuyaoPayload.day_gz }}</b></span>
            <span>旬空 <b>{{ liuyaoPayload.xun_kong.join('、') }}</b></span>
            <span>宫五行 <b>{{ liuyaoPayload.palace_wuxing }}</b></span>
          </template>
          <template v-else-if="daLiuRenPayload">
            <span>日 <b>{{ daLiuRenPayload.day_gan }}{{ daLiuRenPayload.day_zhi }}</b></span>
            <span>时 <b>{{ daLiuRenPayload.hour_zhi }}</b></span>
            <span>月将 <b>{{ daLiuRenPayload.month_general.zhi }}</b></span>
            <span>旬空 <b>{{ daLiuRenPayload.xun_kong.join('、') }}</b></span>
            <span>课体 <b>{{ daLiuRenPayload.gate }}</b></span>
          </template>
          <template v-else-if="ziWeiPayload">
            <span>命宫 <b>{{ ziWeiPayload.ming_gong }}</b></span>
            <span>身宫 <b>{{ ziWeiPayload.shen_gong }}</b></span>
            <span>五行局 <b>{{ ziWeiPayload.wu_xing_ju }}</b></span>
            <span>紫微 <b>{{ ziWeiPayload.ziwei_pos }}</b></span>
            <span>天府 <b>{{ ziWeiPayload.tianfu_pos }}</b></span>
          </template>
          <template v-else-if="xiaoLiuRenPayload">
            <span>月宫 <b>{{ xiaoLiuRenPayload.month_palace.name }}</b></span>
            <span>日宫 <b>{{ xiaoLiuRenPayload.day_palace.name }}</b></span>
            <span>落宫 <b>{{ xiaoLiuRenPayload.final_palace.name }}</b></span>
            <span>时支 <b>{{ xiaoLiuRenPayload.hour_zhi }}</b></span>
          </template>
          <template v-else-if="meiHuaPayload">
            <span>本卦 <b>{{ meiHuaPayload.ben_gua.name }}</b></span>
            <span>互卦 <b>{{ meiHuaPayload.hu_gua.name }}</b></span>
            <span>变卦 <b>{{ meiHuaPayload.bian_gua.name }}</b></span>
            <span>动爻 <b>第{{ meiHuaPayload.moving_line }}爻</b></span>
            <span>体用 <b>{{ meiHuaPayload.relation }}</b></span>
          </template>
          <template v-else-if="qiMenPayload">
            <span>局 <b>{{ qiMenPayload.dun }}{{ qiMenPayload.ju }}局</b></span>
            <span>节气 <b>{{ qiMenPayload.solar_term }}{{ qiMenPayload.yuan }}</b></span>
            <span>值符 <b>{{ qiMenPayload.zhi_fu }}</b></span>
            <span>值使 <b>{{ qiMenPayload.zhi_shi }}</b></span>
            <span>空亡 <b>{{ qiMenPayload.kong.join('、') }}</b></span>
            <span>马星 <b>{{ qiMenPayload.horse }}</b></span>
          </template>
          <span>引擎 <b>{{ chart.engine_version }}</b></span>
          <span v-if="liuyaoPayload" class="scroll-hint">盘面可左右滑动查看变卦</span>
          <span v-else-if="daLiuRenPayload" class="scroll-hint">天地盘可左右滑动查看</span>
          <span v-else-if="ziWeiPayload" class="scroll-hint">十二宫可左右滑动查看</span>
          <span v-else-if="qiMenPayload" class="scroll-hint">九宫可左右滑动查看</span>
        </div>

        <div class="board-scroll">
          <LiuYaoBoard v-if="liuyaoPayload" :payload="liuyaoPayload" :revealed="boardRevealed" />
          <DaLiuRenBoard v-else-if="daLiuRenPayload" :payload="daLiuRenPayload" />
          <ZiWeiBoard v-else-if="ziWeiPayload" :payload="ziWeiPayload" />
          <XiaoLiuRenBoard v-else-if="xiaoLiuRenPayload" :payload="xiaoLiuRenPayload" />
          <MeiHuaBoard v-else-if="meiHuaPayload" :payload="meiHuaPayload" />
          <QiMenBoard v-else-if="qiMenPayload" :payload="qiMenPayload" />
        </div>

        <div v-if="liuyaoPayload && (chart.entropy || mode === 'coin')" class="entropy">
          <b>本次摇卦记录</b>
          <span>背数 {{ chart.entropy ? chart.entropy.backs.join('、') : localBacks.join('、') }}</span>
          <span v-if="chart.entropy">校验 {{ chart.entropy.entropy_proof.slice(0, 16) }}…</span>
          <span v-else>随机源 本地 CSPRNG · 服务端复算</span>
          <span>{{ formatTime(chart.calendar?.solar_time || chart.created_at) }}</span>
        </div>
      </template>

      <div v-else-if="method === 'liuyao' && tosses.length" class="partial-board">
        <div class="partial-seal">
          <div class="partial-count"><b>{{ tosses.length }}</b><em>/ 6</em></div>
          <span>已摇</span>
        </div>
        <div class="partial-main">
          <div class="partial-head">
            <strong>{{ submitting ? '成卦中' : '卦未成' }}</strong>
            <span v-if="submitting">六爻已齐，正在装卦</span>
            <span v-else-if="tosses.length >= 6">六爻已齐 · 点右上按钮重试</span>
            <span v-else>再摇 {{ 6 - tosses.length }} 次 · 自下而上</span>
          </div>
          <div class="partial-lines">
            <div v-for="idx in ORDER" :key="idx" class="pline-row">
              <span class="pline-label">{{ GUAS[idx] }}</span>
              <div
                v-if="idx < tosses.length"
                class="pline-mark"
                :class="{
                  'is-yin': !isYang(idx),
                  'is-moving': strip[idx]?.moving,
                  'is-new': idx === tosses.length - 1
                }"
              >
                <i></i><i v-if="!isYang(idx)"></i>
              </div>
              <div v-else class="pline-wait"></div>
              <span class="pline-foot" :class="{ 'is-dim': idx >= tosses.length }">
                <template v-if="idx < tosses.length">
                  背{{ strip[idx]?.backs }} · {{ strip[idx]?.kind }} · {{ strip[idx]?.moving ? '动' : '静' }}
                </template>
                <template v-else>待摇</template>
              </span>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="empty-board">
        <div class="empty-seal">{{ emptyCopy.seal }}</div>
        <div class="empty-copy">
          <p class="empty-hint">{{ emptyCopy.hint }}</p>
          <ol class="empty-steps">
            <li v-for="(step, index) in emptyCopy.steps" :key="step[0]">
              <i>{{ CN_NUMS[index] }}</i>
              <div>
                <b>{{ step[0] }}</b>
                <em>{{ step[1] }}</em>
              </div>
            </li>
          </ol>
        </div>
      </div>
    </section>

    <section v-if="chart" class="reading-panel">
      <header class="reading-head">
        <span class="seal">断</span>
        <h2>解读</h2>
        <em>{{ readingStatusText }}</em>
        <button
          class="reading-refresh"
          :disabled="readingBusy || followupBusy"
          @click="startReading"
        >
          {{ reading.status === 'error' ? '重试' : '重新解读' }}
        </button>
      </header>

      <div class="reading-ai-overview">
        <span class="ai-label">盘面要点</span>
        <p v-if="readingSections.overview" class="reading-ai-lede">{{ readingSections.overview }}</p>
        <p v-else-if="readingBusy" class="reading-ai-lede is-placeholder">正在把盘面翻译成人话…</p>
        <p v-else-if="reading.status === 'error'" class="reading-ai-lede is-placeholder">
          解读失败，盘面仍可查看。
        </p>
        <p v-else class="reading-ai-lede is-placeholder">暂无解读。</p>
      </div>

      <div class="facts-section">
        <h3><span class="num">一</span>关键事实</h3>
        <dl class="facts-grid">
          <div v-for="f in coreFacts" :key="f.key" class="fact">
            <dt>{{ f.label }}</dt>
            <dd>{{ f.value }}</dd>
          </div>
        </dl>
      </div>

      <div class="reading-columns">
        <article class="reading-block">
          <h3><span class="num">二</span>断语</h3>
          <p v-if="readingSections.verdict" class="reading-ai-text">{{ readingSections.verdict }}</p>
          <p v-else-if="readingBusy" class="reading-note">正在解读…</p>
          <div v-else-if="reading.status === 'error'" class="reading-error">
            <p>{{ reading.error }}</p>
            <button class="reading-retry" @click="startReading">重试</button>
          </div>
          <p v-else class="reading-note">暂无断语。</p>
        </article>
        <article class="reading-block">
          <h3><span class="num">三</span>{{ method === 'ziwei' ? '大限与建议' : '应期与建议' }}</h3>
          <template v-if="readingSections.timing || readingSections.advice">
            <div v-if="readingSections.timing" class="ai-field">
              <label>应期与走向</label>
              <p class="reading-ai-text">{{ readingSections.timing }}</p>
            </div>
            <div v-if="readingSections.advice" class="ai-field">
              <label>建议</label>
              <p class="reading-ai-text">{{ readingSections.advice }}</p>
            </div>
          </template>
          <p v-else-if="readingBusy" class="reading-note">正在解读…</p>
          <p v-else-if="reading.status === 'error'" class="reading-note">解读失败后暂无{{ method === 'ziwei' ? '大限与建议' : '应期与建议' }}。</p>
          <p v-else class="reading-note">暂无{{ method === 'ziwei' ? '大限与建议' : '应期与建议' }}。</p>
        </article>
      </div>

      <p class="reading-disclaimer">
        <template v-if="reading.status === 'error'">
          解读未完成，盘面事实仍以引擎计算为准。可稍后重试，或检查服务端模型配置。
        </template>
        <template v-else-if="reading.mode === 'ai'">
          以上解读由 {{ reading.model }} 生成，并已通过盘面事实校验；仅供传统文化学习与娱乐参考，盘面事实以引擎计算为准。
        </template>
        <template v-else-if="reading.mode === 'offline'">
          以上为离线盘面摘要，只列事实，不含吉凶判断。配置模型后可获得完整四段解读。
        </template>
        <template v-else>解读生成中；盘面事实以引擎计算为准。</template>
      </p>
    </section>

    <section v-if="chart" class="followup-panel">
      <header class="followup-head">
        <span class="seal">问</span>
        <h2>同盘追问</h2>
        <em v-if="followupReady">沿这张{{ chartUnit }}继续，不另起{{ chartUnit }}</em>
        <em v-else-if="followupUnlocked">已追问 {{ MAX_FOLLOWUPS }} 轮</em>
        <em v-else>{{ followupLockedText }}</em>
      </header>

      <template v-if="followupUnlocked">
        <div v-if="followups.length" ref="followupThreadEl" class="followup-thread">
          <article
            v-for="turn in followups"
            :key="turn.id"
            class="followup-turn"
            :class="[`is-${turn.role}`, { 'is-error': turn.status === 'error' }]"
          >
            <header>
              <strong>{{ turn.role === 'user' ? '你问' : turn.model || '命云' }}</strong>
              <time>{{ formatTime(turn.created_at) }}</time>
            </header>
            <p v-if="turn.content" class="followup-text">{{ turn.content }}</p>
            <p v-else-if="turn.status === 'streaming'" class="followup-placeholder">
              正在沿着这张盘回答…
            </p>
            <p v-if="turn.error" class="followup-turn-error">{{ turn.error }}</p>
          </article>
        </div>
        <p v-else class="followup-empty">
          首轮四段已经在上面。可以接着问断语为什么落在那个节点、某条判断对应盘面的哪一处，或问清一个具体细节。
        </p>

        <div v-if="followupReady" class="followup-compose">
          <textarea
            v-model="followupInput"
            rows="2"
            maxlength="200"
            :disabled="followupBusy"
            placeholder="接着问这张盘，例如：应期为什么会拖到出空之后？"
            @keydown="onFollowupKeydown"
          ></textarea>
          <div class="followup-compose-foot">
            <em>已追问 {{ followupUsed }} / {{ MAX_FOLLOWUPS }} 轮 · 同一张盘</em>
            <button
              type="button"
              class="followup-send"
              :disabled="followupBusy || !followupInput.trim()"
              @click="sendFollowup"
            >
              {{ followupBusy ? '回答中…' : '追问' }}
            </button>
          </div>
        </div>
        <p v-else class="followup-locked">
          这张盘已追问 {{ MAX_FOLLOWUPS }} 轮，建议另起一{{ chartUnit }}；追问记录已留在卦历。
        </p>
        <p v-if="followupError" class="followup-error-line">{{ followupError }}</p>
      </template>
    </section>

    <section
      v-if="currentArchiveEntry"
      class="case-seal-panel"
      aria-label="历史记录已收录"
    >
      <div class="case-seal-stamp">
        <b>录</b>
        <span>已藏</span>
      </div>
      <div class="case-seal-copy">
        <span class="case-seal-kicker">{{ methodMeta.volume }} · {{ methodLabel }}</span>
        <strong>此{{ chartUnit }}已收入卦历</strong>
        <p>
          {{ chart?.payload.gua_name }} · {{ formatTime(currentArchiveEntry.saved_at) }} ·
          第 {{ currentCaseOrdinal }} 则
        </p>
      </div>
      <div class="case-seal-actions">
        <button class="seal-btn seal-btn--small" @click="startAnother">再问一事</button>
        <button class="text-btn" @click="navigate('archive')">翻看历史</button>
      </div>
    </section>
        </div>
      </section>

      <GameArchive
        v-else-if="view === 'archive'"
        key="archive"
        :entries="archive"
        @open="openArchive"
        @cast="startMethod('liuyao')"
        @home="navigate('home')"
        @delete="deleteArchiveEntry"
        @clear="clearArchive"
        @export="exportArchive"
      />

      <GameCodex
        v-else
        key="codex"
        @cast="startMethod"
        @home="navigate('home')"
      />

    <footer class="footnote game-foot">
      <span>命云 MingYun · MIT 开源 · 盘面可复核</span>
      <span>仅供传统文化学习、研究与娱乐，不构成任何专业建议。</span>
    </footer>

    <GameNav :current="view" :archive-count="archive.length" @navigate="navigate" />
  </div>
</template>
