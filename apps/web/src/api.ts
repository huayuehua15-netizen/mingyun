import type { CalendarContext, ChartView, Envelope } from './types'

/** 默认经度：中山，东经 113.39°。真太阳时与历法共用。 */
export const DEFAULT_LON = 113.39

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  let body: Envelope<T>
  try {
    body = (await res.json()) as Envelope<T>
  } catch {
    throw new Error(`服务返回了非 JSON 响应（HTTP ${res.status}）`)
  }
  if (!res.ok || body.code !== 0) {
    throw new Error(body.message || `请求失败（HTTP ${res.status}）`)
  }
  return body.data
}

export function fetchCalendar(at?: string): Promise<CalendarContext> {
  const params = new URLSearchParams({ lon: String(DEFAULT_LON) })
  if (at) params.set('at', at)
  return request<CalendarContext>(`/api/v1/meta/calendar?${params.toString()}`)
}

export interface CastRequest {
  mode: 'coin' | 'manual'
  question?: string
  at?: string
  lon?: number
  tz?: string
  use_true_solar_time?: boolean
  day_boundary?: 'zi_chu' | 'zi_zheng'
  lines?: string[]
  coins?: number[][]
}

export function castLiuYao(req: CastRequest): Promise<ChartView> {
  return request<ChartView>('/api/v1/charts/liuyao/cast', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      lon: DEFAULT_LON,
      use_true_solar_time: true,
      ...req
    })
  })
}
