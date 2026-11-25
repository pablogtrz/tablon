import { getToken } from './api'

export interface ProjectEvent {
  type: string
  project_id: string
  issue_id?: string
  identifier?: string
  comment_id?: string
  updated_at?: string
}

export interface SSEFrame {
  event: string
  data: string
}

export interface ParsedSSE {
  frames: SSEFrame[]
  rest: string
}

export function parseSSEBuffer(input: string): ParsedSSE {
  const buffer = input.replace(/\r\n/g, '\n')
  const frames: SSEFrame[] = []
  let rest = buffer
  for (;;) {
    const index = rest.indexOf('\n\n')
    if (index < 0) break
    const block = rest.slice(0, index)
    rest = rest.slice(index + 2)
    const frame = parseFrame(block)
    if (frame) frames.push(frame)
  }
  return { frames, rest }
}

function parseFrame(block: string): SSEFrame | null {
  let event = 'message'
  const data: string[] = []
  for (const line of block.split('\n')) {
    if (!line || line.startsWith(':')) continue
    const colon = line.indexOf(':')
    const field = colon < 0 ? line : line.slice(0, colon)
    let value = colon < 0 ? '' : line.slice(colon + 1)
    if (value.startsWith(' ')) value = value.slice(1)
    if (field === 'event') {
      event = value
    } else if (field === 'data') {
      data.push(value)
    }
  }
  if (data.length === 0) return null
  return { event, data: data.join('\n') }
}

export function reconnectDelay(attempt: number, random = Math.random): number {
  const exponential = Math.min(30_000, 1_000 * 2 ** Math.min(attempt, 5))
  return exponential + Math.floor(random() * 250)
}

export type ProjectEventStatus = 'connecting' | 'live' | 'offline' | 'unauthorized'

export interface ProjectEventHandlers {
  onEvent: (event: ProjectEvent) => void
  onsync?: () => void
  onstatus?: (status: ProjectEventStatus) => void
}

export function subscribeProjectEvents(
  projectRef: string,
  handlers: ProjectEventHandlers,
): () => void {
  let stopped = false
  let hidden = typeof document !== 'undefined' && document.hidden
  let controller: AbortController | null = null
  let timer: ReturnType<typeof setTimeout> | null = null
  let attempt = 0

  const connect = async () => {
    if (stopped || hidden || controller) return
    const ac = new AbortController()
    controller = ac
    handlers.onstatus?.('connecting')
    const headers = new Headers({ Accept: 'text/event-stream' })
    const token = getToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)

    try {
      const response = await fetch(
        `/api/v1/projects/${encodeURIComponent(projectRef)}/events`,
        { headers, signal: ac.signal, cache: 'no-store' },
      )
      if (response.status === 401 || response.status === 403) {
        stopped = true
        handlers.onstatus?.('unauthorized')
        return
      }
      if (!response.ok || !response.body) {
        throw new Error(`event stream failed with ${response.status}`)
      }

      attempt = 0
      handlers.onstatus?.('live')
      handlers.onsync?.()

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const parsed = parseSSEBuffer(buffer)
        buffer = parsed.rest
        for (const frame of parsed.frames) {
          if (frame.event === 'ping') continue
          try {
            const event = JSON.parse(frame.data) as ProjectEvent
            event.type ||= frame.event
            handlers.onEvent(event)
          } catch {
            // Ignore malformed events; the next sync refetches state.
          }
        }
      }
      throw new Error('event stream ended')
    } catch {
      if (stopped || ac.signal.aborted) return
      handlers.onstatus?.('offline')
      timer = setTimeout(() => {
        timer = null
        void connect()
      }, reconnectDelay(attempt))
      attempt += 1
    } finally {
      if (controller === ac) controller = null
    }
  }

  const onVisibility = () => {
    if (document.hidden) {
      hidden = true
      controller?.abort()
      controller = null
      if (timer) clearTimeout(timer)
      timer = null
      return
    }
    hidden = false
    attempt = 0
    void connect()
  }

  document.addEventListener('visibilitychange', onVisibility)
  void connect()

  return () => {
    stopped = true
    controller?.abort()
    controller = null
    if (timer) clearTimeout(timer)
    timer = null
    document.removeEventListener('visibilitychange', onVisibility)
  }
}
