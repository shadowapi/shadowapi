import { test, expect } from '@playwright/test'

test.describe('Real Zitadel Authentication Flow', () => {
  test('user can log in with real Zitadel', async ({ page }) => {
    // Enable console logging
    page.on('console', msg => console.log('BROWSER:', msg.type(), msg.text()))

    // Log all navigation
    page.on('framenavigated', frame => {
      if (frame === page.mainFrame()) {
        console.log('NAVIGATED TO:', frame.url())
      }
    })

    // Go to login page
    await page.goto('http://localtest.me/login')
    console.log('Initial URL:', page.url())

    // Wait for the page to load
    await page.waitForLoadState('networkidle')

    // Fill in credentials
    console.log('Filling credentials...')
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'Admin123!')

    // Submit the form
    console.log('Submitting form...')
    await page.click('button[type="submit"]')

    // Wait a bit to see what happens
    await page.waitForTimeout(2000)
    console.log('After submit, URL:', page.url())

    // Wait for redirect to Zitadel authorize endpoint
    // This should happen automatically after form submission
    await page.waitForURL(/auth\.localtest\.me\/oauth\/v2\/authorize/, { timeout: 10000 })

    console.log('Redirected to Zitadel authorize endpoint')
    console.log('Current URL:', page.url())

    // The page should redirect back to our login page with authRequest parameter
    await page.waitForURL(/localtest\.me\/login\?authRequest=/, { timeout: 15000 })

    console.log('Redirected back to login with authRequest')
    console.log('Current URL:', page.url())

    // The form should auto-submit and complete the authentication
    // Wait for final redirect to home page or dashboard
    await page.waitForURL(/localtest\.me\/$|localtest\.me\/messages/, { timeout: 15000 })

    console.log('Successfully authenticated, redirected to:', page.url())

    // Verify we have tokens in sessionStorage
    const authData = await page.evaluate(() => {
      const data = sessionStorage.getItem('auth_user')
      return data ? JSON.parse(data) : null
    })

    expect(authData).not.toBeNull()
    expect(authData.email).toBe('admin@example.com')
    expect(authData.accessToken).toBeTruthy()
    expect(authData.idToken).toBeTruthy()

    console.log('Auth data verified:', {
      email: authData.email,
      hasAccessToken: !!authData.accessToken,
      hasIdToken: !!authData.idToken,
      hasRefreshToken: !!authData.refreshToken
    })
  })
})
