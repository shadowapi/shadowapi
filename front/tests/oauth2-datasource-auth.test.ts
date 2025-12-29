import { test, expect } from '@playwright/test';

// Helper to login
async function login(page) {
  await page.goto('http://app.localtest.me/login');
  await page.getByPlaceholder('Email').fill('admin@example.com');
  await page.getByPlaceholder('Password').fill('Admin123!');
  await page.getByRole('button', { name: 'Sign in' }).click();

  // Wait for OAuth2 redirect back to login page with login_challenge
  await page.waitForURL(/\/login\?login_challenge=/, { timeout: 15000 });

  // Fill credentials again (form is reset after redirect)
  await page.getByPlaceholder('Email').fill('admin@example.com');
  await page.getByPlaceholder('Password').fill('Admin123!');
  await page.getByRole('button', { name: 'Sign in' }).click();

  // Wait for OAuth2 flow to complete
  await page.waitForURL(/\/workspaces/, { timeout: 15000 });
}

test.describe('OAuth2 Datasource Authentication Status', () => {
  test('should show OAuth authentication status on email_oauth datasource edit page', async ({ page }) => {
    // Capture all API requests and responses
    const apiRequests: { url: string; method: string; status: number; response?: any }[] = [];

    page.on('response', async (response) => {
      const url = response.url();
      if (url.includes('/api/v1/')) {
        const entry: { url: string; method: string; status: number; response?: any } = {
          url,
          method: response.request().method(),
          status: response.status(),
        };
        try {
          const json = await response.json();
          entry.response = json;
          console.log(`API ${entry.method} ${entry.status}:`, url);
          if (entry.status >= 400) {
            console.log('Error response:', JSON.stringify(json, null, 2));
          }
        } catch (e) {
          console.log(`API ${entry.method} ${entry.status} (non-JSON):`, url);
        }
        apiRequests.push(entry);
      }
    });

    // Capture console logs from the browser
    page.on('console', msg => {
      if (msg.type() === 'error' || msg.text().includes('[DataSourceEdit]')) {
        console.log('Browser console:', msg.text());
      }
    });

    // Login first
    await login(page);

    // Navigate directly to the existing email_oauth datasource edit page
    await page.goto('http://app.localtest.me/w/demo/datasources/019b692b-1a9b-7bcf-ba8d-53dd28d9d188');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000); // Wait for OAuth status check

    console.log('Edit page URL:', page.url());

    // Log all API requests we captured
    console.log('\n=== All API Requests ===');
    apiRequests.forEach((r, i) => {
      console.log(`${i + 1}. ${r.method} ${r.status}: ${r.url}`);
      if (r.status >= 400 && r.response) {
        console.log('   Error:', JSON.stringify(r.response));
      }
    });

    // Find the token list API call
    const tokenListCall = apiRequests.find(r =>
      r.url.includes('/oauth2/client/') && r.url.includes('/token')
    );

    if (tokenListCall) {
      console.log('\n=== Token List API Call ===');
      console.log('URL:', tokenListCall.url);
      console.log('Status:', tokenListCall.status);
      console.log('Response:', JSON.stringify(tokenListCall.response, null, 2));

      // The bug: if status is 404, it means the handler is using datasource UUID
      // to look up OAuth2 client instead of first getting oauth2_client_uuid from datasource
      if (tokenListCall.status === 404) {
        console.log('\nBUG CONFIRMED: Token list API returned 404');
        console.log('The handler is incorrectly using datasource UUID to look up OAuth2 client');
      }
    } else {
      console.log('\nNo token list API call found!');
    }

    // Look for the authentication status alert
    const authAlert = page.locator('.ant-alert');
    const alertVisible = await authAlert.isVisible({ timeout: 3000 }).catch(() => false);
    console.log('\nAuth alert visible:', alertVisible);

    if (alertVisible) {
      const alertText = await authAlert.textContent();
      console.log('Alert text:', alertText);
      // We expect to see either "authenticated" or "Not authenticated"
      expect(alertText).toMatch(/(authenticated|Not authenticated)/);
    } else {
      // If no alert is visible, check if the token API call failed
      if (tokenListCall && tokenListCall.status >= 400) {
        // This is the bug - fail the test with a clear message
        expect.soft(tokenListCall.status, 'Token list API should return 200, not ' + tokenListCall.status).toBe(200);
      }
    }

    // Verify the form loaded correctly (Name field should have a value)
    const nameField = page.getByLabel('Name');
    const nameValue = await nameField.inputValue();
    console.log('\nName field value:', nameValue);
    expect(nameValue).not.toBe('');
  });
});
