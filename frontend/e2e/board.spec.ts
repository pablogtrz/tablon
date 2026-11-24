import { expect, test } from '@playwright/test'
import { createIssue, createProject, deleteProject, setToken, type TestProject } from './helpers'

test.describe('board', () => {
  let project: TestProject

  test.beforeEach(async ({ page, request }) => {
    project = await createProject(request, 'Board project')
    await createIssue(request, project.id, 'Board one', 'todo')
    await createIssue(request, project.id, 'Board two', 'todo')
    await setToken(page)
  })

  test.afterEach(async ({ request }) => {
    if (project) await deleteProject(request, project.id)
  })

  test('drags and uses the move menu, then persists after reload', async ({ page }) => {
    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('tab', { name: /Board/ }).click()

    const todo = page.locator('[data-column-status="todo"]')
    await expect(todo.getByText('Board one')).toBeVisible()
    await expect(todo.getByText('Board two')).toBeVisible()

    const first = page.locator('[data-issue-id]').filter({ hasText: 'Board one' })
    await first.dragTo(page.locator('[data-column-status="in-progress"]'))
    await expect(
      page.locator('[data-column-status="in-progress"]').getByText('Board one'),
    ).toBeVisible()

    const second = page.locator('[data-issue-id]').filter({ hasText: 'Board two' })
    await second.locator('summary').click()
    await second.getByRole('button', { name: 'In Review' }).click()
    await expect(
      page.locator('[data-column-status="in-review"]').getByText('Board two'),
    ).toBeVisible()

    await page.reload()
    await page.getByRole('tab', { name: /Board/ }).click()
    await expect(
      page.locator('[data-column-status="in-progress"]').getByText('Board one'),
    ).toBeVisible()
    await expect(
      page.locator('[data-column-status="in-review"]').getByText('Board two'),
    ).toBeVisible()
  })

  test('moves with the keyboard and stays within the mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('tab', { name: /Board/ }).click()

    const card = page.locator('[data-issue-id]').filter({ hasText: 'Board one' })
    await card.focus()
    await page.keyboard.press('Alt+ArrowRight')
    await expect(
      page.locator('[data-column-status="in-progress"]').getByText('Board one'),
    ).toBeVisible()

    const noOverflow = await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth + 1,
    )
    expect(noOverflow).toBe(true)
  })
})
