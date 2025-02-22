import { FC, ReactNode } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { sessionOptions } from './query'

interface ProtectedRouteProps {
  children?: ReactNode
}

export const ProtectedRoute: FC<ProtectedRouteProps> = ({ children }) => {
  const { isError, isSuccess, data, error } = useQuery(sessionOptions())

  if (isSuccess) {
    if (data.active) {
      return children ? <>{children}</> : <Outlet />
    }
  } else if (isError) {
    console.error('error fetching session from auth server', error)
    return (
      <div>
        Error '{error.name}': {error.message}
      </div>
    )
  } else {
    return <></>
  }
  return <Navigate to="/login" replace />
}
