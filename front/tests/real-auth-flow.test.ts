/* eslint-disable no-console */
import { expect, test } from '@playwright/test'

const runReal = process.env.RUN_REAL_ZITADEL === '1'
  ; (runReal ? test.describe : test.describe.skip)('Real Zitadel Authentication Flow', () => {
    test('user can log in with real Zitadel', async ({ page }) => {
      // Enable console logging
      page.on('console', msg => console.log('BROWSER:', msg.type(), msg.text()))

      test.describe('PKCE bootstrap flow', () => {
        test.skip(!zitadelE2EEnabled, 'Requires real Zitadel instance (set ZITADEL_E2E=1)')

        test('automatically redirects to authorize and stores PKCE verifier', async ({ page }) => {
          page.on('framenavigated', (frame) => {
            if (frame === page.mainFrame()) {
              console.log('[NAV]', frame.url())
            }
          })

          await page.goto('http://localtest.me/login?_ts=' + Date.now())

          await page.waitForURL(/auth\.localtest\.me\/oauth\/v2\/authorize/, { timeout: 15000 })
          await page.waitForURL(/localtest\.me\/login\?authRequest=/, { timeout: 20000 })

          const pkce = await page.evaluate(() => sessionStorage.getItem('shadowapi_zitadel_pkce'))
          expect(pkce).not.toBeNull()

          const parsed = JSON.parse(pkce!)
          expect(parsed.codeVerifier).toBeTruthy()
          expect(parsed.state).toBeTruthy()
        })
      })
