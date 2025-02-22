import { useCallback } from 'react'
import { Configuration, FrontendApi } from '@ory/client'

export const FrontendAPI = new FrontendApi(
  new Configuration({
    basePath: '/auth/user',
    baseOptions: {
      withCredentials: true,
    },
  })
)

export const useLogout = () =>
  useCallback(async () => {
    try {
      const response = await FrontendAPI.createBrowserLogoutFlow()
      const logoutUrl = response?.data?.logout_url

      if (logoutUrl) {
        window.location.href = logoutUrl
      } else {
        console.warn('Logout URL not provided, falling back to manual session cleanup.')
        localStorage.clear()
        sessionStorage.clear()
        window.location.href = '/login' // Fallback redirect
      }
    } catch (error) {
      console.error('Logout failed:', error)
      localStorage.clear()
      sessionStorage.clear()
      window.location.href = '/login' // Fallback in case of API failure
    }
  }, [])
