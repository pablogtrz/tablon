import { describe, expect, it } from 'vitest'
import { parseSSEBuffer, reconnectDelay } from './events'

describe('parseSSEBuffer', () => {
  it('parses events and keeps an incomplete tail', () => {
    const parsed = parseSSEBuffer(
      'retry: 3000\n\nevent: issue.updated\ndata: {"type":"issue.updated","project_id":"p"}\n\npartial',
    )
    expect(parsed.frames).toHaveLength(1)
    expect(parsed.frames[0].event).toBe('issue.updated')
    expect(JSON.parse(parsed.frames[0].data).type).toBe('issue.updated')
    expect(parsed.rest).toBe('partial')
  })

  it('ignores comments and heartbeats', () => {
    const parsed = parseSSEBuffer(': ping\n\nevent: message\ndata: ok\n\n')
    expect(parsed.frames).toHaveLength(1)
    expect(parsed.frames[0].data).toBe('ok')
  })

  it('joins multiline data and handles CRLF', () => {
    const parsed = parseSSEBuffer('event: x\r\ndata: one\r\ndata: two\r\n\r\n')
    expect(parsed.frames[0].data).toBe('one\ntwo')
  })
})

describe('reconnectDelay', () => {
  it('backs off exponentially and caps at 30s', () => {
    expect(reconnectDelay(0, () => 0)).toBe(1_000)
    expect(reconnectDelay(1, () => 0)).toBe(2_000)
    expect(reconnectDelay(2, () => 0)).toBe(4_000)
    expect(reconnectDelay(10, () => 0)).toBe(30_000)
  })

  it('adds bounded jitter', () => {
    expect(reconnectDelay(0, () => 1)).toBe(1_250)
  })
})
