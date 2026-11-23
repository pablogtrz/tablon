import DOMPurify from 'dompurify'
import { marked } from 'marked'

export type MarkdownAction =
  | 'bold'
  | 'italic'
  | 'strikethrough'
  | 'heading'
  | 'bullet'
  | 'numbered'
  | 'task'
  | 'quote'
  | 'code'
  | 'codeblock'
  | 'link'
  | 'image'
  | 'table'

const allowedTags = [
  'p',
  'br',
  'strong',
  'em',
  'del',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'ul',
  'ol',
  'li',
  'blockquote',
  'code',
  'pre',
  'a',
  'img',
  'table',
  'thead',
  'tbody',
  'tr',
  'th',
  'td',
  'hr',
  'input',
]

const allowedAttributes = [
  'href',
  'title',
  'src',
  'alt',
  'type',
  'checked',
  'class',
  'align',
]

export function renderMarkdown(source: string): DocumentFragment {
  const fragment = document.createDocumentFragment()
  if (!source.trim()) return fragment

  const raw = marked.parse(source, {
    async: false,
    gfm: true,
    breaks: true,
  }) as string

  const sanitized = DOMPurify.sanitize(raw, {
    RETURN_DOM_FRAGMENT: true,
    ALLOWED_TAGS: allowedTags,
    ALLOWED_ATTR: allowedAttributes,
    ALLOW_DATA_ATTR: false,
    ALLOW_ARIA_ATTR: false,
  }) as DocumentFragment

  fragment.append(sanitized)

  for (const heading of [...fragment.querySelectorAll('h1')]) {
    const replacement = document.createElement('h2')
    for (const attribute of [...heading.attributes]) {
      replacement.setAttribute(attribute.name, attribute.value)
    }
    replacement.innerHTML = heading.innerHTML
    heading.replaceWith(replacement)
  }

  let taskIndex = 0
  for (const input of fragment.querySelectorAll<HTMLInputElement>('input[type="checkbox"]')) {
    input.removeAttribute('disabled')
    input.setAttribute('data-task-index', String(taskIndex))
    taskIndex += 1

    const listItem = input.closest('li')
    listItem?.classList.add('task-list-item')
    const list = listItem?.parentElement
    if (list && (list.tagName === 'UL' || list.tagName === 'OL')) {
      list.classList.add('contains-task-list')
    }
  }

  for (const link of fragment.querySelectorAll<HTMLAnchorElement>('a')) {
    const href = link.getAttribute('href') ?? ''
    if (href.startsWith('#')) continue
    link.setAttribute('target', '_blank')
    link.setAttribute('rel', 'noopener noreferrer')
  }

  for (const image of [...fragment.querySelectorAll<HTMLImageElement>('img')]) {
    const src = image.getAttribute('src') ?? ''
    if (!/^https:\/\//i.test(src)) {
      const alt = image.getAttribute('alt') ?? ''
      image.replaceWith(document.createTextNode(alt ? `[Image: ${alt}]` : ''))
      continue
    }
    image.setAttribute('loading', 'lazy')
    image.setAttribute('referrerpolicy', 'no-referrer')
    image.setAttribute('decoding', 'async')
  }

  for (const table of [...fragment.querySelectorAll('table')]) {
    const wrapper = document.createElement('div')
    wrapper.className = 'markdown-table'
    wrapper.setAttribute('role', 'region')
    wrapper.setAttribute('tabindex', '0')
    wrapper.setAttribute('aria-label', 'Table')
    table.replaceWith(wrapper)
    wrapper.append(table)
  }

  return fragment
}

const taskPattern = /^(\s*(?:[-*+]|\d+\.)\s+\[)([ xX])(\])/

export function toggleTask(source: string, index: number): string {
  const lines = source.split('\n')
  let inFence = false
  let taskIndex = 0

  for (let lineIndex = 0; lineIndex < lines.length; lineIndex += 1) {
    const trimmed = lines[lineIndex].trimStart()
    if (trimmed.startsWith('```') || trimmed.startsWith('~~~')) {
      inFence = !inFence
      continue
    }
    if (inFence) continue

    const match = lines[lineIndex].match(taskPattern)
    if (!match) continue
    if (taskIndex === index) {
      const next = match[2] === ' ' ? 'x' : ' '
      const markerOffset = (match.index ?? 0) + match[1].length
      lines[lineIndex] =
        lines[lineIndex].slice(0, markerOffset) +
        next +
        lines[lineIndex].slice(markerOffset + 1)
      return lines.join('\n')
    }
    taskIndex += 1
  }

  return source
}
