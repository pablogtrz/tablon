import { expect, test } from '@playwright/test'
import {
  createIssue,
  createProject,
  deleteProject,
  setToken,
  type TestIssue,
  type TestProject,
} from './helpers'

test.describe('comments', () => {
  let project: TestProject
  let issue: TestIssue

  test.beforeEach(async ({ page, request }) => {
    project = await createProject(request, 'Comments project')
    issue = await createIssue(request, project.id, 'Comment target', 'todo')
    await setToken(page)
  })

  test.afterEach(async ({ request }) => {
    if (project) await deleteProject(request, project.id)
  })

  test('adds, edits and deletes a Markdown comment', async ({ page }) => {
    await page.goto(`/projects/${project.identifier}/issues/${issue.identifier}`)
    await expect(page.getByRole('heading', { name: /Comments/ })).toBeVisible()

    const composer = page.locator('#comment-composer [data-testid="markdown-editor"] .cm-content')
    await composer.fill('First **comment**')
    await page.getByRole('button', { name: 'Add comment' }).click()

    const list = page.locator('ol')
    await expect(list).toContainText('First comment')
    await expect(list.locator('strong')).toHaveText('comment')

    await page.getByRole('button', { name: /Edit comment/ }).first().click()
    await page.locator('ol [data-testid="markdown-editor"] .cm-content').fill('Edited comment')
    await page.getByRole('button', { name: 'Save comment' }).click()
    await expect(list).toContainText('Edited comment')

    await page.getByRole('button', { name: /Delete comment/ }).first().click()
    const dialog = page.getByRole('dialog', { name: 'Delete comment' })
    await dialog.getByRole('button', { name: 'Delete comment' }).click()
    await expect(page.getByText('No comments yet.')).toBeVisible()
  })
})
