export const config = {
  zitadel: {
    url: import.meta.env.VITE_ZITADEL_URL || 'http://auth.localtest.me',
    clientId: import.meta.env.VITE_ZITADEL_CLIENT_ID || '339013429979316232@shadowapi',
    redirectUri: import.meta.env.VITE_ZITADEL_REDIRECT_URI || 'http://localtest.me/auth/callback',
  },
} as const
