import { ISSUE_STATUS_META } from './meta'
import { ISSUE_STATUS_LABELS, ISSUE_STATUSES } from './types'
import type { Epic, Issue } from './types'

export const LIST_GROUP_BY_OPTIONS = ['none', 'epic', 'status', 'agent'] as const

export type ListGroupBy = (typeof LIST_GROUP_BY_OPTIONS)[number]

export const LIST_GROUP_BY_LABELS: Record<ListGroupBy, string> = {
  none: 'None',
  epic: 'Epic',
  status: 'Status',
  agent: 'Agent',
}

export interface IssueGroup {
  key: string
  label: string
  tone?: string
  issues: Issue[]
  kind: 'all' | 'epic' | 'status' | 'agent' | 'empty'
}

export function isListGroupBy(value: string | null): value is ListGroupBy {
  return value !== null && (LIST_GROUP_BY_OPTIONS as readonly string[]).includes(value)
}

export function groupIssues(
  issues: Issue[],
  groupBy: ListGroupBy,
  epics: Epic[],
): IssueGroup[] {
  if (groupBy === 'none') {
    return [{ key: 'all', label: 'All issues', issues, kind: 'all' }]
  }

  if (groupBy === 'status') {
    return ISSUE_STATUSES.map((status) => ({
      key: status,
      label: ISSUE_STATUS_LABELS[status],
      tone: ISSUE_STATUS_META[status].tone,
      issues: issues.filter((issue) => issue.status === status),
      kind: 'status' as const,
    })).filter((group) => group.issues.length > 0)
  }

  if (groupBy === 'epic') {
    const epicByID = new Map(epics.map((epic) => [epic.id, epic]))
    const groups: IssueGroup[] = epics
      .map((epic) => ({
        key: epic.id,
        label: epic.title,
        issues: issues.filter((issue) => issue.epic_id === epic.id),
        kind: 'epic' as const,
      }))
      .filter((group) => group.issues.length > 0)

    const known = new Set(epicByID.keys())
    const unassigned = issues.filter((issue) => !issue.epic_id || !known.has(issue.epic_id))
    if (unassigned.length > 0) {
      groups.push({ key: 'no-epic', label: 'No epic', issues: unassigned, kind: 'empty' })
    }
    return groups
  }

  const agents = [
    ...new Set(
      issues
        .map((issue) => issue.agent_id)
        .filter((agent): agent is string => Boolean(agent)),
    ),
  ].sort((a, b) => a.localeCompare(b))

  const groups: IssueGroup[] = agents
    .map((agent) => ({
      key: agent,
      label: agent,
      issues: issues.filter((issue) => issue.agent_id === agent),
      kind: 'agent' as const,
    }))
    .filter((group) => group.issues.length > 0)

  const unassigned = issues.filter((issue) => !issue.agent_id)
  if (unassigned.length > 0) {
    groups.push({ key: 'unassigned', label: 'Unassigned', issues: unassigned, kind: 'empty' })
  }
  return groups
}
