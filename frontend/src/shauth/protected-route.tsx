import { ReactNode } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { Button, Result, Spin } from 'antd'
import { useSession } from './query'

interface ProtectedRouteProps {
  children?: ReactNode
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { data, error, isLoading } = useSession()
  const navigate = useNavigate()

  if (isLoading || (!data && !error)) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (error || !data?.active) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
        <Result
          status="warning"
          title="Session Expired"
          subTitle="Your session has expired or is invalid. Please log in again."
          extra={
            <Button type="primary" onClick={() => navigate('/login')}>
              Login
            </Button>
          }
        />
      </div>
    )
  }

  return <>{children ?? <Outlet />}</>
}
