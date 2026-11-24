import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, tick, unmount } from 'svelte'
import MarkdownRenderer from './MarkdownRenderer.svelte'

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
  Reflect.deleteProperty(navigator, 'clipboard')
  Reflect.deleteProperty(document, 'execCommand')
})

describe('MarkdownRenderer copy', () => {
  it('copies code and announces the result', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      configurable: true,
    })

    app = mount(MarkdownRenderer, {
      target: document.body,
      props: { source: '```ts\nconst x = 1\n```' },
    })
    await tick()

    const button = document.querySelector('.markdown-codeblock-copy') as HTMLButtonElement
    button.click()
    await Promise.resolve()
    await tick()

    await vi.waitFor(() => expect(button.textContent).toBe('Copied'))
    expect(writeText).toHaveBeenCalledWith('const x = 1')
    expect(document.querySelector('.markdown-codeblock-status')?.textContent).toBe('Code copied')
  })
})

describe('MarkdownRenderer clipboard fallback', () => {
  it('uses execCommand when the clipboard API is unavailable', async () => {
    const exec = vi.fn(() => true)
    Object.defineProperty(document, 'execCommand', {
      value: exec,
      configurable: true,
    })

    app = mount(MarkdownRenderer, {
      target: document.body,
      props: { source: '```\nplain\n```' },
    })
    await tick()

    const button = document.querySelector('.markdown-codeblock-copy') as HTMLButtonElement
    button.click()
    await Promise.resolve()
    await tick()

    await vi.waitFor(() => expect(button.textContent).toBe('Copied'))
    expect(exec).toHaveBeenCalledWith('copy')
  })
})
