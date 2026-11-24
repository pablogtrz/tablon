import { expect, test } from '@playwright/test'
import {
  createEpic,
  createIssue,
  createProject,
  deleteProject,
  setToken,
  type TestProject,
} from './helpers'

test.describe('list and board preferences', () => {
  let project: TestProject

  test.beforeEach(async ({ page, request }) => {
    project = await createProject(request, 'Preference project')
    await createIssue(request, project.id, 'Preference one', 'todo')
    await createIssue(request, project.id, 'Preference two', 'done')
    await setToken(page)
  })

  test.afterEach(async ({ request }) => {
    if (project) await deleteProject(request, project.id)
  })

  test('persists group by status across reloads', async ({ page }) => {
    await page.goto(`/projects/${project.identifier}`)
    await page.getByLabel('Group by').selectOption('status')
    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Done' })).toBeVisible()

    await page.reload()
    await expect(page.getByLabel('Group by')).toHaveValue('status')
    await expect(page.getByRole('heading', { name: 'Todo' })).toBeVisible()
  })

  test('shows the optional backlog column and remembers it', async ({ page }) => {
    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('tab', { name: /Board/ }).click()

    const toggle = page.getByLabel('Show Backlog column')
    await expect(toggle).not.toBeChecked()
    await toggle.check()
    await expect(page.locator('[data-column-status="backlog"]')).toBeVisible()

    await page.reload()
    await page.getByRole('tab', { name: /Board/ }).click()
    await expect(page.getByLabel('Show Backlog column')).toBeChecked()
    await expect(page.locator('[data-column-status="backlog"]')).toBeVisible()
  })

  test('previews ten Done cards and expands or collapses them', async ({ page, request }) => {
    for (let index = 1; index <= 11; index += 1) {
      await createIssue(request, project.id, `Done ${index}`, 'done')
    }

    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('tab', { name: /Board/ }).click()

    const done = page.locator('[data-column-status="done"]')
    await expect(done.locator('[data-issue-id]')).toHaveCount(10)
    await done.getByRole('button', { name: 'See all 12' }).click()
    await expect(done.locator('[data-issue-id]')).toHaveCount(12)
    await done.getByRole('button', { name: 'Show less' }).click()
    await expect(done.locator('[data-issue-id]')).toHaveCount(10)
  })
})

test.describe('epic detail', () => {
  let project: TestProject

  test.beforeEach(async ({ page, request }) => {
    project = await createProject(request, 'Epic project')
    await setToken(page)
  })

  test.afterEach(async ({ request }) => {
    if (project) await deleteProject(request, project.id)
  })

  test('opens an epic card into a Markdown detail view with its tasks', async ({ page, request }) => {
    const epic = await createEpic(request, project.id, 'Epic launch', '**Ship** the thing')
    await createIssue(request, project.id, 'Epic task', 'in-progress', epic.id)

    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('tab', { name: /Epics/ }).click()
    await page.getByRole('link', { name: 'Epic launch' }).click()

    await expect(page).toHaveURL(new RegExp(`/epics/${epic.id}$`))
    await expect(page.getByRole('heading', { name: 'Epic launch' })).toBeVisible()
    await expect(page.getByText('Ship', { exact: true })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Epic task' }).first()).toBeVisible()
  })
})
