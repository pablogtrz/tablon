import { afterEach, describe, expect, it, vi } from 'vitest'
import { newIdempotencyKey } from './idempotency'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('newIdempotencyKey', () => {
  it('returns a UUID in a secure context', () => {
    const key = newIdempotencyKey()
    expect(key).toMatch(/^[0-9a-f-]{36}$/)
  })

  it('falls back to random bytes without crypto.randomUUID', () => {
    vi.stubGlobal('crypto', {
      getRandomValues: (bytes: Uint8Array) => {
        bytes.fill(1)
        return bytes
      },
    })
    expect(newIdempotencyKey()).toBe('01'.repeat(16))
  })
})
