import type { Range } from '@codemirror/state'
import {
  Decoration,
  EditorView,
  ViewPlugin,
  type DecorationSet,
  type ViewUpdate,
} from '@codemirror/view'

export type MarkdownLineKind =
  | 'code-block'
  | 'code-fence'
  | 'code-lang'
  | 'quote'
  | 'hr'
  | 'heading-1'
  | 'heading-2'
  | 'heading-3'
  | 'heading-4'
  | 'heading-5'
  | 'heading-6'
  | 'list'
  | 'task'

export interface MarkdownLineSpec {
  from: number
  to?: number
  kind: MarkdownLineKind
}

export interface MarkdownInlineSpec {
  from: number
  to: number
  className: string
}

export interface MarkdownScan {
  lines: MarkdownLineSpec[]
  inline: MarkdownInlineSpec[]
}

export function scanMarkdownLines(source: string): MarkdownScan {
  const lines: MarkdownLineSpec[] = []
  const inline: MarkdownInlineSpec[] = []
  const sourceLines = source.split('\n')
  let offset = 0
  let inFence = false
  let fenceChar = ''
  let fenceLength = 0

  for (const line of sourceLines) {
    const lineFrom = offset
    const lineTo = offset + line.length
    const fence = line.match(/^\s*(`{3,}|~{3,})(.*)$/)

    if (fence) {
      const marker = fence[1]
      const rest = fence[2]
      if (!inFence) {
        inFence = true
        fenceChar = marker[0]
        fenceLength = marker.length
        lines.push({ from: lineFrom, kind: 'code-fence' })
        const trimmed = rest.trim()
        if (trimmed) {
          const start = lineFrom + line.length - rest.length + (rest.length - rest.trimStart().length)
          lines.push({
            from: start,
            to: start + trimmed.length,
            kind: 'code-lang',
          })
        }
      } else if (
        marker[0] === fenceChar &&
        marker.length >= fenceLength &&
        rest.trim() === ''
      ) {
        inFence = false
        lines.push({ from: lineFrom, kind: 'code-fence' })
      } else {
        lines.push({ from: lineFrom, kind: 'code-block' })
      }
    } else if (inFence) {
      lines.push({ from: lineFrom, kind: 'code-block' })
    } else {
      scanMarkdownLine(line, lineFrom, lines, inline)
    }

    offset = lineTo + 1
  }

  return { lines, inline }
}

function scanMarkdownLine(
  line: string,
  lineFrom: number,
  lines: MarkdownLineSpec[],
  inline: MarkdownInlineSpec[],
) {
  const heading = line.match(/^(\s*)(#{1,6})(\s+)/)
  if (heading) {
    const level = heading[2].length
    lines.push({ from: lineFrom, kind: `heading-${level}` as MarkdownLineKind })
    inline.push({
      from: lineFrom + heading[1].length,
      to: lineFrom + heading[1].length + heading[2].length,
      className: 'cm-md-heading-marker',
    })
  }

  if (/^\s*>\s?/.test(line)) {
    lines.push({ from: lineFrom, kind: 'quote' })
  }

  if (/^\s*(?:-{3,}|\*{3,}|_{3,})\s*$/.test(line)) {
    lines.push({ from: lineFrom, kind: 'hr' })
  }

  const task = line.match(/^(\s*[-*+]\s+\[[ xX]\]\s+)/)
  const list = line.match(/^(\s*(?:[-*+]|\d+\.)\s+)/)
  if (task) {
    lines.push({ from: lineFrom, kind: 'task' })
    inline.push({
      from: lineFrom + 0,
      to: lineFrom + task[1].length,
      className: 'cm-md-task-marker',
    })
  } else if (list) {
    lines.push({ from: lineFrom, kind: 'list' })
    inline.push({
      from: lineFrom + 0,
      to: lineFrom + list[1].length,
      className: 'cm-md-list-marker',
    })
  }

  const occupied: Array<{ from: number; to: number }> = []

  const add = (start: number, end: number, className: string) => {
    if (start >= end) return
    if (occupied.some((range) => start < range.to && end > range.from)) return
    occupied.push({ from: start, to: end })
    inline.push({ from: lineFrom + start, to: lineFrom + end, className })
  }

  for (const match of line.matchAll(/`([^`\n]+)`/g)) {
    add(match.index, match.index + match[0].length, 'cm-md-code')
  }

  for (const match of line.matchAll(/\[([^\]\n]+)\]\(([^)\n]+)\)/g)) {
    add(match.index, match.index + match[0].length, 'cm-md-link')
  }

  for (const match of line.matchAll(/\*\*([^*\n]+)\*\*/g)) {
    add(match.index, match.index + match[0].length, 'cm-md-strong')
  }

  for (const match of line.matchAll(/__([^_\n]+)__/g)) {
    add(match.index, match.index + match[0].length, 'cm-md-strong')
  }

  for (const match of line.matchAll(/~~([^~\n]+)~~/g)) {
    add(match.index, match.index + match[0].length, 'cm-md-strike')
  }

  for (const match of line.matchAll(/\*([^*\n]+)\*/g)) {
    add(match.index, match.index + match[0].length, 'cm-md-em')
  }

  for (const match of line.matchAll(/(^|[^\w])_([^_\n]+)_(?!\w)/g)) {
    const start = match.index + match[1].length
    add(start, start + match[2].length + 2, 'cm-md-em')
  }
}

function lineClass(kind: MarkdownLineKind): string | null {
  switch (kind) {
    case 'code-block':
      return 'cm-md-code-block'
    case 'code-fence':
      return 'cm-md-code-fence'
    case 'quote':
      return 'cm-md-quote'
    case 'hr':
      return 'cm-md-hr'
    case 'list':
      return 'cm-md-list'
    case 'task':
      return 'cm-md-task'
    default:
      return kind.startsWith('heading-') ? `cm-md-${kind}` : null
  }
}

function buildDecorations(source: string): DecorationSet {
  const ranges: Range<Decoration>[] = []
  const scan = scanMarkdownLines(source)

  for (const spec of scan.lines) {
    if (spec.kind === 'code-lang' && spec.to !== undefined) {
      ranges.push(Decoration.mark({ class: 'cm-md-code-lang' }).range(spec.from, spec.to))
      continue
    }
    const className = lineClass(spec.kind)
    if (className) {
      ranges.push(Decoration.line({ class: className }).range(spec.from))
    }
  }

  for (const spec of scan.inline) {
    ranges.push(Decoration.mark({ class: spec.className }).range(spec.from, spec.to))
  }

  return Decoration.set(ranges, true)
}

export const markdownLineDecorations = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet

    constructor(view: EditorView) {
      this.decorations = buildDecorations(view.state.doc.toString())
    }

    update(update: ViewUpdate) {
      if (update.docChanged) {
        this.decorations = buildDecorations(update.state.doc.toString())
      }
    }
  },
  {
    decorations: (plugin) => plugin.decorations,
  },
)

export const markdownEditorTheme = EditorView.theme({
  '&': {
    backgroundColor: 'transparent',
    color: 'var(--text)',
    fontFamily: 'var(--mono)',
    fontSize: '13px',
  },
  '&.cm-focused': {
    outline: 'none',
  },
  '.cm-scroller': {
    fontFamily: 'inherit',
    lineHeight: '1.7',
    overflow: 'auto',
    padding: '12px',
  },
  '.cm-content': {
    padding: '0',
    caretColor: 'var(--text)',
  },
  '.cm-line': {
    padding: '0',
  },
  '.cm-selectionBackground, ::selection': {
    backgroundColor: 'color-mix(in srgb, var(--brand) 22%, transparent)',
  },
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: 'color-mix(in srgb, var(--brand) 30%, transparent)',
  },
  '.cm-placeholder': {
    color: 'var(--text-3)',
  },
})
