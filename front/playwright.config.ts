import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  use: {
    baseURL: 'http://localtest.me',
    headless: true,
    screenshot: 'only-on-failure',
  },
  reporter: 'html',
});
