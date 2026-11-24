import { describe, expect, it } from 'vitest'
import {
  buildCodeFence,
  normalizeLanguage,
  prefixLines,
  wrapSelection,
} from './markdown-editor'
import { scanMarkdownLines } from './editor-theme'

describe('markdown editor actions', () => {
  it('wraps an empty selection with a placeholder', () => {
    const result = wrapSelection('', 0, 0, '**', '**', 'bold text')
    expect(result.value).toBe('**bold text**')
    expect(result.value.slice(result.from, result.to)).toBe('bold text')
  })

  it('wraps selected text and keeps it selected', () => {
    const result = wrapSelection('hello', 0, 5, '*', '*', 'italic')
    expect(result.value).toBe('*hello*')
    expect(result.value.slice(result.from, result.to)).toBe('hello')
  })

  it('prefixes every selected line', () => {
    const result = prefixLines('one\ntwo', 0, 7, '- ')
    expect(result.value).toBe('- one\n- two')
  })

  it('builds an empty fenced block and places the cursor inside', () => {
    const result = buildCodeFence('', 0, 0, 'ts')
    expect(result.text).toBe('```ts\n\n```')
    expect(result.text.slice(0, result.selectFrom)).toBe('```ts\n')
    expect(result.text.slice(result.selectFrom)).toBe('\n```')
  })

  it('wraps selected text in a fenced block', () => {
    const result = buildCodeFence('const x = 1', 0, 11, 'ts')
    expect(result.text).toBe('```ts\nconst x = 1\n```')
    expect(result.text.slice(result.selectFrom, result.selectTo)).toBe('const x = 1')
  })

  it('sanitizes the language token', () => {
    expect(normalizeLanguage(' ```ts ')).toBe('ts')
    expect(normalizeLanguage('c++')).toBe('c++')
  })

  it('uses a longer fence when the selection contains backticks', () => {
    const result = buildCodeFence('```inner```', 0, 11, 'md')
    expect(result.text.startsWith('````md\n')).toBe(true)
    expect(result.text.endsWith('\n````')).toBe(true)
  })
})

describe('scanMarkdownLines', () => {
  it('marks fences, code lines, quotes and rules', () => {
    const source = [
      '# Title',
      '> quote',
      '```ts',
      'const x = 1',
      '```',
      '---',
      'after',
    ].join('\n')
    const scan = scanMarkdownLines(source)
    expect(scan.lines.map((spec) => spec.kind)).toEqual([
      'heading-1',
      'quote',
      'code-fence',
      'code-lang',
      'code-block',
      'code-fence',
      'hr',
    ])
    const lang = scan.lines.find((spec) => spec.kind === 'code-lang')
    expect(lang?.from).toBe(source.indexOf('ts'))
    expect(lang?.to).toBe(source.indexOf('ts') + 2)
    expect(scan.inline.some((spec) => spec.className === 'cm-md-heading-marker')).toBe(true)
  })

  it('keeps unterminated fences styled to the end', () => {
    const scan = scanMarkdownLines('```\ncode\nmore')
    expect(scan.lines.map((spec) => spec.kind)).toEqual(['code-fence', 'code-block', 'code-block'])
  })

  it('marks inline emphasis, code and links', () => {
    const scan = scanMarkdownLines('**bold** and `code` and [link](https://example.com)')
    expect(scan.inline.map((spec) => spec.className)).toEqual([
      'cm-md-code',
      'cm-md-link',
      'cm-md-strong',
    ])
  })

  it('does not italicize snake_case and trims language chips', () => {
    const source = 'snake_case and _italic_\n```  ts  '
    const scan = scanMarkdownLines(source)
    const emphasis = scan.inline.filter((spec) => spec.className === 'cm-md-em')
    expect(emphasis).toHaveLength(1)
    expect(source.slice(emphasis[0].from, emphasis[0].to)).toBe('_italic_')
    const language = scan.lines.find((spec) => spec.kind === 'code-lang')
    expect(language && source.slice(language.from, language.to)).toBe('ts')
  })
})
