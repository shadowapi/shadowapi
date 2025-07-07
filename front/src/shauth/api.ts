import { useCallback } from 'react'

export const useLogout = () =>
  useCallback(() => {
    const redirect = `${window.location.origin}/logout/callback`
    const base = import.meta.env.VITE_ZITADEL_INSTANCE_URL
    if (base) {
      const logoutUrl = `${base}/oidc/v2/logout?post_logout_redirect_uri=${encodeURIComponent(redirect)}`
      window.location.href = logoutUrl
    } else {
      window.location.href = '/logout/callback'
    }
  }, [])
