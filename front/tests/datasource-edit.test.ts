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

test.describe('DataSource Edit Page', () => {
  test('should load existing datasource data in edit form', async ({ page }) => {
    // Capture all API requests
    const apiRequests: { url: string; response: any }[] = [];
    page.on('response', async (response) => {
      if (response.url().includes('/api/v1/datasource')) {
        try {
          const json = await response.json();
          apiRequests.push({ url: response.url(), response: json });
          console.log('API Response:', response.url(), JSON.stringify(json, null, 2));
        } catch (e) {
          console.log('API Response (non-JSON):', response.url(), response.status());
        }
      }
    });

    // Capture console logs from the browser
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('[DataSourceEdit]') || msg.type() === 'error') {
        console.log('Browser:', text);
      }
    });

    // Login first
    await login(page);

    // Navigate to demo workspace datasources
    await page.goto('http://app.localtest.me/w/demo/datasources');
    await page.waitForLoadState('networkidle');

    console.log('Datasources page URL:', page.url());

    // Navigate directly to the datasource edit page
    await page.goto('http://app.localtest.me/w/demo/datasources/019b692b-1a9b-7bcf-ba8d-53dd28d9d188');
    await page.waitForLoadState('networkidle');

    // Wait for the form to potentially load data
    await page.waitForTimeout(3000);

    // Log the form state
    console.log('Edit page URL:', page.url());

    // Check the title
    const title = await page.getByRole('heading', { level: 4 }).textContent();
    console.log('Page title:', title);

    // Check form field values
    const nameField = page.getByLabel('Name');
    const nameValue = await nameField.inputValue();
    console.log('Name field value:', JSON.stringify(nameValue));

    // Log all API requests captured
    console.log('Total API requests captured:', apiRequests.length);
    apiRequests.forEach((req, i) => {
      console.log(`Request ${i}:`, req.url);
    });

    // Check all input values
    const allInputs = page.locator('input[type="text"], input[id]');
    const inputCount = await allInputs.count();
    console.log('Total input fields:', inputCount);

    for (let i = 0; i < inputCount; i++) {
      const input = allInputs.nth(i);
      const id = await input.getAttribute('id');
      const value = await input.inputValue();
      console.log(`Input ${i}: id=${id}, value="${value}"`);
    }

    // Assertions - The name field should have a value if data is loaded
    expect(title).toBe('Edit Data Source');
    expect(nameValue).not.toBe(''); // This should fail if data isn't loaded
  });

  test('should show form fields based on datasource type', async ({ page }) => {
    await login(page);

    // Navigate to create a new datasource
    await page.goto('http://app.localtest.me/w/demo/datasources/new');
    await page.waitForLoadState('networkidle');

    // Check default type is Email IMAP
    const typeSelector = page.locator('[class*="ant-select"]').first();
    const selectedType = await typeSelector.textContent();
    console.log('Default selected type:', selectedType);

    // Check that Email IMAP fields are visible
    expect(page.getByLabel('Name')).toBeVisible();
    expect(page.getByLabel('Email Address')).toBeVisible();
    expect(page.getByLabel('IMAP Server')).toBeVisible();
    expect(page.getByLabel('SMTP Server')).toBeVisible();

    // Change to Email OAuth
    await page.getByLabel('Type').click();
    await page.locator('.ant-select-item-option-content').getByText('Email OAuth').click();
    await page.waitForTimeout(500);

    // Check that Email OAuth fields are visible
    expect(page.getByLabel('Email Address')).toBeVisible();
    expect(page.getByLabel('Provider')).toBeVisible();
    expect(page.getByLabel('OAuth2 Client')).toBeVisible();

    // IMAP/SMTP fields should not be visible
    await expect(page.getByLabel('IMAP Server')).not.toBeVisible();
    await expect(page.getByLabel('SMTP Server')).not.toBeVisible();
  });
});
