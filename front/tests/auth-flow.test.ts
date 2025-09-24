import { expect, test } from '@playwright/test'

test('guest can sign up, log in, and access protected pages', async ({ page }) => {
  page.on('console', (msg) => console.log('BROWSER:', msg.type(), msg.text()))
  page.on('requestfailed', (req) => console.log('REQ FAILED:', req.method(), req.url(), req.failure()?.errorText))
  page.on('response', async (res) => {
    if (res.status() >= 400) {
      console.log('HTTP ERR:', res.status(), res.url())
    }
  })
  const email = `test.user.${Date.now()}@example.com`
  const password = 'TestPassword123!'

  const usersListResponse = [
    {
      uuid: 'user-123',
      email,
      first_name: 'Test',
      last_name: 'User',
    },
  ]

  await test.step('Redirect unauthenticated user from a protected route', async () => {
    await page.goto('/users')
    await expect(page).toHaveURL(/\/login\?returnTo=%2Fusers$/)
    await expect(page.getByLabel('Email')).toBeVisible()
  })

  await test.step('Allow guests to open the signup page', async () => {
    await page.getByRole('link', { name: /sign up/i }).click()
    await expect(page).toHaveURL(/\/signup$/)
    await expect(page.getByLabel('First Name')).toBeVisible()
    await expect(page.getByRole('button', { name: /^Sign Up$/i })).toBeVisible()
  })

  await page.route('**/api/v1/user', async (route) => {
    const request = route.request()
    const method = request.method()
    console.log('ROUTE /api/v1/user', method, request.url())

    if (method === 'POST') {
      await route.fulfill({
        status: 201,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          uuid: usersListResponse[0].uuid,
          email: usersListResponse[0].email,
          first_name: usersListResponse[0].first_name,
          last_name: usersListResponse[0].last_name,
        }),
      })
      return
    }

    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(usersListResponse),
      })
      return
    }

    console.log('ROUTE /api/v1/user FALLBACK', method)
    await route.fallback()
  })

  await test.step('Submit signup form for a new user', async () => {
    await page.getByLabel('First Name').fill('Test')
    await page.getByLabel('Last Name').fill('User')
    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Password', { exact: true }).fill(password)
    await page.getByLabel('Confirm Password').fill(password)

    await Promise.all([
      page.waitForRequest((req) => req.method() === 'POST' && req.url().endsWith('/api/v1/user')),
      page.getByRole('button', { name: /^Sign Up$/i }).click(),
    ])

    await expect(page).toHaveURL(/\/login\?returnTo=%2F$/)
    await expect(page.getByLabel('Email')).toBeVisible()
  })

  await page.route('**/api/v1/user/session', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*',
        },
        body: JSON.stringify({
          session_token: 'backend-session-token',
          zitadel_url: 'http://auth.localtest.me',
          expires_in: 3600,
        }),
      })
      return
    }

    await route.fallback()
  })

  // Intercept Zitadel session endpoints for both collection and specific-session paths
  await page.route(/\/v2\/sessions(\/.*)?$/, async (route) => {
    const method = route.request().method()

    if (method === 'OPTIONS') {
      await route.fulfill({
        status: 200,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'POST, PATCH, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type, Authorization',
        },
        body: '',
      })
      return
    }

    if (method === 'POST' || method === 'PATCH') {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*',
        },
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

  await test.step('Log in with the newly created user', async () => {
    await page.getByLabel('Email').fill(email)
    await page.getByLabel('Password').fill(password)

    const sessionReq = page.waitForRequest((req) => req.method() === 'POST' && req.url().endsWith('/api/v1/user/session'))
    const zitadelCreateReq = page.waitForRequest((req) => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()))
    const zitadelPatchReq = page.waitForRequest((req) => req.method() === 'PATCH' && /\/v2\/sessions\//.test(req.url()))

    await page.getByRole('button', { name: /^Login$/i }).click()
    await Promise.all([sessionReq, zitadelCreateReq, zitadelPatchReq])

    await expect.poll(async () => {
      return await page.evaluate(() => localStorage.getItem('shadowapi_auth'))
    }, { message: 'Auth not stored in localStorage' }).not.toBeNull()

    await expect(page).toHaveURL('http://localtest.me/')
  })

  await test.step('Access protected content after authentication', async () => {
    await page.goto('/users')
    await expect(page).toHaveURL('http://localtest.me/users')
    await expect(page.getByRole('heading', { name: 'ShadowAPI' })).toBeVisible()
    await expect(page.getByLabel('Email')).not.toBeVisible({ timeout: 1000 })
  })
})
