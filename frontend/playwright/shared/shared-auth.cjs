/**
 * Shared authentication module for ShadowAPI Playwright tests
 * Handles login flow: POST /login with {email, password} → cookie session
 */

const { USERS, ROLES, SHARED_PASSWORD, BASE_URL, getAdmin } = require('./users.cjs')

const CONFIG = {
  baseUrl: BASE_URL,
  timeout: 15000,
  users: USERS,
}

/**
 * Authenticate user on ShadowAPI
 * @param {import('playwright').Page} page
 * @param {Object} options
 * @param {string} [options.email]
 * @param {string} [options.password]
 * @param {string} [options.baseUrl]
 * @param {boolean} [options.verbose=true]
 * @returns {Promise<boolean>}
 */
async function authenticateUser(page, options = {}) {
  const {
    email = USERS.admin.email,
    password = USERS.admin.password,
    baseUrl = CONFIG.baseUrl,
    verbose = true,
  } = options

  const base = baseUrl.replace(/\/$/, '')
  const loginUrl = `${base}/login`
  const log = verbose ? console.log : () => {}

  try {
    log(`Navigating to login page: ${loginUrl}`)
    await page.goto(loginUrl, { waitUntil: 'networkidle', timeout: 20000 })
  } catch (error) {
    log(`Failed to load login page: ${error.message}`)
    return false
  }

  // Check if already authenticated (redirected away from login)
  if (!page.url().includes('/login')) {
    log('Already authenticated - redirected from login')
    return true
  }

  // Fill email
  const emailInput = page.locator('input[type="email"]').first()
  const emailVisible = await emailInput.isVisible({ timeout: 5000 }).catch(() => false)

  if (!emailVisible) {
    log('Email field not visible')
    return false
  }

  await emailInput.fill(email)
  log('Filled email')

  // Fill password
  const passwordInput = page.locator('input[type="password"]').first()
  const passwordVisible = await passwordInput.isVisible({ timeout: 5000 }).catch(() => false)

  if (!passwordVisible) {
    log('Password field not visible')
    return false
  }

  await passwordInput.fill(password)
  log('Filled password')

  // Click Login button
  const loginButton = page.locator('button:has-text("Login")').first()
  if (!(await loginButton.isVisible({ timeout: 5000 }).catch(() => false))) {
    log('Login button not visible')
    return false
  }

  await loginButton.click()
  log('Clicked login button')

  // Wait for redirect away from /login
  await page.waitForTimeout(3000)

  try {
    await page.waitForFunction(
      () => !window.location.href.includes('/login'),
      { timeout: 10000 }
    )
    log('Login redirect detected')
  } catch {
    log('No redirect detected after login')
  }

  const currentUrl = page.url()
  const authenticated = !currentUrl.includes('/login')

  log(authenticated ? 'Authentication successful' : 'Authentication failed')
  return authenticated
}

module.exports = { authenticateUser, CONFIG, USERS, ROLES, SHARED_PASSWORD, getAdmin }
