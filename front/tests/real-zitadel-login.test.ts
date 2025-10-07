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
test.describe('Real Zitadel Authentication', () => {
  test.beforeEach(async ({ page }) => {
    // Enable detailed logging
    page.on('console', msg => {
      const type = msg.type()
      const text = msg.text()
      console.log(`[BROWSER ${type.toUpperCase()}]`, text)
    })

    page.on('request', req => {
      if (req.url().includes('/api/') || req.url().includes('zitadel') || req.url().includes('sessions') || req.url().includes('oauth')) {
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

    page.on('response', async res => {
      if (res.url().includes('/api/') || res.url().includes('zitadel') || res.url().includes('sessions') || res.url().includes('oauth')) {
        console.log(`[RESPONSE] ${res.status()} ${res.url()}`)
        if (res.status() >= 400) {
          try {
            const body = await res.text()
            console.log('  Error body:', body.substring(0, 500))
          } catch (e) {
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

    page.on('requestfailed', req => {
      console.log(`[REQUEST FAILED] ${req.method()} ${req.url()}`)
      console.log(`  Failure: ${req.failure()?.errorText}`)
    })

    page.on('pageerror', err => {
      console.log('[PAGE ERROR]', err.message)
      console.log(err.stack)
    })
  })

  test('complete login flow with real Zitadel', async ({ page }) => {
    console.log('\n=== TEST START: Real Zitadel Login ===\n')

    // Step 1: Navigate to login page
    console.log('Step 1: Navigate to login page')
    await page.goto('http://localtest.me/login')
    await expect(page.getByLabel('Email')).toBeVisible()
    console.log('✓ Login page loaded')

    // Step 2: Fill in credentials
    console.log('\nStep 2: Fill in credentials')
    await page.getByLabel('Email').fill('admin@example.com')
    await page.getByLabel('Password').fill('Admin123!')
    console.log('✓ Credentials filled')

    // Step 3: Submit form and wait for authentication flow
    console.log('\nStep 3: Submit form')
    await page.getByRole('button', { name: /^Login$/i }).click()

    // Step 4: Wait for each authentication step
    console.log('\nStep 4: Waiting for authentication flow...')

    try {
      // Wait for backend session token request
      console.log('  - Waiting for backend session token request...')
      await page.waitForRequest(
        req => req.method() === 'POST' && req.url().includes('/api/v1/user/session'),
        { timeout: 10000 }
      )
      console.log('  ✓ Backend session token requested')

      // Wait for Zitadel session creation
      console.log('  - Waiting for Zitadel session creation...')
      await page.waitForRequest(
        req => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()),
        { timeout: 10000 }
      )
      console.log('  ✓ Zitadel session created')

      // Wait for password verification
      console.log('  - Waiting for password verification...')
      await page.waitForRequest(
        req => req.method() === 'PATCH' && /\/v2\/sessions\//.test(req.url()),
        { timeout: 10000 }
      )
      console.log('  ✓ Password verified')

      // Wait for token exchange
      console.log('  - Waiting for token exchange...')
      await page.waitForRequest(
        req => req.method() === 'POST' && /\/oauth\/v2\/token$/.test(req.url()),
        { timeout: 10000 }
      )
      console.log('  ✓ Tokens exchanged')
    } catch (error) {
      console.error('\n❌ Authentication flow failed at some step')
      console.error(error)
      throw error
    }

    // Step 5: Verify authentication success
    console.log('\nStep 5: Verify authentication success')

    // Check that we were redirected to home page
    await expect.poll(
      async () => page.url(),
      { message: 'Should redirect to home page', timeout: 5000 }
    ).toBe('http://localtest.me/')
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
      expiresIn: Math.floor((authData.expiresAt - Date.now()) / 1000) + 's'
    })

    // Step 6: Verify we can access protected routes
    console.log('\nStep 6: Verify access to protected routes')
    await page.goto('http://localtest.me/users')
    await expect(page).toHaveURL('http://localtest.me/users')
    await expect(page.getByRole('heading', { name: 'ShadowAPI' })).toBeVisible()
    console.log('✓ Protected route accessible')

    console.log('\n=== TEST PASSED ===\n')
  })

  test('shows proper error for invalid credentials', async ({ page }) => {
    console.log('\n=== TEST START: Invalid Credentials ===\n')

    await page.goto('http://localtest.me/login')
    await expect(page.getByLabel('Email')).toBeVisible()

    await page.getByLabel('Email').fill('nonexistent@example.com')
    await page.getByLabel('Password').fill('WrongPassword123!')

    await page.getByRole('button', { name: /^Login$/i }).click()

    // Should show error message
    await expect(page.getByText(/User not found|Invalid email or password/i)).toBeVisible({ timeout: 10000 })
    console.log('✓ Error message displayed for invalid credentials')

    console.log('\n=== TEST PASSED ===\n')
  })
})
