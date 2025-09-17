import { ReactNode } from 'react'
import { Outlet } from 'react-router-dom'

interface ProtectedRouteProps {
  children?: ReactNode
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  // Temporarily disabled session check - just render content without protection
  return children ?? <Outlet />
}
