const localFallbackPublicUrl = (() => {
  if (typeof window === 'undefined') {
    return undefined
  }
  const host = window.location.hostname
  if (host === 'localtest.me' || host.endsWith('.localtest.me')) {
    return 'http://auth.localtest.me'
  }
  return undefined
})()

export const config = {
  zitadel: {
    url: import.meta.env.VITE_ZITADEL_URL,
    publicUrl:
      import.meta.env.VITE_ZITADEL_PUBLIC_URL ??
      import.meta.env.VITE_ZITADEL_EXTERNAL_URL ??
      localFallbackPublicUrl ??
      import.meta.env.VITE_ZITADEL_URL,
    clientId: import.meta.env.VITE_ZITADEL_CLIENT_ID,
    redirectUri: import.meta.env.VITE_ZITADEL_REDIRECT_URL,
  },
} as const
