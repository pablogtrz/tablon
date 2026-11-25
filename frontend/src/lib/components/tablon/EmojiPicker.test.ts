import { afterEach, describe, expect, it } from 'vitest'
import { flushSync, mount, unmount } from 'svelte'
import EmojiPicker from './EmojiPicker.svelte'

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
  localStorage.clear()
})

function trigger() {
  return document.querySelector<HTMLButtonElement>('button[aria-haspopup="dialog"]')
}

function setInput(input: HTMLInputElement, value: string) {
  input.value = value
  input.dispatchEvent(new Event('input', { bubbles: true }))
  flushSync()
}

describe('EmojiPicker', () => {
  it('selects an emoji from the grid and records it as recent', () => {
    app = mount(EmojiPicker, { target: document.body, props: { value: '' } })
    expect(trigger()?.textContent).toContain('Choose emoji')
    trigger()?.click()
    flushSync()

    document.querySelector<HTMLButtonElement>('button[aria-label="Travel"]')?.click()
    flushSync()
    document.querySelector<HTMLButtonElement>('button[aria-label="Rocket"]')?.click()
    flushSync()

    expect(trigger()?.textContent).toContain('🚀')
    expect(trigger()?.textContent).toContain('Change emoji')
    expect(document.body.textContent).toContain('Recent')
  })

  it('supports a custom emoji through the footer input', () => {
    app = mount(EmojiPicker, { target: document.body, props: { value: '' } })
    trigger()?.click()
    flushSync()

    const customButton = [...document.querySelectorAll('button')].find((button) =>
      button.textContent?.includes('Custom emoji'),
    )
    customButton?.click()
    flushSync()

    const input = document.querySelector<HTMLInputElement>(
      'input[placeholder="Paste or type an emoji"]',
    )
    expect(input).toBeTruthy()
    setInput(input as HTMLInputElement, '🧪')

    const use = [...document.querySelectorAll('button')].find(
      (button) => button.textContent?.trim() === 'Use',
    )
    use?.click()
    flushSync()

    expect(trigger()?.textContent).toContain('🧪')
  })

  it('filters the grid with search', () => {
    app = mount(EmojiPicker, { target: document.body, props: { value: '' } })
    trigger()?.click()
    flushSync()

    const search = document.querySelector<HTMLInputElement>('input[type="search"]')
    setInput(search as HTMLInputElement, 'rocket')

    expect(document.body.textContent).toContain('Search results')
    expect(document.querySelectorAll('button[aria-label="Rocket"]')).toHaveLength(1)
  })
})
