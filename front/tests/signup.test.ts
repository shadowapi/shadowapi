import { expect, test } from '@playwright/test'

test('Signup page is accessible to guests and shows form', async ({ page }) => {
  await page.goto('http://localtest.me/signup')
  await expect(page.getByLabel('First Name')).toBeVisible()
  await expect(page.getByRole('button', { name: /^Sign Up$/i })).toBeVisible()
})
