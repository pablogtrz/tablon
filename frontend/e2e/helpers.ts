import type { APIRequestContext, Page } from '@playwright/test'

export const API_URL = 'http://127.0.0.1:18080'
export const TOKEN = 'e2e-token'

export function authHeaders() {
  return { Authorization: `Bearer ${TOKEN}` }
}

export async function setToken(page: Page) {
  await page.addInitScript((token) => {
    sessionStorage.setItem('tablon.apiToken', token)
  }, TOKEN)
}

export interface TestProject {
  id: string
  identifier: string
  name: string
}

export interface TestIssue {
  id: string
  identifier: string
  title: string
}

export interface TestEpic {
  id: string
  title: string
}

function uniqueIdentifier(): string {
  const timestamp = Date.now().toString(36).toUpperCase().slice(-7)
  const random = Math.floor(Math.random() * 36).toString(36).toUpperCase()
  return `E${timestamp}${random}`
}

export async function createProject(
  request: APIRequestContext,
  name = 'E2E project',
): Promise<TestProject> {
  const response = await request.post(`${API_URL}/api/v1/projects`, {
    headers: authHeaders(),
    data: { name, identifier: uniqueIdentifier() },
  })
  if (!response.ok()) {
    throw new Error(`create project failed: ${response.status()} ${await response.text()}`)
  }
  return response.json()
}

export async function createIssue(
  request: APIRequestContext,
  projectId: string,
  title: string,
  status = 'todo',
  epicId?: string,
): Promise<TestIssue> {
  const response = await request.post(`${API_URL}/api/v1/projects/${projectId}/issues`, {
    headers: authHeaders(),
    data: { title, status, ...(epicId ? { epic_id: epicId } : {}) },
  })
  if (!response.ok()) {
    throw new Error(`create issue failed: ${response.status()} ${await response.text()}`)
  }
  return response.json()
}

export async function createEpic(
  request: APIRequestContext,
  projectId: string,
  title: string,
  description = '',
): Promise<TestEpic> {
  const response = await request.post(`${API_URL}/api/v1/projects/${projectId}/epics`, {
    headers: authHeaders(),
    data: { title, description },
  })
  if (!response.ok()) {
    throw new Error(`create epic failed: ${response.status()} ${await response.text()}`)
  }
  return response.json()
}

export async function deleteProject(request: APIRequestContext, projectId: string) {
  await request.delete(`${API_URL}/api/v1/projects/${projectId}`, {
    headers: authHeaders(),
  })
}
