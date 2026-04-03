import { test, expect } from '@playwright/test';

test('debug login flow', async ({ page }) => {
  // Capture all console messages
  page.on('console', msg => {
    console.log(`[browser ${msg.type()}] ${msg.text()}`);
  });

  // Capture all requests
  page.on('request', req => {
    console.log(`[req] ${req.method()} ${req.url()}`);
  });

  page.on('response', resp => {
    console.log(`[resp] ${resp.status()} ${resp.url()}`);
  });

  page.on('requestfailed', req => {
    console.log(`[FAILED] ${req.method()} ${req.url()} ${req.failure()?.errorText}`);
  });

  await page.goto('http://localtest.me/login');

  // Wait to see what happens
  await page.waitForTimeout(10000);

  console.log('Final URL:', page.url());
  const content = await page.content();
  console.log('Page title:', await page.title());
});
