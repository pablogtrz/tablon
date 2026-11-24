import { describe, expect, it } from 'vitest'
import {
  adjacentBoardStatus,
  applyBoardMove,
  groupBoardIssues,
  hiddenBoardCounts,
  isBoardStatus,
} from './board'
import type { Issue, IssueStatus } from './types'

function issue(id: string, status: IssueStatus, createdAt: string): Issue {
  return {
    id,
    identifier: `TST-${id}`,
    project_id: 'project',
    epic_id: null,
    title: `Issue ${id}`,
    description: '',
    status,
    priority: 'none',
    agent_id: null,
    created_at: createdAt,
    updated_at: createdAt,
  }
}

describe('board grouping', () => {
  const issues = [
    issue('1', 'todo', '2026-01-01T00:00:00Z'),
    issue('2', 'in-progress', '2026-01-02T00:00:00Z'),
    issue('3', 'in-review', '2026-01-03T00:00:00Z'),
    issue('4', 'done', '2026-01-04T00:00:00Z'),
    issue('5', 'backlog', '2026-01-05T00:00:00Z'),
    issue('6', 'canceled', '2026-01-06T00:00:00Z'),
  ]

  it('puts exact statuses into the four columns and hides the rest', () => {
    const grouped = groupBoardIssues(issues)
    expect(grouped.todo.map((item) => item.id)).toEqual(['1'])
    expect(grouped['in-progress'].map((item) => item.id)).toEqual(['2'])
    expect(grouped['in-review'].map((item) => item.id)).toEqual(['3'])
    expect(grouped.done.map((item) => item.id)).toEqual(['4'])
    expect(hiddenBoardCounts(issues)).toEqual({ backlog: 1, canceled: 1 })
  })

  it('sorts newest first inside a column', () => {
    const grouped = groupBoardIssues([
      issue('old', 'todo', '2026-01-01T00:00:00Z'),
      issue('new', 'todo', '2026-01-02T00:00:00Z'),
    ])
    expect(grouped.todo.map((item) => item.id)).toEqual(['new', 'old'])
  })

  it('moves status immutably and does not notify on a no-op', () => {
    const moved = applyBoardMove(issues, '1', 'in-progress')
    expect(moved.find((item) => item.id === '1')?.status).toBe('in-progress')
    expect(issues.find((item) => item.id === '1')?.status).toBe('todo')
    expect(applyBoardMove(issues, '1', 'todo')).toBe(issues)
  })

  it('finds adjacent visible columns', () => {
    expect(adjacentBoardStatus('todo', 1)).toBe('in-progress')
    expect(adjacentBoardStatus('in-review', 1)).toBe('done')
    expect(adjacentBoardStatus('todo', -1)).toBeNull()
    expect(adjacentBoardStatus('done', 1)).toBeNull()
    expect(adjacentBoardStatus('backlog', 1)).toBeNull()
    expect(isBoardStatus('done')).toBe(true)
    expect(isBoardStatus('backlog')).toBe(false)
  })
})
