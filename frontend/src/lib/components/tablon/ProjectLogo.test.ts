import { afterEach, describe, expect, it } from 'vitest'
import { mount, unmount } from 'svelte'
import ProjectLogo from './ProjectLogo.svelte'

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
})

describe('ProjectLogo', () => {
  it('uses the curated tile when no custom colours are set', () => {
    app = mount(ProjectLogo, { target: document.body, props: { name: 'Tablon', gradient: 3 } })
    const tile = document.body.firstElementChild as HTMLElement
    expect(tile.getAttribute('style')).toContain('--logo-3')
  })

  it('renders a custom two-stop gradient with its angle', () => {
    app = mount(ProjectLogo, {
      target: document.body,
      props: { name: 'Tablon', from: '#112233', to: '#445566', angle: 45 },
    })
    const tile = document.body.firstElementChild as HTMLElement
    expect(tile.getAttribute('style')).toContain(
      'linear-gradient(45deg, rgb(17, 34, 51) 0%, rgb(68, 85, 102) 100%)',
    )
  })
})
