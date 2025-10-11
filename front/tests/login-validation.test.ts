/* eslint-disable no-console */
import { expect, test, type Page } from '@playwright/test'

const pkceStorageKey = 'shadowapi_zitadel_pkce'
const zitadelE2EEnabled = !!process.env.ZITADEL_E2E

async function openLoginWithAuthRequest(page: Page) {
  const pkceState = 'test-state'
  await page.addInitScript(
    ([storageKey, state]) => {
      window.sessionStorage.setItem(
        storageKey,
        JSON.stringify({
          codeVerifier: 'test-verifier',
          createdAt: Date.now(),
          state,
          returnTo: '/',
          codeChallengeMethod: 'plain',
        }),
      )
    },
    [pkceStorageKey, pkceState],
  )
  await page.goto(`/login?authRequest=test-auth-request&state=${pkceState}`)
  await expect(page.getByLabel('Email')).toBeVisible()
}

test.describe('Login form validation', () => {
  test('shows client-side validation errors for invalid email format', async ({ page }) => {
    await openLoginWithAuthRequest(page)

    await page.getByLabel('Email').fill('invalid-email')
    await page.getByLabel('Password').fill('password123')
    await page.getByRole('button', { name: /^Login$/i }).click()

    await expect(page.getByText('Invalid email address')).toBeVisible()
  })

  test('requires both fields before submitting', async ({ page }) => {
    await openLoginWithAuthRequest(page)

    await page.getByRole('button', { name: /^Login$/i }).click()

    await expect(page.getByText('Email is required')).toBeVisible()
    await expect(page.getByText('Password is required')).toBeVisible()
  })
})

test.describe('Real Zitadel validation (no mocks)', () => {
  test.skip(!zitadelE2EEnabled, 'Requires real Zitadel instance (set ZITADEL_E2E=1)')

  test('shows authentication error for unknown user', async ({ page }) => {
    page.on('console', (msg) => console.log('[BROWSER]', msg.type(), msg.text()))
    await page.goto('/login')
    await page.waitForURL(/\/login.*authRequest=/, { timeout: 20000 })

    await page.getByLabel('Email').fill(`unknown+${Date.now()}@example.com`)
    await page.getByLabel('Password').fill('wrongpassword123!')

    await page.getByRole('button', { name: /^Login$/i }).click()

    await expect(page.getByText(/User could not be found|Invalid email or password/i)).toBeVisible({ timeout: 10000 })
  })
})
