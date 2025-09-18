import { ReactNode, useEffect, useState } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { Flex, ProgressCircle, Text } from '@adobe/react-spectrum'

interface ProtectedRouteProps {
  children?: ReactNode
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const location = useLocation()

  const validateSession = async () => {
    // Fake session validation - simulate API call delay
    await new Promise(resolve => setTimeout(resolve, 300))

    // For now, always consider user as authenticated
    // TODO: Implement real session validation when auth system is ready
    setIsAuthenticated(true)
    setIsLoading(false)
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
