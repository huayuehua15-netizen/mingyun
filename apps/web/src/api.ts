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
  topic?: string
  gender?: string
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

export type InterpretMode = 'ai' | 'offline'

export interface InterpretMeta {
  message_id: string
  mode: InterpretMode
  model: string
  provider: string
  chart_id: string
}

export interface InterpretUsage {
  prompt_tokens: number
  completion_tokens: number
  total_tokens?: number
}

export type InterpretEvent =
  | { event: 'meta'; data: InterpretMeta }
  | { event: 'delta'; data: { text: string } }
  | { event: 'usage'; data: InterpretUsage }
  | { event: 'done'; data: { message_id: string; finish_reason: string } }
  | { event: 'error'; data: { code: number; message: string } }

/**
 * 消费解读接口的 SSE 流。
 *
 * EventSource 不支持 POST，这里用 fetch + ReadableStream 手工解析
 * `event:` / `data:` 帧。服务端约定 meta / delta / usage / done / error。
 */
export async function streamInterpret(
  chartId: string,
  question: string,
  onEvent: (ev: InterpretEvent) => void,
  signal?: AbortSignal
): Promise<void> {
  const res = await fetch(`/api/v1/charts/${encodeURIComponent(chartId)}/interpret`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'text/event-stream'
    },
    body: JSON.stringify({ question, style: 'plain' }),
    signal
  })

  if (!res.ok) {
    let message = `解读请求失败（HTTP ${res.status}）`
    try {
      const body = (await res.json()) as Envelope<unknown>
      if (body?.message) message = body.message
    } catch {
      // 非 JSON 错误体，保留状态码提示。
    }
    throw new Error(message)
  }
  if (!res.body) throw new Error('解读接口没有返回流')

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  const dispatch = (event: string, dataLines: string[]) => {
    if (!dataLines.length) return
    const raw = dataLines.join('\n')
    let data: unknown
    try {
      data = JSON.parse(raw)
    } catch {
      return
    }
    onEvent({ event, data } as InterpretEvent)
  }

  const consume = (text: string) => {
    buffer += text.replace(/\r\n/g, '\n')
    let sep = buffer.indexOf('\n\n')
    while (sep >= 0) {
      const frame = buffer.slice(0, sep)
      buffer = buffer.slice(sep + 2)
      let event = 'message'
      const dataLines: string[] = []
      for (const line of frame.split('\n')) {
        if (!line || line.startsWith(':')) continue
        if (line.startsWith('event:')) event = line.slice(6).trim()
        else if (line.startsWith('data:')) dataLines.push(line.slice(5).replace(/^ /, ''))
      }
      dispatch(event, dataLines)
      sep = buffer.indexOf('\n\n')
    }
  }

  try {
    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      consume(decoder.decode(value, { stream: true }))
    }
    consume(decoder.decode())
    if (buffer.trim()) {
      // 流被直接关闭时，尽量把最后一帧也读出来。
      const frame = buffer
      let event = 'message'
      const dataLines: string[] = []
      for (const line of frame.split('\n')) {
        if (!line || line.startsWith(':')) continue
        if (line.startsWith('event:')) event = line.slice(6).trim()
        else if (line.startsWith('data:')) dataLines.push(line.slice(5).replace(/^ /, ''))
      }
      dispatch(event, dataLines)
    }
  } finally {
    reader.releaseLock()
  }
}
