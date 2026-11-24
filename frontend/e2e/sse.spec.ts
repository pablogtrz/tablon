import { expect, test } from '@playwright/test'
import {
  API_URL,
  authHeaders,
  createIssue,
  createProject,
  deleteProject,
  setToken,
  type TestIssue,
  type TestProject,
} from './helpers'

test('receives issue and comment updates over SSE', async ({ page, request }) => {
  let project: TestProject | null = null
  let issue: TestIssue | null = null
  try {
    project = await createProject(request, 'SSE project')
    issue = await createIssue(request, project.id, 'SSE target', 'todo')
    await setToken(page)
    await page.goto(`/projects/${project.identifier}/issues/${issue.identifier}`)

    await expect(page.locator('#issue-status')).toHaveValue('todo')

    const status = await request.patch(`${API_URL}/api/v1/issues/${issue.identifier}`, {
      headers: authHeaders(),
      data: { status: 'in-progress' },
    })
    expect(status.ok()).toBe(true)
    await expect(page.locator('#issue-status')).toHaveValue('in-progress', { timeout: 8_000 })

    const comment = await request.post(
      `${API_URL}/api/v1/issues/${issue.identifier}/comments`,
      {
        headers: authHeaders(),
        data: { content: 'Live comment' },
      },
    )
    expect(comment.ok()).toBe(true)
    await expect(page.locator('ol')).toContainText('Live comment', { timeout: 8_000 })
  } finally {
    if (project) await deleteProject(request, project.id)
  }
})
