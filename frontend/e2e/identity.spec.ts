import { expect, test, type Page } from '@playwright/test'
import {
  API_URL,
  authHeaders,
  createProject,
  deleteProject,
  setToken,
  type TestProject,
} from './helpers'

async function setInputValue(page: Page, label: string, value: string) {
  await page.getByLabel(label).evaluate((element, next) => {
    const input = element as HTMLInputElement
    input.value = next
    input.dispatchEvent(new Event('input', { bubbles: true }))
  }, value)
}

test.describe('project identity', () => {
  let project: TestProject

  test.beforeEach(async ({ request }) => {
    project = await createProject(request, 'Identity project')
  })

  test.afterEach(async ({ request }) => {
    if (project) await deleteProject(request, project.id)
  })

  test('saves, changes and clears a custom gradient through the pickers', async ({
    page,
    request,
  }) => {
    await setToken(page)
    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('button', { name: 'Edit project' }).click()

    await page.locator('label[title="Custom gradient"]').click()
    await expect(page.getByLabel('Gradient start colour')).toBeVisible()
    await setInputValue(page, 'Gradient start colour', '#112233')
    await setInputValue(page, 'Gradient end colour', '#445566')
    await setInputValue(page, 'Angle', '90')

    const trigger = page.getByRole('button', { name: /Choose emoji/ })
    await trigger.click()
    const picker = page.getByRole('dialog', { name: 'Emoji picker' })
    await expect(picker).toBeVisible()

    const triggerBox = await trigger.boundingBox()
    const pickerBox = await picker.boundingBox()
    expect(triggerBox).not.toBeNull()
    expect(pickerBox).not.toBeNull()
    expect(Math.abs((pickerBox?.x ?? 0) - (triggerBox?.x ?? 0))).toBeLessThan(32)

    await page.keyboard.press('Escape')
    await expect(picker).toBeHidden()
    await expect(trigger).toBeFocused()

    await trigger.click()
    await page.getByRole('button', { name: 'Travel' }).click()
    await page.getByRole('button', { name: 'Rocket' }).click()
    await expect(page.getByRole('button', { name: /🚀/ })).toBeFocused()

    await page.getByRole('button', { name: /Change emoji/ }).click()
    await page.getByRole('button', { name: 'Custom emoji' }).click()
    await page.getByLabel('Custom emoji').fill('🧪')
    await page.getByRole('button', { name: 'Use' }).click()
    await expect(page.getByRole('button', { name: /🧪/ })).toBeFocused()

    await page.getByRole('button', { name: 'Save changes' }).click()
    await expect(page.getByRole('dialog', { name: 'Project settings' })).toBeHidden()

    const saved = await request.get(`${API_URL}/api/v1/projects/${project.identifier}`, {
      headers: authHeaders(),
    })
    expect(saved.ok()).toBe(true)
    const body = await saved.json()
    expect(body.logo_gradient_from).toBe('#112233')
    expect(body.logo_gradient_to).toBe('#445566')
    expect(body.logo_gradient_angle).toBe(90)
    expect(body.logo_emoji).toBe('🧪')

    await page.getByRole('button', { name: 'Edit project' }).click()
    await page.locator('label[title="Amber"]').click()
    await page.getByRole('button', { name: 'Save changes' }).click()
    await expect(page.getByRole('dialog', { name: 'Project settings' })).toBeHidden()

    const cleared = await request.get(`${API_URL}/api/v1/projects/${project.identifier}`, {
      headers: authHeaders(),
    })
    expect(cleared.ok()).toBe(true)
    const clearedBody = await cleared.json()
    expect(clearedBody.logo_gradient).toBe(0)
    expect(clearedBody.logo_gradient_from).toBe('')
    expect(clearedBody.logo_gradient_to).toBe('')
    expect(clearedBody.logo_gradient_angle).toBe(135)
    expect(clearedBody.logo_emoji).toBe('🧪')
  })

  test('uses a bottom sheet on mobile', async ({ page }) => {
    await setToken(page)
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto(`/projects/${project.identifier}`)
    await page.getByRole('button', { name: 'Edit project' }).click()
    await page.getByRole('button', { name: /Choose emoji/ }).click()

    const picker = page.getByRole('dialog', { name: 'Emoji picker' })
    await expect(picker).toBeVisible()
    const box = await picker.boundingBox()
    expect(box).not.toBeNull()

    const viewport = page.viewportSize()
    expect(box?.width).toBeCloseTo(viewport?.width ?? 0, 0)
    expect(Math.abs((box?.y ?? 0) + (box?.height ?? 0) - (viewport?.height ?? 0))).toBeLessThan(2)
  })
})
