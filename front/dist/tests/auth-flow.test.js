"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const test_1 = require("@playwright/test");
(0, test_1.test)('guest can sign up, log in, and access protected pages', async ({ page }) => {
    page.on('console', (msg) => console.log('BROWSER:', msg.type(), msg.text()));
    page.on('requestfailed', (req) => console.log('REQ FAILED:', req.method(), req.url(), req.failure()?.errorText));
    page.on('response', async (res) => {
        if (res.status() >= 400) {
            console.log('HTTP ERR:', res.status(), res.url());
        }
    });
    const email = `test.user.${Date.now()}@example.com`;
    const password = 'TestPassword123!';
    const usersListResponse = [
        {
            uuid: 'user-123',
            email,
            first_name: 'Test',
            last_name: 'User',
        },
    ];
    await test_1.test.step('Redirect unauthenticated user from a protected route', async () => {
        await page.goto('/users');
        await (0, test_1.expect)(page).toHaveURL(/\/login\?returnTo=%2Fusers$/);
        await (0, test_1.expect)(page.getByLabel('Email')).toBeVisible();
    });
    await test_1.test.step('Allow guests to open the signup page', async () => {
        await page.getByRole('link', { name: /sign up/i }).click();
        await (0, test_1.expect)(page).toHaveURL(/\/signup$/);
        await (0, test_1.expect)(page.getByLabel('First Name')).toBeVisible();
        await (0, test_1.expect)(page.getByRole('button', { name: /^Sign Up$/i })).toBeVisible();
    });
    await page.route('**/api/v1/user', async (route) => {
        const request = route.request();
        const method = request.method();
        console.log('ROUTE /api/v1/user', method, request.url());
        if (method === 'POST') {
            await route.fulfill({
                status: 201,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    uuid: usersListResponse[0].uuid,
                    email: usersListResponse[0].email,
                    first_name: usersListResponse[0].first_name,
                    last_name: usersListResponse[0].last_name,
                }),
            });
            return;
        }
        if (method === 'GET') {
            await route.fulfill({
                status: 200,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(usersListResponse),
            });
            return;
        }
        console.log('ROUTE /api/v1/user FALLBACK', method);
        await route.fallback();
    });
    await test_1.test.step('Submit signup form for a new user', async () => {
        await page.getByLabel('First Name').fill('Test');
        await page.getByLabel('Last Name').fill('User');
        await page.getByLabel('Email').fill(email);
        await page.getByLabel('Password', { exact: true }).fill(password);
        await page.getByLabel('Confirm Password').fill(password);
        await Promise.all([
            page.waitForRequest((req) => req.method() === 'POST' && req.url().endsWith('/api/v1/user')),
            page.getByRole('button', { name: /^Sign Up$/i }).click(),
        ]);
        await (0, test_1.expect)(page).toHaveURL(/\/login\?returnTo=%2F$/);
        await (0, test_1.expect)(page.getByLabel('Email')).toBeVisible();
    });
    await page.route('**/api/v1/user/session', async (route) => {
        if (route.request().method() === 'POST') {
            await route.fulfill({
                status: 200,
                headers: {
                    'Content-Type': 'application/json',
                    'Access-Control-Allow-Origin': '*',
                },
                body: JSON.stringify({
                    session_token: 'backend-session-token',
                    zitadel_url: 'http://auth.localtest.me',
                    expires_in: 3600,
                }),
            });
            return;
        }
        await route.fallback();
    });
    // Intercept Zitadel session endpoints for both collection and specific-session paths
    await page.route(/\/v2\/sessions(\/.*)?$/, async (route) => {
        const method = route.request().method();
        if (method === 'OPTIONS') {
            await route.fulfill({
                status: 200,
                headers: {
                    'Access-Control-Allow-Origin': '*',
                    'Access-Control-Allow-Methods': 'POST, PATCH, OPTIONS',
                    'Access-Control-Allow-Headers': 'Content-Type, Authorization',
                },
                body: '',
            });
            return;
        }
        if (method === 'POST' || method === 'PATCH') {
            await route.fulfill({
                status: 200,
                headers: {
                    'Content-Type': 'application/json',
                    'Access-Control-Allow-Origin': '*',
                },
                body: JSON.stringify({
                    sessionId: 'session-123',
                    sessionToken: 'session-token-123',
                    changeDate: new Date().toISOString(),
                }),
            });
            return;
        }
        await route.fallback();
    });
    // Mock the authorize endpoint to redirect back with authRequest parameter
    await page.route('**/oauth/v2/authorize*', async (route) => {
        const url = new URL(route.request().url());
        const redirectUri = url.searchParams.get('redirect_uri');
        // Simulate Zitadel redirecting back to login with authRequest parameter
        await route.fulfill({
            status: 302,
            headers: {
                'Location': `${redirectUri}?authRequest=test-auth-request-123`,
                'Access-Control-Allow-Origin': '*',
            },
        });
    });
    // Mock the auth request finalize endpoint
    await page.route('**/v2/oidc/auth_requests/*', async (route) => {
        if (route.request().method() === 'POST') {
            await route.fulfill({
                status: 200,
                headers: {
                    'Content-Type': 'application/json',
                    'Access-Control-Allow-Origin': '*',
                },
                body: JSON.stringify({
                    callbackUrl: 'http://localtest.me/login?code=test-authorization-code-123&state=test-state',
                }),
            });
            return;
        }
        await route.fallback();
    });
    // Mock the token exchange endpoint
    await page.route('**/oauth/v2/token', async (route) => {
        if (route.request().method() === 'POST') {
            await route.fulfill({
                status: 200,
                headers: {
                    'Content-Type': 'application/json',
                    'Access-Control-Allow-Origin': '*',
                },
                body: JSON.stringify({
                    access_token: 'test-access-token-jwt',
                    id_token: 'test-id-token-jwt',
                    refresh_token: 'test-refresh-token',
                    token_type: 'Bearer',
                    expires_in: 3600,
                }),
            });
            return;
        }
        await route.fallback();
    });
    await test_1.test.step('Log in with the newly created user', async () => {
        await page.getByLabel('Email').fill(email);
        await page.getByLabel('Password').fill(password);
        // Click login - this will trigger navigation to authorize endpoint
        await page.getByRole('button', { name: /^Login$/i }).click();
        // Wait for redirect back to login with authRequest parameter
        await page.waitForURL(/\/login\?authRequest=/);
        // Now wait for the actual authentication flow
        const sessionReq = page.waitForRequest((req) => req.method() === 'POST' && req.url().endsWith('/api/v1/user/session'));
        const zitadelCreateReq = page.waitForRequest((req) => req.method() === 'POST' && /\/v2\/sessions$/.test(req.url()));
        const zitadelPatchReq = page.waitForRequest((req) => req.method() === 'PATCH' && /\/v2\/sessions\//.test(req.url()));
        const finalizeReq = page.waitForRequest((req) => req.method() === 'POST' && /\/v2\/oidc\/auth_requests\//.test(req.url()));
        const tokenReq = page.waitForRequest((req) => req.method() === 'POST' && /\/oauth\/v2\/token/.test(req.url()));
        // The form should auto-submit with pending credentials
        await Promise.all([sessionReq, zitadelCreateReq, zitadelPatchReq, finalizeReq, tokenReq]);
        await test_1.expect.poll(async () => {
            const authData = await page.evaluate(() => sessionStorage.getItem('shadowapi_auth'));
            if (!authData)
                return null;
            const parsed = JSON.parse(authData);
            // Verify we have JWT tokens, not session tokens
            return parsed.accessToken && parsed.idToken ? authData : null;
        }, { message: 'Auth with JWT tokens not stored in sessionStorage' }).not.toBeNull();
        await (0, test_1.expect)(page).toHaveURL('http://localtest.me/');
    });
    await test_1.test.step('Access protected content after authentication', async () => {
        await page.goto('/users');
        await (0, test_1.expect)(page).toHaveURL('http://localtest.me/users');
        await (0, test_1.expect)(page.getByRole('heading', { name: 'ShadowAPI' })).toBeVisible();
        await (0, test_1.expect)(page.getByLabel('Email')).not.toBeVisible({ timeout: 1000 });
    });
});
//# sourceMappingURL=auth-flow.test.js.map