import { expect, test } from '@playwright/test'
import {
  createIssue,
  createProject,
  deleteProject,
  setToken,
  type TestIssue,
  type TestProject,
} from './helpers'

test.describe('unsaved guards', () => {
  let project: TestProject
  let issue: TestIssue

  test.beforeEach(async ({ page, request }) => {
    project = await createProject(request, 'Editor project')
    issue = await createIssue(request, project.id, 'Editor target', 'todo')
    await setToken(page)
  })

  test.afterEach(async ({ request }) => {
    if (project) await deleteProject(request, project.id)
  })

  test('stays on the issue when leaving with an unsaved description', async ({ page }) => {
    await page.goto(`/projects/${project.identifier}/issues/${issue.identifier}`)
    await page.getByRole('button', { name: 'Edit', exact: true }).click()
    const editor = page.locator('[data-testid="markdown-editor"] .cm-content').first()
    await editor.fill('Unsaved description text')

    await page
      .getByRole('region', { name: 'Issue detail' })
      .getByRole('link', { name: project.name })
      .click()
    const dialog = page.getByRole('dialog', { name: 'Leave without saving?' })
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: 'Stay' }).click()

    await expect(editor).toContainText('Unsaved description text')
    await expect(page).toHaveURL(new RegExp(`/projects/${project.identifier}/issues/${issue.identifier}`))
  })
})
