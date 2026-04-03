/**
 * Test 01: Login Flow
 * Tests ShadowAPI login page, authentication, and dashboard access
 *
 * Run: node front/tests/test-01-login.cjs
 * Or:  make test-login
 */

const { chromium } = require('playwright')
const { authenticateUser, CONFIG, USERS } = require('./shared/shared-auth.cjs')

const results = {
  passed: 0,
  failed: 0,
  skipped: 0,
  tests: [],
}

function recordTest(name, passed, details = '') {
  results.tests.push({ name, passed, details })
  if (passed) {
    results.passed++
    console.log(`   PASS ${name}`)
  } else {
    results.failed++
    console.log(`   FAIL ${name}${details ? `: ${details}` : ''}`)
  }
}

function skipTest(name, reason) {
  results.tests.push({ name, passed: null, details: reason })
  results.skipped++
  console.log(`   SKIP ${name}: ${reason}`)
}

async function testLogin() {
  console.log('Testing ShadowAPI Login Flow...\n')

  const browser = await chromium.launch({ headless: false })
  const context = await browser.newContext({ ignoreHTTPSErrors: true })
  const page = await context.newPage()

  // Track API responses
  const apiResponses = {}

  page.on('response', async (response) => {
    const url = response.url()
    try {
      const status = response.status()
      if (url.includes('/login') && response.request().method() === 'POST') {
        apiResponses.login = { status }
      }
      if (url.includes('/api/v1/session')) {
        apiResponses.session = { status, data: status === 200 ? await response.json().catch(() => null) : null }
      }
    } catch (e) {
      // ignore
    }
  })

  page.on('console', (msg) => {
    const type = msg.type()
    if (type === 'error' || type === 'warning') {
      console.log(`[Browser ${type.toUpperCase()}]`, msg.text())
    }
  })

  page.on('pageerror', (error) => {
    console.error('[Page Error]', error.message)
  })

  try {
    // ================================================================
    // PHASE 1: Login Page Loads
    // ================================================================
    console.log('1. Login Page...')

    await page.goto(`${CONFIG.baseUrl}/login`, { waitUntil: 'networkidle', timeout: 20000 })
    await page.waitForTimeout(2000)

    // Test 1.1: Page loads
    const loginPageLoaded = page.url().includes('/login') || page.url() === CONFIG.baseUrl + '/'
    recordTest('Login page loads', loginPageLoaded, `URL: ${page.url()}`)

    // Test 1.2: Email field present
    const hasEmail = (await page.locator('input[type="email"]').count()) > 0
    recordTest('Email field present', hasEmail)

    // Test 1.3: Password field present
    const hasPassword = (await page.locator('input[type="password"]').count()) > 0
    recordTest('Password field present', hasPassword)

    // Test 1.4: Login button present
    const hasLoginButton = (await page.locator('button:has-text("Login")').count()) > 0
    recordTest('Login button present', hasLoginButton)

    console.log('')

    // ================================================================
    // PHASE 2: Invalid Login
    // ================================================================
    console.log('2. Invalid Login Attempt...')

    await page.locator('input[type="email"]').fill('invalid@test.com')
    await page.locator('input[type="password"]').fill('wrongpassword')
    await page.locator('button:has-text("Login")').first().click()
    await page.waitForTimeout(3000)

    // Test 2.1: Should stay on login page
    const stayedOnLogin = page.url().includes('/login')
    recordTest('Invalid login stays on login page', stayedOnLogin, `URL: ${page.url()}`)

    // Test 2.2: Error message shown
    const hasErrorMsg = (await page.locator('text=Invalid email or password').count()) > 0
    recordTest('Error message displayed', hasErrorMsg)

    console.log('')

    // ================================================================
    // PHASE 3: Valid Login
    // ================================================================
    console.log('3. Valid Login...')

    if (!USERS.admin.email || !USERS.admin.password) {
      skipTest('Valid login', 'No admin credentials configured in shared/users.js')
      skipTest('Dashboard redirect', 'Depends on valid login')
      skipTest('Session active', 'Depends on valid login')
    } else {
      // Clear inputs and try valid login
      await page.goto(`${CONFIG.baseUrl}/login`, { waitUntil: 'networkidle', timeout: 20000 })
      await page.waitForTimeout(1000)

      const authenticated = await authenticateUser(page, {
        email: USERS.admin.email,
        password: USERS.admin.password,
      })

      recordTest('Valid login authenticates', authenticated)

      // Test 3.2: Redirected to dashboard
      const onDashboard = !page.url().includes('/login')
      recordTest('Redirected to dashboard', onDashboard, `URL: ${page.url()}`)

      // Test 3.3: Session is active
      if (authenticated) {
        await page.goto(`${CONFIG.baseUrl}/`, { waitUntil: 'networkidle', timeout: 15000 })
        await page.waitForTimeout(2000)
        const stillAuthenticated = !page.url().includes('/login')
        recordTest('Session persists on navigation', stillAuthenticated)
      } else {
        skipTest('Session persists on navigation', 'Login failed')
      }
    }

    console.log('')

    // ================================================================
    // PHASE 4: Protected Routes
    // ================================================================
    console.log('4. Protected Routes...')

    // Clear cookies and open a fresh context to avoid SWR cache
    await browser.close()
    const browser2 = await chromium.launch({ headless: false })
    const context2 = await browser2.newContext({ ignoreHTTPSErrors: true })
    const page2 = await context2.newPage()

    await page2.goto(`${CONFIG.baseUrl}/datasources`, { waitUntil: 'networkidle', timeout: 15000 })
    await page2.waitForTimeout(3000)

    // Session check will fail → should show login or session expired
    const redirectedToLogin = page2.url().includes('/login') || (await page2.locator('text=Session Expired').count()) > 0
    recordTest('Unauthenticated shows login/session expired', redirectedToLogin, `URL: ${page2.url()}`)

    await browser2.close()

    console.log('')
  } catch (error) {
    console.error('Test suite error:', error.message)
    results.failed++
  } finally {
    await browser.close().catch(() => {})
  }

  // Print summary
  console.log(`${'='.repeat(60)}`)
  console.log('TEST SUMMARY')
  console.log('='.repeat(60))
  console.log(`Passed:  ${results.passed}`)
  console.log(`Failed:  ${results.failed}`)
  console.log(`Skipped: ${results.skipped}`)
  console.log(`Total:   ${results.tests.length}`)
  console.log('='.repeat(60))

  if (results.failed > 0) {
    console.log('\nFailed Tests:')
    results.tests
      .filter((t) => t.passed === false)
      .forEach((t) => {
        console.log(`   - ${t.name}${t.details ? `: ${t.details}` : ''}`)
      })
  }

  return results.failed === 0
}

testLogin()
  .then((success) => {
    console.log(success ? '\nAll tests passed!' : '\nSome tests failed')
    process.exit(success ? 0 : 1)
  })
  .catch((error) => {
    console.error('\nTest suite crashed:', error.message)
    process.exit(1)
  })
