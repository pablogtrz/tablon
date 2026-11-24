import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, unmount } from 'svelte'
import BoardCard from './BoardCard.svelte'
import type { Issue } from '$lib/types'

const issue: Issue = {
  id: 'issue-1',
  identifier: 'TABLON-17',
  project_id: 'project-1',
  epic_id: 'epic-1',
  title: 'Kanban board',
  description: '',
  status: 'todo',
  priority: 'high',
  agent_id: null,
  created_at: '2026-11-24T10:00:00Z',
  updated_at: '2026-11-24T10:00:00Z',
}

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
})

function props(overrides: Record<string, unknown> = {}) {
  return {
    issue,
    href: '/projects/TABLON/issues/TABLON-17',
    epicTitle: 'Collaboration',
    saving: false,
    menuId: 'move-issue-1',
    onopen: vi.fn(),
    onmove: vi.fn(),
    onkeymove: vi.fn(),
    ondragstart: vi.fn(),
    ondragend: vi.fn(),
    ...overrides,
  }
}

describe('BoardCard', () => {
  it('renders the issue and moves with Alt+ArrowRight', () => {
    const p = props()
    app = mount(BoardCard, { target: document.body, props: p })
    const card = document.querySelector('[data-issue-id="issue-1"]') as HTMLElement
    expect(card.textContent).toContain('TABLON-17')
    expect(card.textContent).toContain('Kanban board')
    expect(card.textContent).toContain('Collaboration')

    card.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', altKey: true, bubbles: true }))
    expect(p.onkeymove).toHaveBeenCalledWith(issue, 1)
  })

  it('moves through the status menu', () => {
    const p = props()
    app = mount(BoardCard, { target: document.body, props: p })
    const details = document.querySelector('details') as HTMLDetailsElement
    details.open = true
    const done = [...details.querySelectorAll('button')].find((button) =>
      button.textContent?.includes('Done'),
    )
    expect(done).toBeTruthy()
    done?.click()
    expect(p.onmove).toHaveBeenCalledWith(issue, 'done')
  })
})
