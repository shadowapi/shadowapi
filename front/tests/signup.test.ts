import { test, expect } from '@playwright/test';

test('Signup form submission', async ({ page }) => {
    // Step 1: Visit the signup page
    await page.goto('http://localtest.me/signup');

    // Step 2: Wait for redirect to the signup URL with flow_id
    await page.waitForURL(/http:\/\/localtest\.me\/signup\?flow_id=[a-f0-9\-]+/);

    // Verify the redirected URL contains the flow_id parameter
    const url = page.url();
    expect(url).toMatch(/flow_id=[a-f0-9\-]+/);

    // Step 3: Fill out the form fields
    await page.fill('input[name="email"]', 'testuser@example.com');
    await page.fill('input[name="password"]', 'TestPassword123!');
    await page.fill('input[name="passwordConfirm"]', 'TestPassword123!');
    await page.fill('input[name="firstName"]', 'Test');
    await page.fill('input[name="lastName"]', 'User');

    // Step 4: Submit the form
    await page.click('button[type="submit"]');

    // Step 5: Wait for navigation or response
    await page.waitForResponse((response) => response.status() === 200);

    // Optional: Verify a success message or redirection after submission
    const bodyContent = await page.content();
    expect(bodyContent).toContain('Success'); // Adjust this to match your actual success message
});
