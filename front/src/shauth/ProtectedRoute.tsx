import { ReactNode, useEffect, useState } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { Flex, ProgressCircle, Text } from '@adobe/react-spectrum'

interface ProtectedRouteProps {
  children?: ReactNode
}

interface ZitadelSession {
  sessionId: string
  factors: {
    user?: {
      verifiedAt: string
      loginName: string
    }
    password?: {
      verifiedAt: string
    }
  }
  expirationDate?: string
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const location = useLocation()

  const validateSession = async () => {
    try {
      const zitadelUrl = import.meta.env.VITE_ZITADEL_URL || 'http://auth.localtest.me'

      // Try to get current session from Zitadel
      const response = await fetch(`${zitadelUrl}/v2/sessions/_current`, {
        method: 'GET',
        headers: {
          Accept: 'application/json',
        },
        credentials: 'include', // Include cookies for session
      })

      if (!response.ok) {
        setIsAuthenticated(false)
        setIsLoading(false)
        return
      }

      const sessionData: ZitadelSession = await response.json()

      // Validate session has required factors
      const hasUserFactor = sessionData.factors.user?.verifiedAt
      const hasPasswordFactor = sessionData.factors.password?.verifiedAt

      if (!hasUserFactor || !hasPasswordFactor) {
        setIsAuthenticated(false)
        setIsLoading(false)
        return
      }

      // Check if session is expired
      if (sessionData.expirationDate) {
        const expirationTime = new Date(sessionData.expirationDate).getTime()
        const currentTime = new Date().getTime()

        if (currentTime >= expirationTime) {
          setIsAuthenticated(false)
          setIsLoading(false)
          return
        }
      }

      // Session is valid
      setIsAuthenticated(true)
      setIsLoading(false)
    } catch (error) {
      console.error('Session validation failed:', error)
      setIsAuthenticated(false)
      setIsLoading(false)
    }
  }

  useEffect(() => {
    validateSession()
  }, [])

  if (isLoading) {
    return (
      <Flex direction="column" alignItems="center" justifyContent="center" height="100vh" gap="size-200">
        <ProgressCircle aria-label="Validating session..." isIndeterminate />
        <Text>Validating session...</Text>
      </Flex>
    )
  }

  if (!isAuthenticated) {
    // Redirect to login with return URL
    const returnTo = encodeURIComponent(location.pathname + location.search)
    return <Navigate to={`/login?returnTo=${returnTo}`} replace />
  }

  return children ?? <Outlet />
}
