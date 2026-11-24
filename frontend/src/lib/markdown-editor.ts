export interface Selection {
  from: number
  to: number
}

export interface EditResult extends Selection {
  value: string
}

export interface CodeFenceResult {
  text: string
  selectFrom: number
  selectTo: number
}

export const CODE_LANGUAGES = [
  'bash',
  'css',
  'diff',
  'go',
  'html',
  'javascript',
  'json',
  'python',
  'rust',
  'sql',
  'typescript',
  'yaml',
] as const

export function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

export function replaceText(
  value: string,
  from: number,
  to: number,
  text: string,
  selectFrom: number,
  selectTo: number,
): EditResult {
  const next = value.slice(0, from) + text + value.slice(to)
  return {
    value: next,
    from: clamp(selectFrom, 0, next.length),
    to: clamp(selectTo, 0, next.length),
  }
}

export function wrapSelection(
  value: string,
  from: number,
  to: number,
  before: string,
  after: string,
  placeholder: string,
): EditResult {
  const selected = value.slice(from, to)
  const content = selected || placeholder
  return replaceText(
    value,
    from,
    to,
    before + content + after,
    from + before.length,
    from + before.length + content.length,
  )
}

export function prefixLines(
  value: string,
  from: number,
  to: number,
  prefix: string | ((index: number) => string),
): EditResult {
  const lineStart = value.lastIndexOf('\n', from - 1) + 1
  const rawLineEnd = value.indexOf('\n', to)
  const lineEnd = rawLineEnd === -1 ? value.length : rawLineEnd
  const lines = value.slice(lineStart, lineEnd).split('\n')
  const next = lines
    .map((line, index) => (typeof prefix === 'function' ? prefix(index) : prefix) + line)
    .join('\n')
  return replaceText(value, lineStart, lineEnd, next, lineStart, lineStart + next.length)
}

export function insertText(value: string, from: number, to: number, text: string): EditResult {
  return replaceText(value, from, to, text, from + text.length, from + text.length)
}

export function normalizeLanguage(value: string): string {
  return value.trim().replace(/[`\s]+/g, '').slice(0, 32)
}

export function buildCodeFence(
  value: string,
  from: number,
  to: number,
  language: string,
): CodeFenceResult {
  const lang = normalizeLanguage(language)
  const selected = value.slice(from, to)
  const prefix = from > 0 && value[from - 1] !== '\n' ? '\n' : ''
  const suffix = to < value.length && value[to] !== '\n' ? '\n' : ''
  const longestRun = Math.max(
    0,
    ...(selected.match(/`+/g) ?? []).map((run) => run.length),
  )
  const fence = '`'.repeat(Math.max(3, longestRun + 1))
  const opening = fence + lang
  const selectFrom = from + prefix.length + opening.length + 1

  if (selected) {
    return {
      text: `${prefix}${opening}\n${selected}\n${fence}${suffix}`,
      selectFrom,
      selectTo: selectFrom + selected.length,
    }
  }

  return {
    text: `${prefix}${opening}\n\n${fence}${suffix}`,
    selectFrom,
    selectTo: selectFrom,
  }
}
