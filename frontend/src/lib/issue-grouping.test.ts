import { describe, expect, it } from 'vitest'
import { groupIssues } from './issue-grouping'
import type { Epic, Issue, IssueStatus } from './types'

function issue(id: string, status: IssueStatus, epicID: string | null = null, agentID: string | null = null): Issue {
  return {
    id,
    identifier: `TST-${id}`,
    project_id: 'project',
    epic_id: epicID,
    title: `Issue ${id}`,
    description: '',
    status,
    priority: 'none',
    agent_id: agentID,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

function epic(id: string, title: string): Epic {
  return {
    id,
    project_id: 'project',
    title,
    description: '',
    status: 'in-progress',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

const epics = [epic('epic-a', 'Alpha'), epic('epic-b', 'Beta')]

describe('issue grouping', () => {
  it('returns one group when grouping is off', () => {
    const issues = [issue('1', 'todo'), issue('2', 'done')]
    expect(groupIssues(issues, 'none', epics)).toEqual([
      { key: 'all', label: 'All issues', issues, kind: 'all' },
    ])
  })

  it('groups by status in canonical order and hides empty groups', () => {
    const issues = [
      issue('1', 'done'),
      issue('2', 'todo'),
      issue('3', 'todo'),
      issue('4', 'in-review'),
    ]
    const groups = groupIssues(issues, 'status', epics)
    expect(groups.map((group) => group.key)).toEqual(['todo', 'in-review', 'done'])
    expect(groups[0].issues.map((item) => item.id)).toEqual(['2', '3'])
  })

  it('groups by epic and collects unassigned issues', () => {
    const issues = [
      issue('1', 'todo', 'epic-b'),
      issue('2', 'todo', 'epic-a'),
      issue('3', 'todo', null),
      issue('4', 'todo', 'missing-epic'),
    ]
    const groups = groupIssues(issues, 'epic', epics)
    expect(groups.map((group) => group.key)).toEqual(['epic-a', 'epic-b', 'no-epic'])
    expect(groups[2].issues.map((item) => item.id)).toEqual(['3', '4'])
  })

  it('groups by agent alphabetically and keeps unassigned last', () => {
    const issues = [
      issue('1', 'todo', null, 'zeta'),
      issue('2', 'todo', null, 'alpha'),
      issue('3', 'todo'),
      issue('4', 'todo', null, 'zeta'),
    ]
    const groups = groupIssues(issues, 'agent', epics)
    expect(groups.map((group) => group.key)).toEqual(['alpha', 'zeta', 'unassigned'])
    expect(groups[1].issues.map((item) => item.id)).toEqual(['1', '4'])
  })
})
