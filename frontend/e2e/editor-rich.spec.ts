import { expect, test } from '@playwright/test'
import {
  createIssue,
  createProject,
  deleteProject,
  setToken,
  type TestIssue,
  type TestProject,
} from './helpers'

test.skip(
  ({ browserName }) => browserName !== 'chromium',
  'clipboard permissions are Chromium-only in the smoke suite',
)

test('styles Markdown and inserts a copyable code block', async ({ page, context, request }) => {
  let project: TestProject | null = null
  let issue: TestIssue | null = null
  try {
    project = await createProject(request, 'Rich editor project')
    issue = await createIssue(request, project.id, 'Rich editor target', 'todo')
    await setToken(page)
    await context.grantPermissions(['clipboard-read', 'clipboard-write'])
    await page.goto(`/projects/${project.identifier}/issues/${issue.identifier}`)

    await page.getByRole('button', { name: 'Edit', exact: true }).click()
    const editor = page.locator('[data-testid="markdown-editor"] .cm-content').first()
    await editor.fill('# Heading\n\nSome **bold** text')

    const headingLine = page
      .locator('[data-testid="markdown-editor"] .cm-line')
      .filter({ hasText: 'Heading' })
      .first()
    await expect(headingLine).toHaveClass(/cm-md-heading-1/)
    await expect(
      page.locator('[data-testid="markdown-editor"] .cm-md-strong').first(),
    ).toBeVisible()

    await page.getByRole('tab', { name: 'Preview' }).first().click()
    const preview = page.locator('[aria-label="Markdown preview"]').first()
    await expect(preview.locator('h2')).toHaveText('Heading')
    await expect(preview.locator('strong')).toHaveText('bold')

    await page.getByRole('tab', { name: 'Write' }).first().click()
    await page.getByRole('button', { name: 'Link' }).first().click()
    const linkDialog = page.getByRole('dialog', { name: 'Insert link' })
    await linkDialog.getByLabel('URL').fill('https://example.com')
    await linkDialog.getByRole('button', { name: 'Insert link' }).click()
    await page.keyboard.type(' after link')
    await expect(editor).toContainText('after link')

    await page.getByRole('button', { name: 'Code block' }).first().click()
    const dialog = page.getByRole('dialog', { name: 'Insert code block' })
    await dialog.getByLabel('Language').fill('typescript')
    await dialog.getByRole('button', { name: 'Insert code block' }).click()
    await page.keyboard.type('const x = 1')

    await page.getByRole('tab', { name: 'Preview' }).first().click()
    const codeBlock = preview.locator('.markdown-codeblock')
    await expect(codeBlock.locator('.markdown-codeblock-lang')).toHaveText('typescript')
    await expect(codeBlock.locator('code')).toContainText('const x = 1')

    await codeBlock.getByRole('button', { name: 'Copy typescript block' }).click()
    await expect(codeBlock.getByRole('button', { name: 'Copy typescript block' })).toHaveText(
      'Copied',
    )
    const clipboard = await page.evaluate(() => navigator.clipboard.readText())
    expect(clipboard).toBe('const x = 1')
  } finally {
    if (project) await deleteProject(request, project.id)
  }
})
