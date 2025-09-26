import { expect, test } from '@playwright/test'

test.describe('Login Form Validation', () => {
  test.beforeEach(async ({ page }) => {
    // Set up console and request logging
    page.on('console', (msg) => console.log('BROWSER:', msg.type(), msg.text()))
    page.on('requestfailed', (req) => console.log('REQ FAILED:', req.method(), req.url(), req.failure()?.errorText))
    page.on('response', async (res) => {
      if (res.status() >= 400) {
        console.log('HTTP ERR:', res.status(), res.url())
      }
    })

    // Navigate to login page
    await page.goto('/login')
    await expect(page.getByLabel('Email')).toBeVisible()
  })

  test('shows client-side validation errors for invalid email format', async ({ page }) => {
    // Fill invalid email and valid password
    await page.getByLabel('Email').fill('invalid-email')
    await page.getByLabel('Password').fill('password123')
    
    // Try to submit
    await page.getByRole('button', { name: /^Login$/i }).click()
    
    // Should show email validation error
    await expect(page.getByText('Invalid email address')).toBeVisible()
  })

  test('shows client-side validation errors for missing required fields', async ({ page }) => {
    // Try to submit without filling fields
    await page.getByRole('button', { name: /^Login$/i }).click()
    
    // Should show required field errors
    await expect(page.getByText('Email is required')).toBeVisible()
    await expect(page.getByText('Password is required')).toBeVisible()
  })

  test('shows Zitadel server-side validation errors under email field', async ({ page }) => {
    // Mock session token endpoint
    await page.route('**/api/v1/user/session', async (route) => {
      await route.fulfill({
        status: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_token: 'test-session-token',
          zitadel_url: 'http://auth.localtest.me',
          expires_in: 3600,
        }),
      })
    })

    // Mock Zitadel session creation to return user not found error
    await page.route(/\/v2\/sessions$/, async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 404,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            code: 'NOT_FOUND',
            message: 'User not found',
            details: [{
              '@type': 'type.googleapis.com/google.rpc.BadRequest',
              violations: [{
                field: 'loginName',
                description: 'User not found'
              }]
            }]
          }),
        })
        return
      }
      await route.fallback()
    })

    // Fill valid email format but user doesn't exist
    await page.getByLabel('Email').fill('nonexistent@example.com')
    await page.getByLabel('Password').fill('password123')
    
    // Submit form
    await page.getByRole('button', { name: /^Login$/i }).click()
    
    // Wait for request to complete and check for field-specific error
    await page.waitForRequest((req) => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()))
    
    // Should show Zitadel error under email field
    await expect(page.getByText('User not found')).toBeVisible()
    
    // The email field should be marked as invalid
    const emailField = page.getByLabel('Email')
    await expect(emailField).toHaveAttribute('aria-invalid', 'true')
  })

  test('shows Zitadel authentication errors for wrong password', async ({ page }) => {
    // Mock session token endpoint
    await page.route('**/api/v1/user/session', async (route) => {
      await route.fulfill({
        status: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_token: 'test-session-token',
          zitadel_url: 'http://auth.localtest.me',
          expires_in: 3600,
        }),
      })
    })

    // Mock successful user session creation
    await page.route(/\/v2\/sessions$/, async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            sessionId: 'session-123',
            sessionToken: 'session-token-123',
            changeDate: new Date().toISOString(),
          }),
        })
        return
      }
      await route.fallback()
    })

    // Mock password verification failure
    await page.route(/\/v2\/sessions\/.*/, async (route) => {
      if (route.request().method() === 'PATCH') {
        await route.fulfill({
          status: 412,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            code: 'PRECONDITION_FAILED',
            message: 'Invalid email or password',
            details: [{
              '@type': 'type.googleapis.com/google.rpc.BadRequest',
              violations: [{
                field: 'password',
                description: 'Invalid password'
              }]
            }]
          }),
        })
        return
      }
      await route.fallback()
    })

    // Fill valid credentials but wrong password
    await page.getByLabel('Email').fill('user@example.com')
    await page.getByLabel('Password').fill('wrongpassword')
    
    // Submit form
    await page.getByRole('button', { name: /^Login$/i }).click()
    
    // Wait for requests to complete
    await page.waitForRequest((req) => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()))
    await page.waitForRequest((req) => req.method() === 'PATCH' && /\/v2\/sessions\//.test(req.url()))
    
    // Should show authentication error under email field (common UX pattern)
    await expect(page.getByText('Invalid email or password')).toBeVisible()
    
    // The email field should be marked as invalid
    const emailField = page.getByLabel('Email')
    await expect(emailField).toHaveAttribute('aria-invalid', 'true')
  })

  test('clears field errors when retrying login', async ({ page }) => {
    // Mock session token endpoint
    await page.route('**/api/v1/user/session', async (route) => {
      await route.fulfill({
        status: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_token: 'test-session-token',
          zitadel_url: 'http://auth.localtest.me',
          expires_in: 3600,
        }),
      })
    })

    // First attempt - mock user not found error
    let attemptCount = 0
    await page.route(/\/v2\/sessions$/, async (route) => {
      if (route.request().method() === 'POST') {
        attemptCount++
        if (attemptCount === 1) {
          // First attempt fails
          await route.fulfill({
            status: 404,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              code: 'NOT_FOUND',
              message: 'User not found',
            }),
          })
        } else {
          // Second attempt succeeds
          await route.fulfill({
            status: 200,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              sessionId: 'session-123',
              sessionToken: 'session-token-123',
              changeDate: new Date().toISOString(),
            }),
          })
        }
        return
      }
      await route.fallback()
    })

    // First login attempt
    await page.getByLabel('Email').fill('user@example.com')
    await page.getByLabel('Password').fill('password123')
    await page.getByRole('button', { name: /^Login$/i }).click()
    
    // Wait for error to appear
    await expect(page.getByText('User not found')).toBeVisible()
    
    // Clear email field and try again
    await page.getByLabel('Email').clear()
    await page.getByLabel('Email').fill('different@example.com')
    await page.getByRole('button', { name: /^Login$/i }).click()
    
    // Error should be cleared
    await expect(page.getByText('User not found')).not.toBeVisible()
  })
})