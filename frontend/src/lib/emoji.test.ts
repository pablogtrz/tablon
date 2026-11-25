import { describe, expect, it } from 'vitest'
import { firstGrapheme } from './emoji'

describe('firstGrapheme', () => {
  it('keeps one complete multi-codepoint emoji', () => {
    expect(firstGrapheme('👨‍👩‍👧‍👦')).toBe('👨‍👩‍👧‍👦')
    expect(firstGrapheme('🏴‍☠️')).toBe('🏴‍☠️')
    expect(firstGrapheme('🇪🇸')).toBe('🇪🇸')
    expect(firstGrapheme('❤️')).toBe('❤️')
  })

  it('drops later graphemes and handles empty input', () => {
    expect(firstGrapheme('🚀✨')).toBe('🚀')
    expect(firstGrapheme('  ')).toBe('')
  })
})
