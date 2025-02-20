import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: './tests',
    timeout: 30000,
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
});
