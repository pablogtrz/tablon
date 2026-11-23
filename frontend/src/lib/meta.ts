import type { Component } from 'svelte'
import Activity from '@lucide/svelte/icons/activity'
import Archive from '@lucide/svelte/icons/archive'
import CalendarClock from '@lucide/svelte/icons/calendar-clock'
import Circle from '@lucide/svelte/icons/circle'
import CircleCheck from '@lucide/svelte/icons/circle-check'
import CircleDashed from '@lucide/svelte/icons/circle-dashed'
import CircleDot from '@lucide/svelte/icons/circle-dot'
import CircleX from '@lucide/svelte/icons/circle-x'
import Eye from '@lucide/svelte/icons/eye'
import Flame from '@lucide/svelte/icons/flame'
import Minus from '@lucide/svelte/icons/minus'
import SignalHigh from '@lucide/svelte/icons/signal-high'
import SignalLow from '@lucide/svelte/icons/signal-low'
import SignalMedium from '@lucide/svelte/icons/signal-medium'
import {
  EPIC_STATUS_LABELS,
  EPIC_STATUS_TONES,
  ISSUE_PRIORITY_LABELS,
  ISSUE_PRIORITY_TONES,
  ISSUE_STATUS_LABELS,
  ISSUE_STATUS_TONES,
  PROJECT_STATUS_LABELS,
  PROJECT_STATUS_TONES,
  type EpicStatus,
  type IssuePriority,
  type IssueStatus,
  type ProjectStatus,
} from './types'

export interface StatusMeta {
  label: string
  tone: string
  soft: string
  icon: Component
}

export interface PriorityMeta {
  label: string
  tone: string
  icon: Component
}

export interface SelectOption<T extends string = string> {
  value: T
  label: string
  tone?: string
  icon?: Component
}

export const ISSUE_STATUS_META: Record<IssueStatus, StatusMeta> = {
  backlog: {
    label: ISSUE_STATUS_LABELS.backlog,
    tone: ISSUE_STATUS_TONES.backlog,
    soft: 'var(--status-backlog-soft)',
    icon: CircleDashed,
  },
  todo: {
    label: ISSUE_STATUS_LABELS.todo,
    tone: ISSUE_STATUS_TONES.todo,
    soft: 'var(--status-todo-soft)',
    icon: Circle,
  },
  'in-progress': {
    label: ISSUE_STATUS_LABELS['in-progress'],
    tone: ISSUE_STATUS_TONES['in-progress'],
    soft: 'var(--status-progress-soft)',
    icon: CircleDot,
  },
  'in-review': {
    label: ISSUE_STATUS_LABELS['in-review'],
    tone: ISSUE_STATUS_TONES['in-review'],
    soft: 'var(--status-review-soft)',
    icon: Eye,
  },
  done: {
    label: ISSUE_STATUS_LABELS.done,
    tone: ISSUE_STATUS_TONES.done,
    soft: 'var(--status-done-soft)',
    icon: CircleCheck,
  },
  canceled: {
    label: ISSUE_STATUS_LABELS.canceled,
    tone: ISSUE_STATUS_TONES.canceled,
    soft: 'var(--status-canceled-soft)',
    icon: CircleX,
  },
}

export const PRIORITY_META: Record<IssuePriority, PriorityMeta> = {
  urgent: {
    label: ISSUE_PRIORITY_LABELS.urgent,
    tone: ISSUE_PRIORITY_TONES.urgent,
    icon: Flame,
  },
  high: {
    label: ISSUE_PRIORITY_LABELS.high,
    tone: ISSUE_PRIORITY_TONES.high,
    icon: SignalHigh,
  },
  medium: {
    label: ISSUE_PRIORITY_LABELS.medium,
    tone: ISSUE_PRIORITY_TONES.medium,
    icon: SignalMedium,
  },
  low: {
    label: ISSUE_PRIORITY_LABELS.low,
    tone: ISSUE_PRIORITY_TONES.low,
    icon: SignalLow,
  },
  none: {
    label: ISSUE_PRIORITY_LABELS.none,
    tone: ISSUE_PRIORITY_TONES.none,
    icon: Minus,
  },
}

export const PROJECT_STATUS_META: Record<ProjectStatus, StatusMeta> = {
  planning: {
    label: PROJECT_STATUS_LABELS.planning,
    tone: PROJECT_STATUS_TONES.planning,
    soft: 'var(--status-backlog-soft)',
    icon: CalendarClock,
  },
  active: {
    label: PROJECT_STATUS_LABELS.active,
    tone: PROJECT_STATUS_TONES.active,
    soft: 'var(--status-progress-soft)',
    icon: Activity,
  },
  archived: {
    label: PROJECT_STATUS_LABELS.archived,
    tone: PROJECT_STATUS_TONES.archived,
    soft: 'var(--status-canceled-soft)',
    icon: Archive,
  },
  done: {
    label: PROJECT_STATUS_LABELS.done,
    tone: PROJECT_STATUS_TONES.done,
    soft: 'var(--status-done-soft)',
    icon: CircleCheck,
  },
}

export const EPIC_STATUS_META: Record<EpicStatus, StatusMeta> = {
  backlog: {
    label: EPIC_STATUS_LABELS.backlog,
    tone: EPIC_STATUS_TONES.backlog,
    soft: 'var(--status-backlog-soft)',
    icon: CircleDashed,
  },
  'in-progress': {
    label: EPIC_STATUS_LABELS['in-progress'],
    tone: EPIC_STATUS_TONES['in-progress'],
    soft: 'var(--status-progress-soft)',
    icon: CircleDot,
  },
  done: {
    label: EPIC_STATUS_LABELS.done,
    tone: EPIC_STATUS_TONES.done,
    soft: 'var(--status-done-soft)',
    icon: CircleCheck,
  },
}

export const ISSUE_STATUS_OPTIONS: SelectOption<IssueStatus>[] = (
  Object.keys(ISSUE_STATUS_META) as IssueStatus[]
).map((value) => ({
  value,
  label: ISSUE_STATUS_META[value].label,
  tone: ISSUE_STATUS_META[value].tone,
  icon: ISSUE_STATUS_META[value].icon,
}))

export const PRIORITY_OPTIONS: SelectOption<IssuePriority>[] = (
  Object.keys(PRIORITY_META) as IssuePriority[]
).map((value) => ({
  value,
  label: PRIORITY_META[value].label,
  tone: PRIORITY_META[value].tone,
  icon: PRIORITY_META[value].icon,
}))

export const PROJECT_STATUS_OPTIONS: SelectOption<ProjectStatus>[] = (
  Object.keys(PROJECT_STATUS_META) as ProjectStatus[]
).map((value) => ({
  value,
  label: PROJECT_STATUS_META[value].label,
  tone: PROJECT_STATUS_META[value].tone,
  icon: PROJECT_STATUS_META[value].icon,
}))

export const EPIC_STATUS_OPTIONS: SelectOption<EpicStatus>[] = (
  Object.keys(EPIC_STATUS_META) as EpicStatus[]
).map((value) => ({
  value,
  label: EPIC_STATUS_META[value].label,
  tone: EPIC_STATUS_META[value].tone,
  icon: EPIC_STATUS_META[value].icon,
}))

export const ANY_STATUS_OPTION: SelectOption = {
  value: '',
  label: 'Any status',
  tone: 'var(--text-3)',
}

export const ANY_PRIORITY_OPTION: SelectOption = {
  value: '',
  label: 'Any priority',
  tone: 'var(--text-3)',
}
