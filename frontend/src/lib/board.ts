import { ISSUE_STATUS_LABELS } from './types'
import type { Issue, IssueStatus } from './types'

export const BOARD_STATUSES = ['todo', 'in-progress', 'in-review', 'done'] as const

export type BoardStatus = (typeof BOARD_STATUSES)[number]

export const BOARD_STATUSES_WITH_BACKLOG = ['backlog', ...BOARD_STATUSES] as const

export type VisibleBoardStatus = (typeof BOARD_STATUSES_WITH_BACKLOG)[number]

export const HIDDEN_BOARD_STATUSES = ['backlog', 'canceled'] as const

export type HiddenBoardStatus = (typeof HIDDEN_BOARD_STATUSES)[number]

export interface BoardColumn {
  status: VisibleBoardStatus
  label: string
}

export const BOARD_COLUMNS: BoardColumn[] = BOARD_STATUSES.map((status) => ({
  status,
  label: ISSUE_STATUS_LABELS[status],
}))

export function visibleBoardStatuses(includeBacklog: boolean): VisibleBoardStatus[] {
  return includeBacklog ? [...BOARD_STATUSES_WITH_BACKLOG] : [...BOARD_STATUSES]
}

export function visibleBoardColumns(includeBacklog: boolean): BoardColumn[] {
  return visibleBoardStatuses(includeBacklog).map((status) => ({
    status,
    label: ISSUE_STATUS_LABELS[status],
  }))
}

export function isBoardStatus(status: IssueStatus): status is BoardStatus {
  return (BOARD_STATUSES as readonly IssueStatus[]).includes(status)
}

export function isVisibleBoardStatus(
  status: IssueStatus,
  includeBacklog: boolean,
): status is VisibleBoardStatus {
  return includeBacklog ? status !== 'canceled' : isBoardStatus(status)
}

function compareIssues(a: Issue, b: Issue): number {
  if (a.created_at !== b.created_at) {
    return a.created_at < b.created_at ? 1 : -1
  }
  return a.id < b.id ? 1 : a.id > b.id ? -1 : 0
}

export function sortIssues(issues: Issue[]): Issue[] {
  return [...issues].sort(compareIssues)
}

export function groupBoardIssues(
  issues: Issue[],
  includeBacklog = false,
): Record<VisibleBoardStatus, Issue[]> {
  const grouped: Record<VisibleBoardStatus, Issue[]> = {
    backlog: [],
    todo: [],
    'in-progress': [],
    'in-review': [],
    done: [],
  }
  const statuses = visibleBoardStatuses(includeBacklog)
  for (const issue of issues) {
    if (statuses.includes(issue.status as VisibleBoardStatus)) {
      grouped[issue.status as VisibleBoardStatus].push(issue)
    }
  }
  for (const status of statuses) {
    grouped[status] = sortIssues(grouped[status])
  }
  return grouped
}

export function hiddenBoardCounts(
  issues: Issue[],
  includeBacklog = false,
): Record<HiddenBoardStatus, number> {
  return {
    backlog: includeBacklog ? 0 : issues.filter((issue) => issue.status === 'backlog').length,
    canceled: issues.filter((issue) => issue.status === 'canceled').length,
  }
}

export function applyBoardMove(issues: Issue[], issueID: string, status: IssueStatus): Issue[] {
  let changed = false
  const next = issues.map((issue) => {
    if (issue.id !== issueID || issue.status === status) return issue
    changed = true
    return { ...issue, status }
  })
  return changed ? sortIssues(next) : issues
}

export function adjacentBoardStatus(
  status: IssueStatus,
  direction: -1 | 1,
  statuses: VisibleBoardStatus[] = [...BOARD_STATUSES],
): VisibleBoardStatus | null {
  const index = statuses.indexOf(status as VisibleBoardStatus)
  if (index < 0) return null
  const next = index + direction
  if (next < 0 || next >= statuses.length) return null
  return statuses[next]
}

export function boardStatusLabel(status: IssueStatus): string {
  return ISSUE_STATUS_LABELS[status]
}
