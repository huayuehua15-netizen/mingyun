<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { castLiuYao, fetchCalendar, streamInterpret } from './api'
import type { InterpretEvent } from './api'
import type { CalendarContext, ChartView } from './types'
import LiuYaoBoard from './components/LiuYaoBoard.vue'

const question = ref('')
const mode = ref<'coin' | 'manual'>('coin')
const manualLines = ref<string[]>(['少阳', '少阳', '少阳', '少阳', '少阳', '少阳'])
const lineOptions = ['少阳', '少阴', '老阳', '老阴']

const chart = ref<ChartView | null>(null)
const calendar = ref<CalendarContext | null>(null)
const error = ref('')

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
const needsQuestion = computed(() => !chart.value && !questionReady.value)

const castLabel = computed(() => {
  if (submitting.value) return '成卦中'
  if (shaking.value) return '摇卦中'
  if (chart.value) return '再 起 卦'
  if (!questionReady.value) return '先 问 一 事'
  if (mode.value === 'manual') return '起 卦'
  if (tosses.value.length >= 6) return '重试成卦'
  return `摇 ${GUAS[tosses.value.length]}`
})

const castDisabled = computed(() => shaking.value || submitting.value)

const almanacTerms = computed(() => {
  if (!calendar.value) return []
  const c = calendar.value
  return [
    { k: '节气', v: `${c.prev_term.name}后` },
    { k: '月将', v: `${c.month_general.zhi}将${c.month_general.name}` },
    { k: '旬空', v: c.xun_kong.join('、') },
    { k: '时辰', v: `${c.hour_gz}时` }
  ]
})

onMounted(async () => {
  const params = new URLSearchParams(window.location.search)
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

  if (params.get('cast') === '1') await autoCast()
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

function setMode(next: 'coin' | 'manual') {
  if (shaking.value || submitting.value || mode.value === next) return
  resetReading()
  mode.value = next
  chart.value = null
  tosses.value = []
  error.value = ''
}

function reset() {
  resetReading()
  question.value = ''
  chart.value = null
  tosses.value = []
  error.value = ''
}

function resetReading() {
  readingAbort?.abort()
  readingAbort = null
  readingSeq += 1
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
  } catch (e) {
    if (ctrl.signal.aborted || seq !== readingSeq) return
    reading.value.status = 'error'
    reading.value.error = e instanceof Error ? e.message : '解读失败'
  } finally {
    if (readingAbort === ctrl) readingAbort = null
  }
}

function lineState(v: string): string {
  return v.startsWith('老') ? '动' : '静'
}

function scrollToBoard() {
  if (previewMode.value) return
  boardEl.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function onCast() {
  if (shaking.value || submitting.value) return
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

function formatTime(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleString('zh-CN', { hour12: false })
}
</script>

<template>
  <div class="page">
    <header class="topbar">
      <a class="brand" href="/">
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
      </a>
      <nav class="nav">
        <a class="is-active" href="#cast">六爻</a>
        <span class="is-soon">大六壬<sup>筹备</sup></span>
        <span class="is-soon">紫微斗数<sup>筹备</sup></span>
        <i class="nav-div"></i>
        <a href="https://github.com/huayuehua15-netizen/mingyun" target="_blank" rel="noopener">GitHub ↗</a>
      </nav>
    </header>

    <section class="masthead">
      <div class="masthead-copy">
        <div class="kicker"><i></i>AI 术数排盘 · 开源</div>
        <h1>问一事，起一卦</h1>
        <p>铜钱六摇，装卦断事。盘面由确定性引擎计算，解读交由可自主配置的大模型——它能编故事，但改不了卦。</p>
      </div>
      <aside v-if="calendar" class="almanac">
        <div class="almanac-date">{{ calendar.year_gz }}年 {{ calendar.month_gz }}月 {{ calendar.day_gz }}日</div>
        <div class="almanac-list">
          <div v-for="t in almanacTerms" :key="t.k">
            <label>{{ t.k }}</label>
            <span>{{ t.v }}</span>
          </div>
        </div>
        <div class="almanac-foot">
          真太阳时{{ calendar.true_solar_used ? '已校正' : '未校正' }} · 东经 113.39°
        </div>
        <div class="almanac-seal">问</div>
      </aside>
    </section>

    <section id="cast" class="cast-panel">
      <div class="cast-row">
        <div class="ask">
          <label for="question">所问之事 <span>QUESTION</span></label>
          <input
            id="question"
            ref="questionEl"
            v-model="question"
            maxlength="200"
            placeholder="例如：这次考试能否通过？"
            @keydown.enter.prevent="onCast"
          />
        </div>
        <div class="mode-switch">
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

      <div v-if="mode === 'coin'" class="coin-strip">
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

      <div v-else class="coin-strip is-manual">
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
        <header class="board-head fade-in">
          <div class="gua-block">
            <div class="gua-kicker">
              本卦 · {{ chart.payload.palace }} · {{ chart.payload.rank }} · 宫五行 {{ chart.payload.palace_wuxing }}
            </div>
            <h2>{{ chart.payload.gua_name }}</h2>
          </div>
          <div v-if="chart.payload.changed_gua" class="gua-change">
            <span>之</span>
            <strong>{{ chart.payload.changed_gua }}</strong>
            <em>{{ chart.payload.changed_palace }}</em>
          </div>
          <div v-else class="gua-change is-quiet">
            <span>六爻安静</span>
            <em>无变卦</em>
          </div>
        </header>

        <div class="board-meta fade-in">
          <span>月建 <b>{{ chart.payload.month_gz }}</b></span>
          <span>日辰 <b>{{ chart.payload.day_gz }}</b></span>
          <span>旬空 <b>{{ chart.payload.xun_kong.join('、') }}</b></span>
          <span>宫五行 <b>{{ chart.payload.palace_wuxing }}</b></span>
          <span>引擎 <b>{{ chart.engine_version }}</b></span>
          <span class="scroll-hint">盘面可左右滑动查看变卦</span>
        </div>

        <div class="board-scroll">
          <LiuYaoBoard :payload="chart.payload" :revealed="boardRevealed" />
        </div>

        <div v-if="chart.entropy || mode === 'coin'" class="entropy">
          <b>本次摇卦记录</b>
          <span>背数 {{ chart.entropy ? chart.entropy.backs.join('、') : localBacks.join('、') }}</span>
          <span v-if="chart.entropy">校验 {{ chart.entropy.entropy_proof.slice(0, 16) }}…</span>
          <span v-else>随机源 本地 CSPRNG · 服务端复算</span>
          <span>{{ formatTime(chart.calendar?.solar_time || chart.created_at) }}</span>
        </div>
      </template>

      <div v-else-if="tosses.length" class="partial-board">
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
        <div class="empty-seal">未占</div>
        <div class="empty-copy">
          <p class="empty-hint">心诚则灵 · 先问一事</p>
          <ol class="empty-steps">
            <li>
              <i>一</i>
              <div><b>问事</b><em>问题越具体，盘越好读</em></div>
            </li>
            <li>
              <i>二</i>
              <div><b>摇卦</b><em>六次铜钱，自下而上</em></div>
            </li>
            <li>
              <i>三</i>
              <div><b>断卦</b><em>盘面即事实，AI 只负责说人话</em></div>
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
          :disabled="readingBusy"
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
          <h3><span class="num">三</span>应期与建议</h3>
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
          <p v-else-if="reading.status === 'error'" class="reading-note">解读失败后暂无应期与建议。</p>
          <p v-else class="reading-note">暂无应期与建议。</p>
        </article>
      </div>

      <p class="reading-disclaimer">
        <template v-if="reading.status === 'error'">
          解读未完成，盘面事实仍以引擎计算为准。可稍后重试，或检查服务端模型配置。
        </template>
        <template v-else-if="reading.mode === 'ai'">
          以上解读由 {{ reading.model }} 生成，仅供传统文化学习与娱乐参考；盘面事实以引擎计算为准。
        </template>
        <template v-else-if="reading.mode === 'offline'">
          以上为离线盘面摘要，只列事实，不含吉凶判断。配置模型后可获得完整四段解读。
        </template>
        <template v-else>解读生成中；盘面事实以引擎计算为准。</template>
      </p>
    </section>

    <footer class="footnote">
      <span>命云 MingYun · MIT 开源 · 盘面可复核</span>
      <span>仅供传统文化学习、研究与娱乐，不构成任何专业建议。</span>
    </footer>
  </div>
</template>
