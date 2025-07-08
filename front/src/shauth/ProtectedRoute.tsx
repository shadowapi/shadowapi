import { ReactNode } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { Button, Content, Flex, Heading, IllustratedMessage, ProgressCircle, Text } from '@adobe/react-spectrum'
import Timeout from '@spectrum-icons/illustrations/Timeout'
import Refresh from '@spectrum-icons/workflow/Refresh'
import { useQuery } from '@tanstack/react-query'
import { sessionOptions } from './query'

interface ProtectedRouteProps {
  children?: ReactNode
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isLoading, isError, data, error } = useQuery(sessionOptions())

  if (isLoading) {
    return (
      <Flex height="100vh" alignItems="center" justifyContent="center">
        <ProgressCircle aria-label="Loading session" isIndeterminate />
      </Flex>
    )
  }

  if (isError) {
    return (
      <Flex height="100vh" alignItems="center" justifyContent="center">
        <IllustratedMessage>
          <Timeout />
          <Heading>Something went wrong</Heading>
          <Content>
            {error instanceof Error ? `${error.name}: ${error.message}` : 'An unexpected error occurred.'}
          </Content>
          <Button variant="cta" marginTop="size-100" onPress={() => window.location.reload()}>
            <Refresh />
            <Text>Reload Page</Text>
          </Button>
        </IllustratedMessage>
      </Flex>
    )
  }

  if (data?.active) {
    return children ?? <Outlet />
  }

  return <Navigate to="/login" replace />
}
