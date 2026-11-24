import type {
  Comment,
  Epic,
  EpicStatus,
  Issue,
  IssuePriority,
  IssueStatus,
  Page,
  Project,
  ProjectListItem,
  ProjectStatus,
  SystemStatus,
} from './types'

const TOKEN_KEY = 'tablon.apiToken'

export class ApiError extends Error {
  status: number
  code: string
  details?: unknown

  constructor(status: number, code: string, message: string, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

export function getToken(): string {
  return sessionStorage.getItem(TOKEN_KEY) ?? ''
}

export function setToken(token: string) {
  const value = token.trim()
  if (value) {
    sessionStorage.setItem(TOKEN_KEY, value)
  } else {
    sessionStorage.removeItem(TOKEN_KEY)
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(path, { ...init, headers })
  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  let body: unknown = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      body = null
    }
  }

  if (!response.ok) {
    const envelope = body as { error?: { code?: string; message?: string; details?: unknown } } | null
    throw new ApiError(
      response.status,
      envelope?.error?.code ?? 'request_failed',
      envelope?.error?.message ?? response.statusText,
      envelope?.error?.details,
    )
  }

  return body as T
}

function json(body: unknown): RequestInit {
  return {
    method: 'POST',
    body: JSON.stringify(body),
  }
}

function patch(body: unknown): RequestInit {
  return {
    method: 'PATCH',
    body: JSON.stringify(body),
  }
}

function withIdempotency(init: RequestInit, key?: string): RequestInit {
  if (!key) return init
  const headers = new Headers(init.headers)
  headers.set('Idempotency-Key', key)
  return { ...init, headers }
}

export interface CreateProjectInput {
  name: string
  identifier?: string
  description?: string
  status?: ProjectStatus
  logo_gradient?: number
  logo_emoji?: string
}

export interface UpdateProjectInput {
  name?: string
  description?: string
  status?: ProjectStatus
  logo_gradient?: number
  logo_emoji?: string
}

export interface CreateEpicInput {
  title: string
  description?: string
  status?: EpicStatus
}

export interface UpdateEpicInput {
  title?: string
  description?: string
  status?: EpicStatus
}

export interface CreateIssueInput {
  title: string
  description?: string
  status?: IssueStatus
  priority?: IssuePriority
  epic_id?: string | null
  agent_id?: string
}

export interface UpdateIssueInput {
  title?: string
  description?: string
  status?: IssueStatus
  priority?: IssuePriority
  epic_id?: string | null
}

export interface CreateCommentInput {
  content: string
  agent_id?: string
}

export interface UpdateCommentInput {
  content: string
}

export interface IssueListParams {
  status?: string
  priority?: string
  epic?: string
  q?: string
  sort?: string
  order?: 'asc' | 'desc'
  limit?: number
  cursor?: string
}

export async function listProjects(includeArchived = false): Promise<ProjectListItem[]> {
  const suffix = includeArchived ? '?include_archived=true' : ''
  const page = await request<{ items: ProjectListItem[] }>(`/api/v1/projects${suffix}`)
  return page.items
}

export function getProject(ref: string): Promise<Project> {
  return request<Project>(`/api/v1/projects/${encodeURIComponent(ref)}`)
}

export function createProject(input: CreateProjectInput): Promise<Project> {
  return request<Project>('/api/v1/projects', json(input))
}

export function updateProject(ref: string, input: UpdateProjectInput): Promise<Project> {
  return request<Project>(`/api/v1/projects/${encodeURIComponent(ref)}`, patch(input))
}

export function deleteProject(ref: string): Promise<void> {
  return request<void>(`/api/v1/projects/${encodeURIComponent(ref)}`, { method: 'DELETE' })
}

export async function listEpics(projectRef: string): Promise<Epic[]> {
  const page = await request<{ items: Epic[] }>(`/api/v1/projects/${encodeURIComponent(projectRef)}/epics`)
  return page.items
}

export function createEpic(projectRef: string, input: CreateEpicInput): Promise<Epic> {
  return request<Epic>(`/api/v1/projects/${encodeURIComponent(projectRef)}/epics`, json(input))
}

export function getEpic(ref: string): Promise<Epic> {
  return request<Epic>(`/api/v1/epics/${encodeURIComponent(ref)}`)
}

export function updateEpic(ref: string, input: UpdateEpicInput): Promise<Epic> {
  return request<Epic>(`/api/v1/epics/${encodeURIComponent(ref)}`, patch(input))
}

export function deleteEpic(ref: string): Promise<void> {
  return request<void>(`/api/v1/epics/${encodeURIComponent(ref)}`, { method: 'DELETE' })
}

export function listIssues(projectRef: string, params: IssueListParams = {}): Promise<Page<Issue>> {
  const query = new URLSearchParams()
  if (params.status) query.set('status', params.status)
  if (params.priority) query.set('priority', params.priority)
  if (params.epic) query.set('epic_id', params.epic)
  if (params.q) query.set('q', params.q)
  if (params.sort) query.set('sort', params.sort)
  if (params.order) query.set('order', params.order)
  if (params.limit) query.set('limit', String(params.limit))
  if (params.cursor) query.set('cursor', params.cursor)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return request<Page<Issue>>(`/api/v1/projects/${encodeURIComponent(projectRef)}/issues${suffix}`)
}

export interface AllIssues {
  items: Issue[]
  truncated: boolean
}

export async function listAllIssues(
  projectRef: string,
  params: IssueListParams = {},
  cap = 1000,
): Promise<AllIssues> {
  const items: Issue[] = []
  let cursor: string | undefined
  for (;;) {
    const page = await listIssues(projectRef, { ...params, limit: 200, cursor })
    items.push(...page.items)
    if (items.length > cap) {
      items.length = cap
      return { items, truncated: true }
    }
    if (!page.next_cursor) {
      return { items, truncated: false }
    }
    if (items.length === cap) {
      return { items, truncated: true }
    }
    cursor = page.next_cursor
  }
}

export function getIssue(ref: string): Promise<Issue> {
  return request<Issue>(`/api/v1/issues/${encodeURIComponent(ref)}`)
}

export function createIssue(
  projectRef: string,
  input: CreateIssueInput,
  idempotencyKey?: string,
): Promise<Issue> {
  return request<Issue>(
    `/api/v1/projects/${encodeURIComponent(projectRef)}/issues`,
    withIdempotency(json(input), idempotencyKey),
  )
}

export function updateIssue(ref: string, input: UpdateIssueInput): Promise<Issue> {
  return request<Issue>(`/api/v1/issues/${encodeURIComponent(ref)}`, patch(input))
}

export function deleteIssue(ref: string): Promise<void> {
  return request<void>(`/api/v1/issues/${encodeURIComponent(ref)}`, { method: 'DELETE' })
}

export function updateIssueStatus(ref: string, status: IssueStatus): Promise<Issue> {
  return request<Issue>(
    `/api/v1/issues/${encodeURIComponent(ref)}/status`,
    patch({ status }),
  )
}

export function listComments(
  issueRef: string,
  params: { limit?: number; cursor?: string } = {},
): Promise<Page<Comment>> {
  const query = new URLSearchParams()
  if (params.limit) query.set('limit', String(params.limit))
  if (params.cursor) query.set('cursor', params.cursor)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return request<Page<Comment>>(
    `/api/v1/issues/${encodeURIComponent(issueRef)}/comments${suffix}`,
  )
}

export interface AllComments {
  items: Comment[]
  truncated: boolean
  nextCursor?: string
}

export async function listAllComments(issueRef: string, cap = 400): Promise<AllComments> {
  const items: Comment[] = []
  let cursor: string | undefined
  for (;;) {
    const page = await listComments(issueRef, { limit: 200, cursor })
    items.push(...page.items)
    if (items.length >= cap) {
      return {
        items: items.slice(0, cap),
        truncated: Boolean(page.next_cursor),
        nextCursor: page.next_cursor,
      }
    }
    if (!page.next_cursor) {
      return { items, truncated: false }
    }
    cursor = page.next_cursor
  }
}

export function createComment(
  issueRef: string,
  input: CreateCommentInput,
  idempotencyKey?: string,
): Promise<Comment> {
  return request<Comment>(
    `/api/v1/issues/${encodeURIComponent(issueRef)}/comments`,
    withIdempotency(json(input), idempotencyKey),
  )
}

export function updateComment(ref: string, input: UpdateCommentInput): Promise<Comment> {
  return request<Comment>(`/api/v1/comments/${encodeURIComponent(ref)}`, patch(input))
}

export function deleteComment(ref: string): Promise<void> {
  return request<void>(`/api/v1/comments/${encodeURIComponent(ref)}`, { method: 'DELETE' })
}

export function claimIssue(ref: string, agentId: string): Promise<Issue> {
  return request<Issue>(`/api/v1/issues/${encodeURIComponent(ref)}/claim`, json({ agent_id: agentId }))
}

export function releaseIssue(ref: string, agentId: string): Promise<Issue> {
  return request<Issue>(`/api/v1/issues/${encodeURIComponent(ref)}/release`, json({ agent_id: agentId }))
}

export function getSystemStatus(): Promise<SystemStatus> {
  return request<SystemStatus>('/api/v1/system/status')
}
