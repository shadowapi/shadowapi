import { type ReactNode, useEffect, useRef } from 'react';
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

// Helper to get the base domain (e.g., localtest.me from internal.localtest.me)
const getBaseDomain = (): string => {
  const hostname = window.location.hostname;
  const parts = hostname.split('.');
  if (parts.length >= 2) {
    return parts.slice(-2).join('.');
  }
  return hostname;
};

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, tenantNotFound } = useAuth();
  const location = useLocation();
  const redirectingRef = useRef(false);

  // Handle redirect for non-existent tenant via effect
  useEffect(() => {
    if (tenantNotFound && !redirectingRef.current) {
      redirectingRef.current = true;
      const baseDomain = getBaseDomain();
      window.location.href = `${window.location.protocol}//${baseDomain}/page/tenant`;
    }
  }, [tenantNotFound]);

  if (isLoading || tenantNotFound) {
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

  // Always redirect root domain users to tenant selection
  // (even if authenticated - root domain has no tenant context)
  if (isRootDomain()) {
    return <Navigate to="/page/tenant" state={{ from: location }} replace />;
  }

  if (!isAuthenticated) {
    // On tenant subdomain, redirect to login
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
}
