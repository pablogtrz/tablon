import { ISSUE_STATUS_LABELS } from './types'
import type { Issue, IssueStatus } from './types'

export const BOARD_STATUSES = ['todo', 'in-progress', 'in-review', 'done'] as const

export type BoardStatus = (typeof BOARD_STATUSES)[number]

export const HIDDEN_BOARD_STATUSES = ['backlog', 'canceled'] as const

export type HiddenBoardStatus = (typeof HIDDEN_BOARD_STATUSES)[number]

export interface BoardColumn {
  status: BoardStatus
  label: string
}

export const BOARD_COLUMNS: BoardColumn[] = BOARD_STATUSES.map((status) => ({
  status,
  label: ISSUE_STATUS_LABELS[status],
}))

export function isBoardStatus(status: IssueStatus): status is BoardStatus {
  return (BOARD_STATUSES as readonly IssueStatus[]).includes(status)
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

export function groupBoardIssues(issues: Issue[]): Record<BoardStatus, Issue[]> {
  const grouped: Record<BoardStatus, Issue[]> = {
    todo: [],
    'in-progress': [],
    'in-review': [],
    done: [],
  }
  for (const issue of issues) {
    if (isBoardStatus(issue.status)) {
      grouped[issue.status].push(issue)
    }
  }
  for (const status of BOARD_STATUSES) {
    grouped[status] = sortIssues(grouped[status])
  }
  return grouped
}

export function hiddenBoardCounts(issues: Issue[]): Record<HiddenBoardStatus, number> {
  return {
    backlog: issues.filter((issue) => issue.status === 'backlog').length,
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
): BoardStatus | null {
  if (!isBoardStatus(status)) return null
  const index = BOARD_STATUSES.indexOf(status) + direction
  if (index < 0 || index >= BOARD_STATUSES.length) return null
  return BOARD_STATUSES[index]
}

export function boardStatusLabel(status: IssueStatus): string {
  return ISSUE_STATUS_LABELS[status]
}
