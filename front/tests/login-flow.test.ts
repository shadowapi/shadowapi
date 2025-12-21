import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  test('should login successfully with valid credentials', async ({ page }) => {
    // Navigate to login page
    await page.goto('/login');

    // Verify we're on the login page
    await expect(page.getByRole('heading', { name: 'MeshPump' })).toBeVisible();

    // Step 1: Fill in credentials and submit to initiate OAuth2 flow
    await page.getByPlaceholder('Email').fill('admin@example.com');
    await page.getByPlaceholder('Password').fill('Admin123!');
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Wait for OAuth2 redirect back to login page with login_challenge
    await page.waitForURL(/\/login\?login_challenge=/, { timeout: 15000 });

    // Step 2: Fill credentials again (form is reset after redirect)
    await page.getByPlaceholder('Email').fill('admin@example.com');
    await page.getByPlaceholder('Password').fill('Admin123!');
    await page.getByRole('button', { name: 'Sign in' }).click();

    // Wait for OAuth2 flow to complete and redirect to home
    await page.waitForURL('/', { timeout: 15000 });

    // Verify we're on the home page (authenticated)
    await expect(page).toHaveURL('/');
  });
});
