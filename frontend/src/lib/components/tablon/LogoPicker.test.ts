import { afterEach, describe, expect, it } from 'vitest'
import { flushSync, mount, unmount } from 'svelte'
import LogoPicker from './LogoPicker.svelte'

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
})

describe('LogoPicker', () => {
  it('reveals the gradient maker when Custom is selected', () => {
    app = mount(LogoPicker, {
      target: document.body,
      props: { value: 0, from: '', to: '', angle: 135 },
    })

    expect(document.body.textContent).not.toContain('Angle')

    const custom = document.querySelector<HTMLInputElement>('input[aria-label="Custom gradient"]')
    custom?.click()
    flushSync()

    expect(document.body.textContent).toContain('Angle')
    expect(
      document.querySelector<HTMLInputElement>('input[aria-label="Gradient start colour"]')?.value,
    ).toBe('#ffa443')
  })

  it('returns to a curated tile when one is selected', () => {
    app = mount(LogoPicker, {
      target: document.body,
      props: { value: 0, from: '#112233', to: '#445566', angle: 45 },
    })

    expect(document.body.textContent).toContain('Angle')

    const amber = document.querySelector<HTMLInputElement>('input[aria-label="Amber"]')
    amber?.click()
    flushSync()

    expect(document.body.textContent).not.toContain('Angle')
  })
})
