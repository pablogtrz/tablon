import { describe, expect, it } from 'vitest'
import { renderMarkdown, toggleTask } from './markdown'

function html(source: string): string {
  const fragment = renderMarkdown(source)
  const div = document.createElement('div')
  div.append(fragment)
  return div.innerHTML
}

describe('renderMarkdown', () => {
  it('renders safe markdown and strips dangerous html', () => {
    const output = html('<script>alert(1)</script>\n\n**bold** <img src="x" onerror="alert(2)">')
    expect(output).toContain('<strong>bold</strong>')
    expect(output).not.toContain('<script')
    expect(output).not.toContain('onerror')
  })

  it('shifts h1 to h2 so the issue title remains the only h1', () => {
    expect(html('# Section')).toContain('<h2>Section</h2>')
  })

  it('renders GFM tables inside a focusable scroll region', () => {
    const output = html('| A | B |\n| - | - |\n| 1 | 2 |')
    expect(output).toContain('class="markdown-table"')
    expect(output).toContain('role="region"')
    expect(output).toContain('tabindex="0"')
    expect(output).toContain('<table>')
  })

  it('adds safe attributes to links and https images', () => {
    const output = html('[Tablon](https://example.com)\n\n![logo](https://example.com/logo.png)')
    expect(output).toContain('target="_blank"')
    expect(output).toContain('rel="noopener noreferrer"')
    expect(output).toContain('loading="lazy"')
    expect(output).toContain('referrerpolicy="no-referrer"')
  })

  it('drops non-https images and keeps their alt text', () => {
    const output = html('![logo](http://example.com/logo.png)')
    expect(output).not.toContain('<img')
    expect(output).toContain('[Image: logo]')
  })

  it('makes task checkboxes interactive and indexed', () => {
    const output = html('- [ ] first\n- [x] second')
    expect(output).toContain('data-task-index="0"')
    expect(output).toContain('data-task-index="1"')
    expect(output).not.toContain('disabled')
    expect(output).toContain('task-list-item')
    expect(output).toContain('contains-task-list')
  })

  it('drops non-checkbox inputs and arbitrary classes', () => {
    const output = html(
      '<input type="password" placeholder="token">\n\n<p class="fixed inset-0 z-50">overlay</p>\n\n- [x] done',
    )
    expect(output).not.toContain('type="password"')
    expect(output).not.toContain('placeholder')
    expect(output).not.toContain('fixed')
    expect(output).not.toContain('inset-0')
    expect(output).not.toContain('z-50')
    expect(output).toContain('type="checkbox"')
    expect(output).toContain('task-list-item')
  })

  it('keeps only the language class on code blocks', () => {
    const output = html('```ts\nconst x = 1\n```')
    expect(output).toContain('class="language-ts"')
  })

  it('keeps safe table alignment and drops invalid alignment', () => {
    const centered = html('| A |\n| :-: |\n| 1 |')
    expect(centered).toContain('align="center"')
  })
})

describe('toggleTask', () => {
  it('toggles the matching task marker', () => {
    const source = '- [ ] first\n- [x] second'
    expect(toggleTask(source, 0)).toBe('- [x] first\n- [x] second')
    expect(toggleTask(source, 1)).toBe('- [ ] first\n- [ ] second')
  })

  it('ignores task markers inside fenced code blocks', () => {
    const source = '- [ ] first\n```\n- [ ] inside code\n```\n- [ ] last'
    expect(toggleTask(source, 1)).toBe('- [ ] first\n```\n- [ ] inside code\n```\n- [x] last')
  })

  it('keeps nested task markers aligned', () => {
    const source = '- [ ] parent\n  - [ ] child'
    expect(toggleTask(source, 1)).toBe('- [ ] parent\n  - [x] child')
  })

  it('returns the source unchanged when the index does not exist', () => {
    const source = '- [ ] only'
    expect(toggleTask(source, 4)).toBe(source)
  })
})

describe('code blocks', () => {
  it('wraps fenced code with a language header and copy button', () => {
    const output = html('```ts\nconst x = 1\n```')
    expect(output).toContain('class="markdown-codeblock"')
    expect(output).toContain('data-language="ts"')
    expect(output).toContain('class="markdown-codeblock-lang">ts</span>')
    expect(output).toContain('class="markdown-codeblock-copy"')
    expect(output).toContain('aria-label="Copy ts block"')
    expect(output).toContain('<pre><code class="language-ts">')
  })

  it('labels unlabelled fences as Code', () => {
    const output = html('```\nplain\n```')
    expect(output).toContain('class="markdown-codeblock-lang">Code</span>')
    expect(output).toContain('aria-label="Copy code block"')
  })
})
