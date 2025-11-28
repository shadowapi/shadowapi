import { Page } from '@playwright/test'

export async function setSessionAuth(page: Page, opts?: { email?: string; accessToken?: string; idToken?: string; refreshToken?: string; expiresInSec?: number }) {
  const {
    email = 'admin@example.com',
    accessToken = 'test-access-token',
    idToken = 'test-id-token',
    refreshToken = '',
    expiresInSec = 3600,
  } = opts || {}

  await page.addInitScript(({ email, accessToken, idToken, refreshToken, expiresInSec }) => {
    try {
      const data = {
        email,
        accessToken,
        idToken,
        refreshToken,
        expiresAt: Date.now() + expiresInSec * 1000,
      }
      sessionStorage.setItem('shadowapi_auth', JSON.stringify(data))
    } catch (e) {
      // ignore
    }
  }, { email, accessToken, idToken, refreshToken, expiresInSec })
}

