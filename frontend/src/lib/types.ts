export const ISSUE_STATUSES = ['backlog', 'todo', 'in-progress', 'in-review', 'done', 'canceled'] as const

export type IssueStatus = (typeof ISSUE_STATUSES)[number]

export const ISSUE_STATUS_LABELS: Record<IssueStatus, string> = {
  backlog: 'Backlog',
  todo: 'Todo',
  'in-progress': 'In Progress',
  'in-review': 'In Review',
  done: 'Done',
  canceled: 'Canceled',
}

export const ISSUE_STATUS_TONES: Record<IssueStatus, string> = {
  backlog: 'var(--status-backlog)',
  todo: 'var(--status-todo)',
  'in-progress': 'var(--status-progress)',
  'in-review': 'var(--status-review)',
  done: 'var(--status-done)',
  canceled: 'var(--status-canceled)',
}

export const ISSUE_PRIORITIES = ['urgent', 'high', 'medium', 'low', 'none'] as const

export type IssuePriority = (typeof ISSUE_PRIORITIES)[number]

export const ISSUE_PRIORITY_LABELS: Record<IssuePriority, string> = {
  urgent: 'Urgent',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
  none: 'None',
}

export const ISSUE_PRIORITY_TONES: Record<IssuePriority, string> = {
  urgent: 'var(--priority-urgent)',
  high: 'var(--priority-high)',
  medium: 'var(--priority-medium)',
  low: 'var(--priority-low)',
  none: 'var(--priority-none)',
}

export const PROJECT_STATUSES = ['planning', 'active', 'archived', 'done'] as const

export type ProjectStatus = (typeof PROJECT_STATUSES)[number]

export const PROJECT_STATUS_LABELS: Record<ProjectStatus, string> = {
  planning: 'Planning',
  active: 'Active',
  archived: 'Archived',
  done: 'Done',
}

export const PROJECT_STATUS_TONES: Record<ProjectStatus, string> = {
  planning: 'var(--status-backlog)',
  active: 'var(--status-progress)',
  archived: 'var(--text-3)',
  done: 'var(--status-done)',
}

export const EPIC_STATUSES = ['backlog', 'in-progress', 'done'] as const

export type EpicStatus = (typeof EPIC_STATUSES)[number]

export const EPIC_STATUS_LABELS: Record<EpicStatus, string> = {
  backlog: 'Backlog',
  'in-progress': 'In Progress',
  done: 'Done',
}

export const EPIC_STATUS_TONES: Record<EpicStatus, string> = {
  backlog: 'var(--status-backlog)',
  'in-progress': 'var(--status-progress)',
  done: 'var(--status-done)',
}

export interface Project {
  id: string
  identifier: string
  name: string
  description: string
  status: ProjectStatus
  logo_gradient: number
  logo_emoji: string
  created_at: string
  updated_at: string
}

export interface ProjectListItem extends Project {
  epic_count: number
  issue_count: number
  done_count: number
}

export interface Epic {
  id: string
  project_id: string
  title: string
  description: string
  status: EpicStatus
  created_at: string
  updated_at: string
}

export interface Issue {
  id: string
  identifier: string
  project_id: string
  epic_id: string | null
  title: string
  description: string
  status: IssueStatus
  priority: IssuePriority
  agent_id: string | null
  created_at: string
  updated_at: string
}

export interface Comment {
  id: string
  issue_id: string
  content: string
  agent_id: string | null
  created_at: string
  updated_at: string
}

export interface Page<T> {
  items: T[]
  next_cursor?: string
}

export interface SystemIdleStatus {
  enabled: boolean
  state: 'active' | 'idle'
  timeout_seconds: number
  last_activity: string
  idle_since: string | null
}

export interface SystemStatus {
  version: string
  auth_enabled: boolean
  database: string
  idle: SystemIdleStatus
}
