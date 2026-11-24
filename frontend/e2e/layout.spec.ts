import { expect, test } from '@playwright/test'
import {
  createIssue,
  createProject,
  deleteProject,
  setToken,
  type TestIssue,
  type TestProject,
} from './helpers'

const widths = [360, 390, 768, 1024, 1440]

test('has no horizontal page overflow at target widths', async ({ page, request }) => {
  let project: TestProject | null = null
  let issue: TestIssue | null = null
  try {
    project = await createProject(request, 'Layout project')
    issue = await createIssue(request, project.id, 'Layout issue', 'todo')
    await setToken(page)

    for (const width of widths) {
      await page.setViewportSize({ width, height: 900 })

      await page.goto('/')
      await expect(page.getByRole('heading', { name: /Projects/ })).toBeVisible()
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
        `dashboard overflows at ${width}px`,
      ).toBe(true)

      await page.goto(`/projects/${project.identifier}`)
      await expect(page.getByRole('heading', { name: project.name })).toBeVisible()
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
        `project overflows at ${width}px`,
      ).toBe(true)

      await page.getByRole('tab', { name: /Board/ }).click()
      await expect(page.getByRole('region', { name: 'Issue board' })).toBeVisible()
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
        `board overflows at ${width}px`,
      ).toBe(true)

      await page.goto(`/projects/${project.identifier}/issues/${issue.identifier}`)
      await expect(page.getByRole('heading', { name: /Comments/ })).toBeVisible()
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
        `issue overflows at ${width}px`,
      ).toBe(true)
    }
  } finally {
    if (project) await deleteProject(request, project.id)
  }
})
