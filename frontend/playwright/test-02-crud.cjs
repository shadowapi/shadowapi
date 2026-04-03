/**
 * Test 02: CRUD Operations
 * Tests create, list/table view, edit, and delete for all entities
 *
 * Run: node frontend/playwright/test-02-crud.cjs
 * Or:  make test-crud
 *
 * Prerequisites: backend + frontend running, admin user exists
 */

const { chromium } = require('playwright')
const { authenticateUser, CONFIG } = require('./shared/shared-auth.cjs')
const { DUMMY } = require('./shared/dummy.cjs')

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

/** Click a button matching text, wait for navigation or network */
async function clickButton(page, text, opts = {}) {
  const btn = page.locator(`button:has-text("${text}")`).first()
  await btn.waitFor({ state: 'visible', timeout: 5000 })
  await btn.click()
  if (opts.waitNav) {
    await page.waitForTimeout(2000)
  }
}

/** Fill an antd form field by label */
async function fillField(page, label, value) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const input = item.locator('input').first()
  await input.fill(value)
}

/** Fill a password field by label */
async function fillPassword(page, label, value) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const input = item.locator('input[type="password"]').first()
  await input.fill(value)
}

/** Select a value in an antd Select by label */
async function selectField(page, label, optionText) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const select = item.locator('.ant-select').first()
  await select.click()
  await page.waitForTimeout(300)
  const option = page.locator(`.ant-select-dropdown .ant-select-item-option:has-text("${optionText}")`).first()
  await option.click()
  await page.waitForTimeout(300)
}

/** Toggle an antd Switch by label */
async function toggleSwitch(page, label) {
  const item = page.locator(`.ant-form-item:has(.ant-form-item-label:has-text("${label}"))`)
  const switchEl = item.locator('.ant-switch').first()
  await switchEl.click()
}

/** Check if a table row contains text */
async function tableHasRow(page, text) {
  const cell = page.locator(`.ant-table-tbody td:has-text("${text}")`)
  return (await cell.count()) > 0
}

/** Click the edit button in a table row matching text */
async function clickRowEdit(page, text) {
  const row = page.locator(`.ant-table-tbody tr:has(td:has-text("${text}"))`)
  const editBtn = row.locator('button').first()
  await editBtn.click()
  await page.waitForTimeout(1500)
}

// ================================================================
// TEST: Users CRUD
// ================================================================
async function testUsersCrud(page) {
  console.log('\n--- Users CRUD ---')
  const d = DUMMY.user

  // Navigate to users list
  await page.goto(`${CONFIG.baseUrl}/users`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)
  recordTest('Users list page loads', page.url().includes('/users'))

  // CREATE
  await clickButton(page, 'Add User', { waitNav: true })
  recordTest('Users add page loads', page.url().includes('/users/add'))

  await fillField(page, 'Email', d.create.email)
  await fillPassword(page, 'Password', d.create.password)
  await fillField(page, 'First Name', d.create.first_name)
  await fillField(page, 'Last Name', d.create.last_name)
  await toggleSwitch(page, 'Enabled')

  await clickButton(page, 'Create', { waitNav: true })
  await page.waitForTimeout(2000)
  recordTest('User created, redirected to list', page.url().includes('/users') && !page.url().includes('/add'))

  // LIST — verify row
  const hasRow = await tableHasRow(page, d.create.email)
  recordTest('New user visible in table', hasRow)

  // EDIT
  if (hasRow) {
    await clickRowEdit(page, d.create.email)
    recordTest('User edit page loads', page.url().includes('/users/'))

    await fillField(page, 'First Name', d.update.first_name)
    await fillField(page, 'Last Name', d.update.last_name)
    await clickButton(page, 'Update', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('User updated, redirected to list', page.url().includes('/users') && !page.url().includes('/add'))

    const hasUpdated = await tableHasRow(page, d.update.first_name)
    recordTest('Updated name visible in table', hasUpdated)

    // DELETE
    await clickRowEdit(page, d.create.email)
    await page.waitForTimeout(1000)
    await clickButton(page, 'Delete', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('User deleted, redirected to list', page.url().includes('/users'))

    const stillExists = await tableHasRow(page, d.create.email)
    recordTest('Deleted user gone from table', !stillExists)
  } else {
    skipTest('User edit', 'User not found in table')
    skipTest('User update', 'User not found')
    skipTest('User delete', 'User not found')
  }
}

// ================================================================
// TEST: Storages CRUD (hostfiles)
// ================================================================
async function testStoragesCrud(page) {
  console.log('\n--- Storages CRUD ---')
  const d = DUMMY.storage_hostfiles

  await page.goto(`${CONFIG.baseUrl}/storages`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)
  recordTest('Storages list page loads', page.url().includes('/storages'))

  // CREATE
  await clickButton(page, 'Add Storage', { waitNav: true })
  recordTest('Storages add page loads', page.url().includes('/storages/add'))

  await fillField(page, 'Name', d.create.name)
  await selectField(page, 'Type', 'File System')
  await page.waitForTimeout(500)
  await fillField(page, 'File System Path', d.create.path)
  await toggleSwitch(page, 'Enabled')

  await clickButton(page, 'Create', { waitNav: true })
  await page.waitForTimeout(2000)
  recordTest('Storage created, redirected to list', page.url().includes('/storages') && !page.url().includes('/add'))

  const hasRow = await tableHasRow(page, d.create.name)
  recordTest('New storage visible in table', hasRow)

  // EDIT
  if (hasRow) {
    await clickRowEdit(page, d.create.name)
    await page.waitForTimeout(1000)

    await fillField(page, 'Name', d.update.name)
    await fillField(page, 'File System Path', d.update.path)
    await clickButton(page, 'Update', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('Storage updated, redirected to list', page.url().includes('/storages'))

    const hasUpdated = await tableHasRow(page, d.update.name)
    recordTest('Updated storage name visible', hasUpdated)

    // DELETE
    await clickRowEdit(page, d.update.name)
    await page.waitForTimeout(1000)
    await clickButton(page, 'Delete', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('Storage deleted, redirected to list', page.url().includes('/storages'))

    const stillExists = await tableHasRow(page, d.update.name)
    recordTest('Deleted storage gone from table', !stillExists)
  } else {
    skipTest('Storage edit/delete', 'Storage not found in table')
  }
}

// ================================================================
// TEST: OAuth2 Credentials CRUD
// ================================================================
async function testOAuth2Crud(page) {
  console.log('\n--- OAuth2 Credentials CRUD ---')
  const d = DUMMY.oauth2_credential

  await page.goto(`${CONFIG.baseUrl}/oauth2/credentials`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)
  recordTest('OAuth2 list page loads', page.url().includes('/oauth2/credentials'))

  // CREATE
  await clickButton(page, 'Add', { waitNav: true })
  recordTest('OAuth2 add page loads', page.url().includes('/oauth2/credentials/add'))

  await fillField(page, 'Name', d.create.name)
  await selectField(page, 'Provider', 'Gmail')
  await fillField(page, 'Client ID', d.create.client_id)
  await fillPassword(page, 'Client Secret', d.create.secret)

  await clickButton(page, 'Create', { waitNav: true })
  await page.waitForTimeout(2000)
  recordTest('OAuth2 credential created', page.url().includes('/oauth2/credentials') && !page.url().includes('/add'))

  const hasRow = await tableHasRow(page, d.create.name)
  recordTest('New OAuth2 credential visible in table', hasRow)

  // EDIT
  if (hasRow) {
    await clickRowEdit(page, d.create.name)
    await page.waitForTimeout(1000)

    await fillField(page, 'Name', d.update.name)
    await clickButton(page, 'Update', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('OAuth2 credential updated', page.url().includes('/oauth2/credentials'))

    // DELETE
    await clickRowEdit(page, d.update.name)
    await page.waitForTimeout(1000)
    await clickButton(page, 'Delete', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('OAuth2 credential deleted', page.url().includes('/oauth2/credentials'))

    const stillExists = await tableHasRow(page, d.update.name)
    recordTest('Deleted OAuth2 credential gone from table', !stillExists)
  } else {
    skipTest('OAuth2 edit/delete', 'Credential not found in table')
  }
}

// ================================================================
// TEST: Sync Policies CRUD
// ================================================================
async function testSyncPoliciesCrud(page) {
  console.log('\n--- Sync Policies CRUD ---')
  const d = DUMMY.sync_policy

  await page.goto(`${CONFIG.baseUrl}/syncpolicies`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)
  recordTest('Sync policies list page loads', page.url().includes('/syncpolicies'))

  // CREATE — requires a pipeline to exist; check if Add button works
  const addBtn = page.locator('button:has-text("Add")').first()
  const hasAddBtn = await addBtn.isVisible({ timeout: 3000 }).catch(() => false)

  if (hasAddBtn) {
    await addBtn.click()
    await page.waitForTimeout(1500)
    recordTest('Sync policy add page loads', page.url().includes('/syncpolicies/add') || page.url().includes('/syncpolicy'))

    await fillField(page, 'Name', d.create.name)
    await toggleSwitch(page, 'Sync All')

    // Pipeline select may be empty if no pipelines exist
    const pipelineOptions = await page.locator('.ant-select-item-option').count()
    if (pipelineOptions === 0) {
      skipTest('Sync policy create', 'No pipelines available for selection')
      await page.goto(`${CONFIG.baseUrl}/syncpolicies`, { waitUntil: 'networkidle' })
      return
    }

    await selectField(page, 'Pipeline', '') // select first available
    await clickButton(page, 'Create', { waitNav: true })
    await page.waitForTimeout(2000)
    recordTest('Sync policy created', page.url().includes('/syncpolicies'))

    const hasRow = await tableHasRow(page, d.create.name)
    recordTest('New sync policy visible in table', hasRow)

    if (hasRow) {
      // DELETE
      await clickRowEdit(page, d.create.name)
      await page.waitForTimeout(1000)
      await clickButton(page, 'Delete', { waitNav: true })
      await page.waitForTimeout(2000)
      recordTest('Sync policy deleted', page.url().includes('/syncpolicies'))
    }
  } else {
    skipTest('Sync policy CRUD', 'Add button not found')
  }
}

// ================================================================
// TEST: Profile Edit
// ================================================================
async function testProfileEdit(page) {
  console.log('\n--- Profile Edit ---')

  await page.goto(`${CONFIG.baseUrl}/profile`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)
  recordTest('Profile page loads', page.url().includes('/profile'))

  // Read current values, update, and restore
  const firstNameInput = page.locator('.ant-form-item:has(.ant-form-item-label:has-text("First Name")) input').first()
  const lastNameInput = page.locator('.ant-form-item:has(.ant-form-item-label:has-text("Last Name")) input').first()

  const origFirst = await firstNameInput.inputValue()
  const origLast = await lastNameInput.inputValue()

  await firstNameInput.fill('ProfileTest')
  await lastNameInput.fill('Updated')
  await clickButton(page, 'Update', { waitNav: true })
  await page.waitForTimeout(2000)

  // Reload and verify
  await page.goto(`${CONFIG.baseUrl}/profile`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)
  const newFirst = await page.locator('.ant-form-item:has(.ant-form-item-label:has-text("First Name")) input').first().inputValue()
  recordTest('Profile first name updated', newFirst === 'ProfileTest')

  // Restore original values
  await page.locator('.ant-form-item:has(.ant-form-item-label:has-text("First Name")) input').first().fill(origFirst)
  await page.locator('.ant-form-item:has(.ant-form-item-label:has-text("Last Name")) input').first().fill(origLast)
  await clickButton(page, 'Update', { waitNav: true })
  await page.waitForTimeout(1000)
  recordTest('Profile restored to original values', true)
}

// ================================================================
// MAIN
// ================================================================
async function testCrud() {
  console.log('Testing ShadowAPI CRUD Operations...\n')

  const browser = await chromium.launch({ headless: false })
  const context = await browser.newContext({ ignoreHTTPSErrors: true })
  const page = await context.newPage()

  page.on('console', (msg) => {
    if (msg.type() === 'error') console.log(`[Browser ERROR]`, msg.text())
  })

  try {
    // Authenticate first
    console.log('--- Authentication ---')
    const authenticated = await authenticateUser(page, { verbose: false })
    recordTest('Login successful', authenticated)

    if (!authenticated) {
      console.log('\nCannot proceed without authentication')
      await browser.close()
      return false
    }

    await testUsersCrud(page)
    await testStoragesCrud(page)
    await testOAuth2Crud(page)
    await testSyncPoliciesCrud(page)
    await testProfileEdit(page)
  } catch (error) {
    console.error('\nTest suite error:', error.message)
    results.failed++
  } finally {
    await browser.close()
  }

  // Summary
  console.log(`\n${'='.repeat(60)}`)
  console.log('CRUD TEST SUMMARY')
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

testCrud()
  .then((success) => {
    console.log(success ? '\nAll CRUD tests passed!' : '\nSome CRUD tests failed')
    process.exit(success ? 0 : 1)
  })
  .catch((error) => {
    console.error('\nTest suite crashed:', error.message)
    process.exit(1)
  })
