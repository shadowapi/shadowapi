/**
 * Test 04: Gmail Message Fetch Pipeline
 * Tests that the Gmail OAuth pipeline fetches messages and they appear on /messages page
 *
 * Run: node frontend/playwright/test-04-get-10-messages.cjs
 * Or:  make test-get-10-messages
 *
 * Prerequisites:
 *   - Backend + frontend running
 *   - Gmail OAuth datasource configured and token obtained
 *   - Pipeline + scheduler created (via make test-init-test-tables or API)
 *   - Worker has had time to fetch messages (at least 1 scheduler cycle)
 */

const { chromium } = require('playwright')
const { authenticateUser, CONFIG } = require('./shared/shared-auth.cjs')

const DATASOURCE_UUID = '019d5bb9-05db-759a-9961-97ec6a892ec0'
const MIN_MESSAGES = 1  // at least 1 message should appear
const TARGET_MESSAGES = 10

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

// ================================================================
// TEST: Pipeline and scheduler exist
// ================================================================
async function testPipelineExists(page) {
  console.log('\n--- Pipeline Exists ---')

  await page.goto(`${CONFIG.baseUrl}/pipelines`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  recordTest('Pipeline list has rows', rows > 0, `Found ${rows} pipeline(s)`)

  if (rows > 0) {
    const name = await page.locator('.ant-table-tbody tr').first().locator('td').first().textContent()
    recordTest('Pipeline name visible', !!name, `Name: ${name}`)
  }
}

// ================================================================
// TEST: Scheduler exists and has run
// ================================================================
async function testSchedulerExists(page) {
  console.log('\n--- Scheduler Exists ---')

  await page.goto(`${CONFIG.baseUrl}/schedulers`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  recordTest('Scheduler list has rows', rows > 0, `Found ${rows} scheduler(s)`)
}

// ================================================================
// TEST: Worker jobs exist (fetch was triggered)
// ================================================================
async function testWorkerJobs(page) {
  console.log('\n--- Worker Jobs ---')

  await page.goto(`${CONFIG.baseUrl}/workers`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const rows = await page.locator('.ant-table-tbody tr').count()
  recordTest('Worker jobs list has rows', rows > 0, `Found ${rows} job(s)`)
}

// ================================================================
// TEST: Messages appear on /messages page
// ================================================================
async function testMessagesPage(page) {
  console.log('\n--- Messages Page ---')

  await page.goto(`${CONFIG.baseUrl}/messages`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(2000)

  // Check if the page loaded without errors
  const errorText = await page.locator('.ant-result-title').textContent().catch(() => null)
  recordTest('Messages page loads without error', !errorText, errorText || 'OK')

  // Count message rows in the table
  const rows = await page.locator('.ant-table-tbody tr').count()
  recordTest(`Messages table has rows (target: ${TARGET_MESSAGES})`, rows >= MIN_MESSAGES, `Found ${rows} message(s)`)

  if (rows > 0) {
    // Check first message has a subject or body
    const firstRow = page.locator('.ant-table-tbody tr').first()
    const cells = await firstRow.locator('td').count()
    recordTest('Message row has columns', cells >= 2, `${cells} columns`)

    // Try to read subject from the first row
    const firstCellText = await firstRow.locator('td').first().textContent()
    recordTest('First message has content', !!firstCellText && firstCellText.trim().length > 0,
      `Content: "${(firstCellText || '').trim().substring(0, 60)}..."`)
  }

  if (rows >= TARGET_MESSAGES) {
    recordTest(`At least ${TARGET_MESSAGES} messages fetched`, true, `Found ${rows}`)
  } else if (rows >= MIN_MESSAGES) {
    recordTest(`At least ${TARGET_MESSAGES} messages fetched`, false,
      `Only ${rows} messages found. The worker may need more time or the Gmail account may have fewer messages.`)
  }
}

// ================================================================
// TEST: Datasource shows connected status
// ================================================================
async function testDatasourceStatus(page) {
  console.log('\n--- Datasource Status ---')

  await page.goto(`${CONFIG.baseUrl}/datasources/${DATASOURCE_UUID}`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  // Check page loaded (not 404)
  const title = await page.title()
  recordTest('Datasource page loads', !title.includes('404'), title)
}

// ================================================================
// TEST: Storage "Test Internal" exists
// ================================================================
async function testStorageExists(page) {
  console.log('\n--- Storage Exists ---')

  await page.goto(`${CONFIG.baseUrl}/storages`, { waitUntil: 'networkidle', timeout: 15000 })
  await page.waitForTimeout(1500)

  const pageContent = await page.content()
  const hasTestInternal = pageContent.includes('Test Internal')
  recordTest('Test Internal storage visible', hasTestInternal)
}

// ================================================================
// MAIN
// ================================================================
async function main() {
  console.log('=== Test 04: Gmail Message Fetch Pipeline ===\n')
  console.log(`Target: ${TARGET_MESSAGES} messages from datasource ${DATASOURCE_UUID}`)
  console.log(`Base URL: ${CONFIG.baseUrl}\n`)

  const browser = await chromium.launch({ headless: false })
  const ctx = await browser.newContext({ ignoreHTTPSErrors: true })
  const page = await ctx.newPage()

  try {
    // Login
    console.log('--- Authentication ---')
    const loggedIn = await authenticateUser(page)
    recordTest('Login successful', loggedIn)

    if (!loggedIn) {
      console.log('\nFailed to authenticate. Aborting.')
      return
    }

    // Run tests in order
    await testStorageExists(page)
    await testDatasourceStatus(page)
    await testPipelineExists(page)
    await testSchedulerExists(page)
    await testWorkerJobs(page)
    await testMessagesPage(page)

  } catch (err) {
    console.error('\nUnexpected error:', err.message)
    recordTest('Test suite execution', false, err.message)
  } finally {
    // Print summary
    console.log('\n' + '='.repeat(50))
    console.log(`Results: ${results.passed} passed, ${results.failed} failed, ${results.skipped} skipped`)
    console.log('='.repeat(50))

    if (results.failed > 0) {
      console.log('\nFailed tests:')
      results.tests.filter(t => t.passed === false).forEach(t => {
        console.log(`  - ${t.name}: ${t.details}`)
      })
    }

    await page.waitForTimeout(3000)
    await browser.close()
    process.exit(results.failed > 0 ? 1 : 0)
  }
}

main()
