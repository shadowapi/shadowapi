import { type ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router';
import { Spin } from 'antd';
import { useAuth } from './useAuth';

interface ProtectedRouteProps {
  children: ReactNode;
}

// Helper to check if we're on root domain (no subdomain)
const isRootDomain = (): boolean => {
  const hostname = window.location.hostname;
  const parts = hostname.split('.');
  // Root domain has exactly 2 parts (e.g., localtest.me)
  // Subdomain has 3+ parts (e.g., acme.localtest.me)
  return parts.length <= 2;
};

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '100vh',
        }}
      >
        <Spin size="large" />
      </div>
    );
  }

  if (!isAuthenticated) {
    // On root domain, redirect to tenant selection
    if (isRootDomain()) {
      return <Navigate to="/page/tenant" state={{ from: location }} replace />;
    }
    // On tenant subdomain, redirect to login
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
}
