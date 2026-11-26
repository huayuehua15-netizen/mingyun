export interface Term {
  name: string
  at: string
  is_jie: boolean
}

export interface MonthGeneral {
  zhi: string
  name: string
  after_term: string
}

export interface CalendarContext {
  solar_time: string
  true_solar_time: string
  lunar_date: string
  year_gz: string
  month_gz: string
  day_gz: string
  hour_gz: string
  xun_kong: [string, string]
  prev_term: Term
  next_term: Term
  month_general: MonthGeneral
  day_boundary: string
  true_solar_used: boolean
  warnings?: string[]
}

export type FactWeight = 'core' | 'normal' | 'minor'

export interface Fact {
  key: string
  label: string
  value: string
  anchor?: string
  weight: FactWeight
}

export interface ChangedLine {
  kind: string
  gan: string
  zhi: string
  wu_xing: string
  liu_qin: string
  relation?: string
}

export interface LinePayload {
  position: number
  kind: string
  yin_yang: '阴' | '阳'
  moving: boolean
  gan: string
  zhi: string
  wu_xing: string
  liu_qin: string
  liu_shen: string
  shi?: boolean
  ying?: boolean
  marks?: string[]
  change?: ChangedLine
}

export interface FuShen {
  under_position: number
  liu_qin: string
  gan: string
  zhi: string
  wu_xing: string
}

export interface UseShenCandidate {
  name: string
  positions?: number[]
  fu_positions?: number[]
}

export interface UseShenPlan {
  topic: string
  topic_label: string
  rule_id: string
  confidence: 'definite' | 'needs_gender' | 'open' | 'fuzzy' | string
  primary?: string[]
  secondary?: string[]
  taboo?: string[]
  rationale: string
  candidates?: UseShenCandidate[]
}

export interface LiuYaoPayload {
  question?: string
  gua_name: string
  upper: string
  lower: string
  palace: string
  palace_wuxing: string
  rank: string
  shi_yao: number
  ying_yao: number
  lines: LinePayload[]
  changed_gua?: string
  changed_palace?: string
  fu_shen?: FuShen[]
  month_gz: string
  day_gz: string
  month_zhi: string
  day_zhi: string
  xun_kong: [string, string]
  use_shen?: UseShenPlan
}

export interface Entropy {
  entropy_id: string
  backs: number[]
  entropy_proof: string
  source: string
  created_at: string
}

export interface ChartView {
  chart_id: string
  method: string
  engine_version: string
  question?: string
  calendar: CalendarContext
  school?: Record<string, string>
  payload: LiuYaoPayload
  facts: Fact[]
  warnings?: string[]
  entropy?: Entropy
  created_at: string
}

export interface Envelope<T> {
  code: number
  message: string
  data: T
  request_id: string
  server_time: string
}

export type ViewName = 'home' | 'cast' | 'archive' | 'codex'

/**
 * 卦例存档只保存在浏览器本机，不上传服务端。
 * chart 存完整盘面，便于从「卦例」直接翻回原盘。
 */
export interface ArchiveEntry {
  chart_id: string
  question: string
  gua_name: string
  changed_gua?: string
  saved_at: string
  mode: 'ai' | 'offline' | ''
  model: string
  reading: string
  chart: ChartView
}
