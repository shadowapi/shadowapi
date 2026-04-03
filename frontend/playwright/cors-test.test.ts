import { test, expect } from '@playwright/test';

test('check DNS resolution in browser', async ({ browser }) => {
  // Launch with specific DNS settings
  const context = await browser.newContext();
  const page = await context.newPage();

  // Navigate to a known working page
  const resp1 = await page.goto('http://localtest.me/health');
  console.log('localtest.me/health status:', resp1?.status());

  // Try the API URL
  const resp2 = await page.goto('http://api.localtest.me/api/v1/auth/oauth2/session');
  console.log('api.localtest.me status:', resp2?.status());
  console.log('api.localtest.me URL:', page.url());

  // Check if maybe there's a service worker or redirect
  const serverHeader = resp2?.headers()['server'];
  console.log('Server header:', serverHeader);

  // Try with a unique path to avoid caching
  const ts = Date.now();
  const resp3 = await page.goto(`http://api.localtest.me/api/v1/auth/oauth2/session?t=${ts}`);
  console.log('api.localtest.me with cache buster:', resp3?.status());

  await context.close();
});
