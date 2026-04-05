/**
 * Test 03: Form Data Loading
 * Tests that edit forms load existing data correctly for all entity types
 *
 * Run: node frontend/playwright/test-03-form-load.cjs
 * Or:  make test-form-load
 *
 * Prerequisites: backend + frontend running, admin user exists, some entities created
 */

const { chromium } = require('playwright')
const { authenticateUser, CONFIG } = require('./shared/shared-auth.cjs')

const results = { passed: 0, failed: 0, skipped: 0, tests: [] }

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

/** Read the value of an antd form field by label */
async function readField(page, label) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const input = item.locator('input').first()
  return input.inputValue()
}

/** Read the selected value text of an antd Select by label */
async function readSelect(page, label) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const sel = item.locator('.ant-select-selection-item').first()
  const exists = await sel.isVisible({ timeout: 3000 }).catch(() => false)
  return exists ? sel.textContent() : null
}

/** Check if a Switch is checked by label */
async function readSwitch(page, label) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const switchEl = item.locator('.ant-switch').first()
  const cls = await switchEl.getAttribute('class')
  return cls?.includes('ant-switch-checked') ?? false
}

/** Click the first edit button in a table row matching text */
async function clickRowEdit(page, text) {
  const row = page.locator(`.ant-table-tbody tr:has(td:has-text("${text}"))`)
  const editBtn = row.locator('button').first()
  await editBtn.click()
  await page.waitForTimeout(2000)
}

/** Get the first table cell text for a given column index */
async function getFirstRowCell(page, colIdx) {
  const cell = page.locator(`.ant-table-tbody tr`).first().locator('td').nth(colIdx)
  return cell.textContent()
}

// ================================================================
// TEST: Users form loads data
// ================================================================
async function testUserFormLoad(page) {
  console.log('\n--- User Edit Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/users`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  if (rows === 0) {
    skipTest('User form load', 'No users in table')
    return
  }

  // Click first user edit
  const firstEditBtn = page.locator('.ant-table-tbody tr').first().locator('button').first()
  await firstEditBtn.click()
  await page.waitForTimeout(2000)

  recordTest('User edit page loads', page.url().includes('/users/'))

  const email = await readField(page, 'Email')
  recordTest('User email field populated', email.length > 0)

  const firstName = await readField(page, 'First Name')
  const lastName = await readField(page, 'Last Name')
  recordTest('User name fields populated', firstName.length > 0 || lastName.length > 0)
}

// ================================================================
// TEST: Storages form loads data
// ================================================================
async function testStorageFormLoad(page) {
  console.log('\n--- Storage Edit Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/storages`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  if (rows === 0) {
    skipTest('Storage form load', 'No storages in table')
    return
  }

  const firstEditBtn = page.locator('.ant-table-tbody tr').first().locator('button').first()
  await firstEditBtn.click()
  await page.waitForTimeout(2000)

  recordTest('Storage edit page loads', page.url().includes('/storages/'))

  const name = await readField(page, 'Name')
  recordTest('Storage name field populated', name.length > 0)

  const typeSelect = await readSelect(page, 'Type')
  recordTest('Storage type field populated', typeSelect !== null && typeSelect.length > 0)
}

// ================================================================
// TEST: DataSource form loads data
// ================================================================
async function testDataSourceFormLoad(page) {
  console.log('\n--- DataSource Edit Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/datasources`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  if (rows === 0) {
    skipTest('DataSource form load', 'No datasources in table')
    return
  }

  const firstEditBtn = page.locator('.ant-table-tbody tr').first().locator('button').first()
  await firstEditBtn.click()
  await page.waitForTimeout(3000) // extra wait for cascaded fetches

  recordTest('DataSource edit page loads', page.url().includes('/datasources/'))

  const name = await readField(page, 'Name')
  recordTest('DataSource name field populated', name.length > 0)

  const typeSelect = await readSelect(page, 'Type')
  recordTest('DataSource type field populated', typeSelect !== null && typeSelect.length > 0)

  const userSelect = await readSelect(page, 'User')
  recordTest('DataSource user field populated', userSelect !== null && userSelect.length > 0)
}

// ================================================================
// TEST: OAuth2 Credentials form loads data
// ================================================================
async function testOAuth2FormLoad(page) {
  console.log('\n--- OAuth2 Credential Edit Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/oauth2/credentials`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  if (rows === 0) {
    skipTest('OAuth2 form load', 'No OAuth2 credentials in table')
    return
  }

  const firstEditBtn = page.locator('.ant-table-tbody tr').first().locator('button').first()
  await firstEditBtn.click()
  await page.waitForTimeout(2000)

  recordTest('OAuth2 edit page loads', page.url().includes('/oauth2/credentials/'))

  const name = await readField(page, 'Name')
  recordTest('OAuth2 name field populated', name.length > 0)

  const provider = await readSelect(page, 'Provider')
  recordTest('OAuth2 provider field populated', provider !== null && provider.length > 0)

  const clientId = await readField(page, 'Client ID')
  recordTest('OAuth2 client ID field populated', clientId.length > 0)
}

// ================================================================
// TEST: Scheduler form loads data
// ================================================================
async function testSchedulerFormLoad(page) {
  console.log('\n--- Scheduler Edit Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/schedulers`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  if (rows === 0) {
    skipTest('Scheduler form load', 'No schedulers in table')
    return
  }

  const firstEditBtn = page.locator('.ant-table-tbody tr').first().locator('button').first()
  await firstEditBtn.click()
  await page.waitForTimeout(2000)

  recordTest('Scheduler edit page loads', page.url().includes('/schedulers/'))

  const pipeline = await readSelect(page, 'Pipeline')
  recordTest('Scheduler pipeline field populated', pipeline !== null && pipeline.length > 0)

  const scheduleType = await readSelect(page, 'Schedule Type')
  recordTest('Scheduler type field populated', scheduleType !== null && scheduleType.length > 0)
}

// ================================================================
// TEST: Sync Policy form loads data
// ================================================================
async function testSyncPolicyFormLoad(page) {
  console.log('\n--- Sync Policy Edit Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/syncpolicies`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  if (rows === 0) {
    skipTest('Sync policy form load', 'No sync policies in table')
    return
  }

  const firstEditBtn = page.locator('.ant-table-tbody tr').first().locator('button').first()
  await firstEditBtn.click()
  await page.waitForTimeout(2000)

  recordTest('Sync policy edit page loads', page.url().includes('/syncpolicy/'))

  const name = await readField(page, 'Name')
  recordTest('Sync policy name field populated', name.length > 0)

  const pipeline = await readSelect(page, 'Pipeline')
  recordTest('Sync policy pipeline field populated', pipeline !== null && pipeline.length > 0)
}

// ================================================================
// TEST: Profile form loads data
// ================================================================
async function testProfileFormLoad(page) {
  console.log('\n--- Profile Form Load ---')

  await page.goto(`${CONFIG.baseUrl}/profile`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  recordTest('Profile page loads', page.url().includes('/profile'))

  const email = await readField(page, 'Email').catch(() => '')
  const firstName = await readField(page, 'First Name').catch(() => '')
  recordTest('Profile fields populated', email.length > 0 || firstName.length > 0)
}

// ================================================================
// MAIN
// ================================================================
async function testFormLoad() {
  console.log('Testing ShadowAPI Form Data Loading...\n')

  const browser = await chromium.launch({ headless: false })
  const context = await browser.newContext({ ignoreHTTPSErrors: true })
  const page = await context.newPage()

  page.on('console', (msg) => {
    if (msg.type() === 'error') console.log(`[Browser ERROR]`, msg.text())
  })

  try {
    console.log('--- Authentication ---')
    const authenticated = await authenticateUser(page, { verbose: false })
    recordTest('Login successful', authenticated)

    if (!authenticated) {
      console.log('\nCannot proceed without authentication')
      await browser.close()
      return false
    }

    await testUserFormLoad(page)
    await testStorageFormLoad(page)
    await testDataSourceFormLoad(page)
    await testOAuth2FormLoad(page)
    await testSchedulerFormLoad(page)
    await testSyncPolicyFormLoad(page)
    await testProfileFormLoad(page)
  } catch (error) {
    console.error('\nTest suite error:', error.message)
    results.failed++
  } finally {
    await browser.close()
  }

  // Summary
  console.log(`\n${'='.repeat(60)}`)
  console.log('FORM LOAD TEST SUMMARY')
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

testFormLoad()
  .then((success) => {
    console.log(success ? '\nAll form load tests passed!' : '\nSome form load tests failed')
    process.exit(success ? 0 : 1)
  })
  .catch((error) => {
    console.error('\nTest suite crashed:', error.message)
    process.exit(1)
  })
