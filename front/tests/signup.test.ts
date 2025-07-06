import { expect, test } from '@playwright/test'

test('Signup page redirects to Zitadel', async ({ page }) => {
  await page.goto('http://localtest.me/signup')

  const button = page.getByRole('button', { name: /sign up with zitadel/i })
  await expect(button).toBeVisible()
})
