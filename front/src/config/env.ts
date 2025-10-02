export const config = {
  zitadel: {
    url: import.meta.env.VITE_ZITADEL_URL,
    clientId: import.meta.env.VITE_ZITADEL_CLIENT_ID,
    redirectUri: import.meta.env.VITE_ZITADEL_REDIRECT_URL,
  },
} as const
