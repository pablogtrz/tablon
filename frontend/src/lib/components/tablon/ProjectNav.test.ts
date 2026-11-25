import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { flushSync, mount, unmount } from 'svelte'
import ProjectNav from './ProjectNav.svelte'
import { projectNav } from '$lib/project-nav.svelte'
import { route } from '$lib/router.svelte'
import type { ProjectListItem } from '$lib/types'

const project: ProjectListItem = {
  id: 'project-1',
  identifier: 'TABLON',
  name: 'Tablon',
  description: '',
  status: 'active',
  logo_gradient: 0,
  logo_emoji: '🧭',
  logo_gradient_from: '',
  logo_gradient_to: '',
  logo_gradient_angle: 135,
  created_at: '2026-11-24T10:00:00Z',
  updated_at: '2026-11-24T10:00:00Z',
  epic_count: 1,
  issue_count: 4,
  done_count: 2,
}

let app: ReturnType<typeof mount> | null = null

beforeEach(() => {
  projectNav.items = [project]
  projectNav.loading = false
  projectNav.loaded = true
  projectNav.error = ''
})

afterEach(() => {
  if (app) unmount(app)
  app = null
  route.path = '/'
  document.body.innerHTML = ''
})

function toggle() {
  return document.querySelector<HTMLButtonElement>('button[aria-controls]')
}

describe('ProjectNav', () => {
  it('opens on project routes and marks the current project', () => {
    route.path = '/projects/TABLON'
    app = mount(ProjectNav, { target: document.body })

    expect(toggle()?.getAttribute('aria-expanded')).toBe('true')
    expect(document.body.textContent).toContain('Tablon')
    expect(document.querySelector('a[aria-current="page"]')?.textContent).toContain('Tablon')
  })

  it('collapses on settings routes', () => {
    route.path = '/settings'
    app = mount(ProjectNav, { target: document.body })

    expect(toggle()?.getAttribute('aria-expanded')).toBe('false')
    expect(document.querySelector('a[aria-current="page"]')).toBeNull()
  })

  it('collapses when the route leaves Projects and reopens on return', () => {
    route.path = '/projects/TABLON'
    app = mount(ProjectNav, { target: document.body })
    expect(toggle()?.getAttribute('aria-expanded')).toBe('true')

    route.path = '/settings'
    flushSync()
    expect(toggle()?.getAttribute('aria-expanded')).toBe('false')

    route.path = '/projects/TABLON'
    flushSync()
    expect(toggle()?.getAttribute('aria-expanded')).toBe('true')
  })

  it('toggles from the chevron', () => {
    route.path = '/projects/TABLON'
    app = mount(ProjectNav, { target: document.body })

    toggle()?.click()
    flushSync()
    expect(toggle()?.getAttribute('aria-expanded')).toBe('false')
    expect(document.querySelector('a[href="/projects/TABLON"]')).toBeNull()
  })
})
