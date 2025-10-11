import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  timeout: 20000, // 20 second timeout as requested
  expect: { timeout: 5000 },
  use: {
    baseURL: 'http://localtest.me/', // Replace with your frontend server URL
    headless: true,
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  reporter: [['html', { open: 'on-failure' }]],
})
