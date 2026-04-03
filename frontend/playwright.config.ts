import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: 'playwright',
  testMatch: '**/*test*.{js,ts,mjs,cjs}',
  timeout: 120000,
  use: {
    baseURL: 'https://shadowapi.local',
    headless: false,
    ignoreHTTPSErrors: true,
    actionTimeout: 60000,
    navigationTimeout: 60000,
  },
})
