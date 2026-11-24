import { expect, test } from '@playwright/test'
import { TOKEN } from './helpers'

test('unlocks system status with the API token', async ({ page }) => {
  await page.goto('/settings')
  await expect(page.getByText(/requires an API token/i)).toBeVisible()

  const input = page.getByLabel('Bearer token')
  await input.fill(TOKEN)
  await expect(input).toHaveValue(TOKEN)
  await page.getByRole('button', { name: 'Save token' }).click()
  await expect
    .poll(() => page.evaluate(() => sessionStorage.getItem('tablon.apiToken')))
    .toBe(TOKEN)

  await expect(page.getByText('Enabled')).toBeVisible()
})
