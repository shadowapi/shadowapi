/* eslint-disable no-console */
import { expect, test } from '@playwright/test'

const zitadelE2EEnabled = !!process.env.ZITADEL_E2E

test.describe('End-to-end auth without Zitadel mocks', () => {
  test.skip(!zitadelE2EEnabled, 'Requires real Zitadel instance (set ZITADEL_E2E=1)')

  test('redirects to login, completes PKCE flow, and lands on requested route', async ({ page }) => {
    const email = process.env.ZITADEL_E2E_USER ?? 'admin@example.com'
    const password = process.env.ZITADEL_E2E_PASSWORD ?? 'Admin123!'

    page.on('console', (msg) => console.log('[BROWSER]', msg.type(), msg.text()))
    page.on('requestfailed', (req) => console.log('[REQ FAILED]', req.method(), req.url(), req.failure()?.errorText))

    await test.step('Navigate to protected page and wait for login with authRequest', async () => {
      await page.goto('/users')
      await page.waitForURL(/\/login\?returnTo=%2Fusers/, { timeout: 20000 })
      await page.waitForURL(/\/login.*authRequest=/, { timeout: 20000 })
      await expect(page.getByLabel('Email')).toBeVisible()
    })

    await test.step('Submit credentials and wait for Zitadel requests', async () => {
      await page.getByLabel('Email').fill(email)
      await page.getByLabel('Password').fill(password)

      const sessionReq = page.waitForRequest(
        (req) => req.method() === 'POST' && req.url().includes('/api/v1/user/session'),
      )
      const zitadelSessionReq = page.waitForRequest(
        (req) => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()),
      )
      const passwordVerifyReq = page.waitForRequest(
        (req) => req.method() === 'PATCH' && /\/v2\/sessions\//.test(req.url()),
      )
      const finalizeReq = page.waitForRequest(
        (req) => req.method() === 'POST' && /\/v2\/oidc\/auth_requests\//.test(req.url()),
      )
      const tokenReq = page.waitForRequest((req) => req.method() === 'POST' && /\/oauth\/v2\/token$/.test(req.url()))

      await page.getByRole('button', { name: /^Login$/i }).click()

      await Promise.all([sessionReq, zitadelSessionReq, passwordVerifyReq, finalizeReq, tokenReq])
    })

    await test.step('Verify redirect and stored tokens', async () => {
      await page.waitForURL('**/users', { timeout: 20000 })

      const authData = await page.evaluate(() => {
        const stored = sessionStorage.getItem('shadowapi_auth')
        return stored ? JSON.parse(stored) : null
      })

      expect(authData).not.toBeNull()
      expect(authData.email).toBe(email)
      expect(authData.accessToken).toBeTruthy()
      expect(authData.idToken).toBeTruthy()
    })
  })
})
