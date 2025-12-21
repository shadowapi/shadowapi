import { useEffect } from 'react';
import { useNavigate } from 'react-router';
import { Spin } from 'antd';
import { useAuth } from '../lib/auth';

/**
 * RootRedirect handles the root path on the app subdomain.
 * - Authenticated users → /workspaces
 * - Unauthenticated users → /login
 *
 * Note: This runs on app.{domain}, not on the root domain.
 * The root domain serves SSR content directly.
 */
function RootRedirect() {
  const navigate = useNavigate();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        navigate('/workspaces', { replace: true });
      } else {
        navigate('/login', { replace: true });
      }
    }
  }, [isAuthenticated, isLoading, navigate]);

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
      <Spin size="large" />
    </div>
  );
}

export default RootRedirect;
