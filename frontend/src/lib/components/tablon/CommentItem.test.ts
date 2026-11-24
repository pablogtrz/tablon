import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, tick, unmount } from 'svelte'
import CommentItem from './CommentItem.svelte'
import type { Comment } from '$lib/types'

const comment: Comment = {
  id: 'comment-1',
  issue_id: 'issue-1',
  content: '**Started** work',
  agent_id: 'agent-a',
  created_at: '2026-11-24T10:00:00Z',
  updated_at: '2026-11-24T11:00:00Z',
}

let app: ReturnType<typeof mount> | null = null

afterEach(() => {
  if (app) unmount(app)
  app = null
  document.body.innerHTML = ''
})

describe('CommentItem', () => {
  it('renders markdown, agent and edited state', async () => {
    app = mount(CommentItem, {
      target: document.body,
      props: { comment, onedit: vi.fn(), ondelete: vi.fn() },
    })
    await tick()
    const article = document.querySelector('article') as HTMLElement
    expect(article.querySelector('strong')?.textContent).toBe('Started')
    expect(article.textContent).toContain('agent-a')
    expect(article.textContent).toContain('edited')
  })

  it('renders a local label and action callbacks', () => {
    const onedit = vi.fn()
    const ondelete = vi.fn()
    app = mount(CommentItem, {
      target: document.body,
      props: { comment: { ...comment, agent_id: null }, onedit, ondelete },
    })
    expect(document.body.textContent).toContain('Local')
    const buttons = document.querySelectorAll('button')
    buttons[0].click()
    buttons[1].click()
    expect(onedit).toHaveBeenCalled()
    expect(ondelete).toHaveBeenCalled()
  })
})
