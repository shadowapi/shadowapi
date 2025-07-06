import { useCallback } from 'react'

export const useLogout = () =>
  useCallback(() => {
    const logoutUrl = `${import.meta.env.VITE_ZITADEL_INSTANCE_URL}/oidc/v2/logout?post_logout_redirect_uri=${encodeURIComponent(import.meta.env.VITE_ZITADEL_REDIRECT_URI)}`
    window.location.href = logoutUrl
  }, [])
