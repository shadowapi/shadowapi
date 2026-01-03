import { test, expect } from '@playwright/test';

test.describe('Message Query', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.getByPlaceholder('Email').fill('admin@example.com');
    await page.getByPlaceholder('Password').fill('Admin123!');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await page.waitForURL(/\/login\?login_challenge=/, { timeout: 15000 });
    await page.getByPlaceholder('Email').fill('admin@example.com');
    await page.getByPlaceholder('Password').fill('Admin123!');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await page.waitForURL(/\/workspaces/, { timeout: 15000 });
  });

  test('should trigger message query job on postgres storage', async ({ page, request }) => {
    // Get the access token cookie
    const cookies = await page.context().cookies();
    const accessToken = cookies.find(c => c.name === 'shadowapi_access_token')?.value;
    expect(accessToken).toBeTruthy();

    // Call the message query endpoint
    // Storage UUID for "Test Postgres" storage
    const storageUUID = '019b8317-79e1-7df7-b42f-b364e5b77d8d';

    const response = await request.post(
      `http://api.localtest.me/api/v1/storage/postgres/${storageUUID}/messages/query`,
      {
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        },
        data: {
          limit: 10,
        },
      }
    );

    // Verify the response
    expect(response.status()).toBe(202);
    const body = await response.json();
    expect(body.uuid).toBeTruthy();
    expect(body.nats_subject).toContain('shadowapi.data.workspace.internal.messages');
    expect(body.status).toBe('pending');
    // tables_queried should be populated with configured tables
    expect(Array.isArray(body.tables_queried)).toBe(true);

    console.log('Message query job created:', body);
  });
});
