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

export interface SurnameHint {
  position: number
  zhi: string
  animal: string
  surname?: string
  is_shi?: boolean
  is_ying?: boolean
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
  surname_hints?: SurnameHint[]
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

export interface DaLiuRenTianPan {
  earth: string
  heaven: string
  heaven_wuxing: string
  liu_qin: string
  tian_jiang: string
  shen_sha?: string[]
}

export interface DaLiuRenSiKe {
  index: number
  name: string
  lower: string
  upper: string
  lower_wuxing: string
  upper_wuxing: string
  relation: string
}

export interface DaLiuRenChuan {
  index: number
  name: string
  branch: string
  palace: string
  tian_jiang: string
  liu_qin: string
}

export interface DaLiuRenShenSha {
  name: string
  zhi: string
  value: string
}

export interface DaLiuRenPayload {
  question?: string
  month_general: MonthGeneral
  hour_zhi: string
  day_gan: string
  day_zhi: string
  month_zhi: string
  xun_kong: [string, string]
  tian_pan: DaLiuRenTianPan[]
  si_ke: DaLiuRenSiKe[]
  san_chuan: DaLiuRenChuan[]
  ke_ti: string
  gate: string
  shen_sha?: DaLiuRenShenSha[]
  school?: Record<string, string>
}

export interface ZiWeiStar {
  name: string
  kind: string
  hua?: string
}

export interface ZiWeiSihua {
  gan: string
  star: string
  hua: string
}

export interface ZiWeiDaXian {
  index: number
  name: string
  branch: string
  start_age: number
  end_age: number
  range: string
}

export interface ZiWeiPalace {
  index: number
  branch: string
  stem: string
  stem_branch: string
  name: string
  is_ming?: boolean
  is_shen?: boolean
  major_stars?: ZiWeiStar[]
  minor_stars?: ZiWeiStar[]
  sihua?: string[]
  daxian_start?: number
  daxian_end?: number
  daxian_range?: string
}

export interface ZiWeiPayload {
  question?: string
  gender: string
  calendar_type: string
  birth_date: string
  birth_time: string
  is_leap_month: boolean
  leap_month_policy: string
  lunar_date: string
  solar_date: string
  year_gz: string
  ming_gong: string
  shen_gong: string
  wu_xing_ju: string
  ziwei_pos: string
  tianfu_pos: string
  palaces: ZiWeiPalace[]
  da_xian: ZiWeiDaXian[]
  sihua_year: ZiWeiSihua[]
  school?: Record<string, string>
}

export type ChartPayload = LiuYaoPayload | DaLiuRenPayload | ZiWeiPayload

export function isLiuYaoPayload(payload: ChartPayload): payload is LiuYaoPayload {
  return 'gua_name' in payload && 'lines' in payload
}

export function isDaLiuRenPayload(payload: ChartPayload): payload is DaLiuRenPayload {
  return 'tian_pan' in payload && 'san_chuan' in payload
}

export function isZiWeiPayload(payload: ChartPayload): payload is ZiWeiPayload {
  return 'palaces' in payload && 'ming_gong' in payload
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
  payload: ChartPayload
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

export type ArchiveStatus = 'pending' | 'reading' | 'done' | 'error'

export type FollowupStatus = 'streaming' | 'done' | 'error'

/** 同一张盘内的一条追问对话，role 只允许 user / assistant。 */
export interface FollowupMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
  model?: string
  status?: FollowupStatus
  error?: string
}

/**
 * 卦历记录只保存在浏览器本机，不上传服务端。
 * chart 存完整盘面，便于从「卦历」直接翻回原盘。
 */
export interface ArchiveEntry {
  chart_id: string
  question: string
  /** 旧六爻记录字段；三术统一展示优先用 title / subtitle。 */
  gua_name?: string
  changed_gua?: string
  title?: string
  subtitle?: string
  method?: string
  saved_at: string
  cast_at?: string
  read_at?: string
  status?: ArchiveStatus
  topic_label?: string
  mode: 'ai' | 'offline' | ''
  model: string
  reading: string
  reading_error?: string
  followups?: FollowupMessage[]
  chart: ChartView
}
