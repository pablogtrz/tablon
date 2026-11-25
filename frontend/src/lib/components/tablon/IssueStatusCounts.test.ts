import { afterEach, describe, expect, it } from 'vitest'
import { mount, unmount } from 'svelte'
import IssueStatusCounts from './IssueStatusCounts.svelte'
import type { IssueStatus } from '$lib/types'

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
})

function counts(overrides: Partial<Record<IssueStatus, number>> = {}): Record<IssueStatus, number> {
  return {
    backlog: 0,
    todo: 0,
    'in-progress': 0,
    'in-review': 0,
    done: 0,
    canceled: 0,
    ...overrides,
  }
}

function countFor(label: string): string | null {
  const term = [...document.querySelectorAll('dt')].find((node) => node.textContent?.trim() === label)
  return term?.parentElement?.querySelector('dd')?.textContent?.trim() ?? null
}

describe('IssueStatusCounts', () => {
  it('renders a number-first stack for every present status', () => {
    app = mount(IssueStatusCounts, {
      target: document.body,
      props: { counts: counts({ todo: 2, 'in-review': 3, done: 1 }) },
    })

    expect(countFor('Todo')).toBe('2')
    expect(countFor('In Review')).toBe('3')
    expect(countFor('Done')).toBe('1')
    expect(countFor('Backlog')).toBeNull()
    expect(countFor('Canceled')).toBeNull()
    expect(document.querySelector('dl')?.getAttribute('aria-label')).toContain('2 Todo')
  })

  it('shows label totals and an empty state', () => {
    app = mount(IssueStatusCounts, {
      target: document.body,
      props: { counts: counts(), label: 'Tasks by status' },
    })

    expect(document.body.textContent).toContain('Tasks by status')
    expect(document.body.textContent).toContain('0 tasks')
    expect(document.body.textContent).toContain('No tasks yet')
  })

  it('uses the singular task label for one task', () => {
    app = mount(IssueStatusCounts, {
      target: document.body,
      props: { counts: counts({ todo: 1 }), label: 'Tasks by status' },
    })

    expect(document.body.textContent).toContain('1 task')
    expect(document.body.textContent).not.toContain('1 tasks')
  })
})
