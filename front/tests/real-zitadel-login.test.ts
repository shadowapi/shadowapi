/* eslint-disable no-console */
import { test, expect } from '@playwright/test'

/**
 * This test runs against a real Zitadel instance without mocking.
 *
 * Prerequisites:
 * 1. Docker compose stack must be running: `docker compose watch`
 * 2. Test user must exist in Zitadel: admin@example.com / Admin123!
 *
 * Run with: npx playwright test real-zitadel-login.test.ts --headed
 */
const runReal = process.env.RUN_REAL_ZITADEL === '1'
  ; (runReal ? test.describe : test.describe.skip)('Real Zitadel Authentication', () => {
    test.beforeEach(async ({ page }) => {
      // Enable detailed logging
      page.on('console', (msg) => {
        const type = msg.type()
        const text = msg.text()
        console.log(`[BROWSER ${type.toUpperCase()}]`, text)
      })

      page.on('request', (req) => {
        if (
          req.url().includes('/api/') ||
          req.url().includes('zitadel') ||
          req.url().includes('sessions') ||
          req.url().includes('oauth')
        ) {
          console.log(`[REQUEST] ${req.method()} ${req.url()}`)
          const headers = req.headers()
          if (headers.authorization) {
            console.log(`  Authorization: ${headers.authorization.substring(0, 30)}...`)
          }
          if (req.postData()) {
            try {
              const data = JSON.parse(req.postData() || '{}')
              console.log('  Body:', JSON.stringify(data, null, 2))
            } catch {
              console.log('  Body:', req.postData())
            }
          }
        }
      })

      page.on('response', async (res) => {
        if (
          res.url().includes('/api/') ||
          res.url().includes('zitadel') ||
          res.url().includes('sessions') ||
          res.url().includes('oauth')
        ) {
          console.log(`[RESPONSE] ${res.status()} ${res.url()}`)
          if (res.status() >= 400) {
            try {
              const body = await res.text()
              console.log('  Error body:', body.substring(0, 500))
            } catch {
              console.log('  Could not read response body')
            }
          } else {
            try {
              const body = await res.json()
              console.log('  Response:', JSON.stringify(body, null, 2).substring(0, 300))
            } catch {
              // Non-JSON response
            }
          }
        }
      })

      page.on('requestfailed', (req) => {
        console.log(`[REQUEST FAILED] ${req.method()} ${req.url()}`)
        console.log(`  Failure: ${req.failure()?.errorText}`)
      })

      page.on('pageerror', (err) => {
        console.log('[PAGE ERROR]', err.message)
        console.log(err.stack)
      })
    })

    test('complete login flow with real Zitadel', async ({ page }) => {
      console.log('\n=== TEST START: Real Zitadel Login ===\n')

      // Clear any existing auth state
      await page.goto('http://localtest.me')
      await page.evaluate(() => {
        sessionStorage.clear()
        localStorage.clear()
      })

      // Step 1: Navigate to login page (auto-redirect)
      console.log('Step 1: Navigate to login page')
      await page.goto('http://localtest.me/login?_t=' + Date.now())
      console.log('✓ Login page loaded (auto-redirect flow)')

      // Step 3: Set up request watchers BEFORE clicking submit (to avoid race conditions)
      console.log('\nStep 3: Setting up request watchers')
      const sessionRequestPromise = page.waitForRequest(
        (req) => req.method() === 'POST' && req.url().includes('/api/v1/user/session'),
        { timeout: 10000 },
      )
      const zitadelSessionPromise = page.waitForRequest(
        (req) => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()),
        { timeout: 10000 },
      )
      const passwordVerifyPromise = page.waitForRequest(
        (req) => req.method() === 'PATCH' && /\/v2\/sessions\//.test(req.url()),
        { timeout: 10000 },
      )
      const finalizePromise = page.waitForRequest(
        (req) => req.method() === 'POST' && /\/v2\/oidc\/auth_requests\//.test(req.url()),
        { timeout: 10000 },
      )
      const tokenExchangePromise = page.waitForRequest(
        (req) => req.method() === 'POST' && /\/oauth\/v2\/token$/.test(req.url()),
        { timeout: 10000 },
      )

      // Step 4: Proceed with flow (no local form submit in new flow)
      console.log('\nStep 4: Continue flow (no local form)')

      // Step 5: Wait for each authentication step
      console.log('\nStep 5: Waiting for authentication flow...')

      try {
        console.log('  - Waiting for backend session token request...')
        await sessionRequestPromise
        console.log('  ✓ Backend session token requested')

        console.log('  - Waiting for Zitadel session creation...')
        await zitadelSessionPromise
        console.log('  ✓ Zitadel session created')

        console.log('  - Waiting for password verification...')
        await passwordVerifyPromise
        console.log('  ✓ Password verified')

        console.log('  - Waiting for auth request finalization...')
        await finalizePromise
        console.log('  ✓ Auth request finalized')

        console.log('  - Waiting for PKCE token exchange...')
        await tokenExchangePromise
        console.log('  ✓ PKCE exchange completed')
      } catch (error) {
        console.error('\n❌ Authentication flow failed at some step')
        console.error(error)
        throw error
      }

      // Step 6: Verify authentication success
      console.log('\nStep 6: Verify authentication success')

      // Check that we were redirected to home page
      await expect
        .poll(async () => page.url(), { message: 'Should redirect to home page', timeout: 5000 })
        .toBe('http://localtest.me/')
      console.log('✓ Redirected to home page')

      // Check that tokens are stored in sessionStorage
      const authData = await page.evaluate(() => {
        const data = sessionStorage.getItem('shadowapi_auth')
        return data ? JSON.parse(data) : null
      })

      expect(authData, 'Auth data should be stored').not.toBeNull()
      expect(authData.email, 'Email should match').toBe('admin@example.com')
      expect(authData.accessToken, 'Should have access token').toBeTruthy()
      expect(authData.idToken, 'Should have ID token').toBeTruthy()
      expect(authData.expiresAt, 'Should have expiry time').toBeGreaterThan(Date.now())

      console.log('✓ Auth data stored correctly:', {
        email: authData.email,
        hasAccessToken: !!authData.accessToken,
        hasIdToken: !!authData.idToken,
        hasRefreshToken: !!authData.refreshToken,
        expiresIn: Math.floor((authData.expiresAt - Date.now()) / 1000) + 's',
      })

      // Step 7: Verify we can access protected routes
      console.log('\nStep 7: Verify access to protected routes')
      await page.goto('http://localtest.me/users')
      await expect(page).toHaveURL('http://localtest.me/users')
      await expect(page.getByRole('heading', { name: 'ShadowAPI' })).toBeVisible()
      console.log('✓ Protected route accessible')

      console.log('\n=== TEST PASSED ===\n')
    })

    test('shows proper error for invalid credentials', async ({ page }) => {
      console.log('\n=== TEST START: Invalid Credentials ===\n')

      // Clear any existing auth state
      await page.goto('http://localtest.me')
      await page.evaluate(() => {
        sessionStorage.clear()
        localStorage.clear()
      })

      await page.goto('http://localtest.me/login')

      await page.getByLabel('Email').fill('nonexistent@example.com')
      await page.getByLabel('Password').fill('WrongPassword123!')

      await page.getByRole('button', { name: /^Login$/i }).click()

      // Should show error message (checking for the actual error text from Zitadel)
      await expect(page.getByText(/User could not be found|Invalid email or password/i)).toBeVisible({ timeout: 5000 })
      console.log('✓ Error message displayed for invalid credentials')

      console.log('\n=== TEST PASSED ===\n')
    })

    test('can access users page after login', async ({ page }) => {
      console.log('\n=== TEST START: Users Page Access ===\n')

      // Clear any existing auth state
      await page.goto('http://localtest.me')
      await page.evaluate(() => {
        sessionStorage.clear()
        localStorage.clear()
      })

      // Step 1: Login
      console.log('Step 1: Login')
      await page.goto('http://localtest.me/login?_t=' + Date.now())
      await page.getByLabel('Email').fill('admin@example.com')
      await page.getByLabel('Password').fill('Admin123!')

      const sessionRequestPromise = page.waitForRequest(
        (req) => req.method() === 'POST' && req.url().includes('/api/v1/user/session'),
        { timeout: 10000 },
      )

      await page.getByRole('button', { name: /^Login$/i }).click()
      await sessionRequestPromise

      // Wait for redirect to home
      await expect
        .poll(async () => page.url(), { message: 'Should redirect to home page', timeout: 5000 })
        .toBe('http://localtest.me/')
      console.log('✓ Login successful')

      // Step 2: Navigate to users page
      console.log('\nStep 2: Navigate to users page')
      const usersResponsePromise = page.waitForResponse(
        (res) => res.url().includes('/api/v1/user') && res.request().method() === 'GET',
        { timeout: 10000 },
      )

      await page.goto('http://localtest.me/users')

      // Wait for the API response
      console.log('  - Waiting for users API response...')
      const usersResponse = await usersResponsePromise
      console.log('  ✓ Users API response received:', usersResponse.status())

      // Step 3: Verify response was successful
      console.log('\nStep 3: Verify API response')
      expect(usersResponse.status()).toBe(200)
      const users = await usersResponse.json()
      console.log('✓ Users API returned successfully, got', users.length || 0, 'users')

      // Step 4: Verify page loaded
      console.log('\nStep 4: Verify users page loaded')
      await expect(page).toHaveURL('http://localtest.me/users')
      await expect(page.getByRole('heading', { name: 'ShadowAPI' })).toBeVisible()
      console.log('✓ Users page loaded')

      // Step 5: Check auth token was sent
      console.log('\nStep 5: Verify authentication token was sent')
      const authData = await page.evaluate(() => {
        const data = sessionStorage.getItem('shadowapi_auth')
        return data ? JSON.parse(data) : null
      })
      expect(authData.accessToken, 'Should have access token').toBeTruthy()
      console.log('✓ Access token present:', authData.accessToken.substring(0, 30) + '...')

      console.log('\n=== TEST PASSED ===\n')
    })
  })
