import { useEffect } from 'react';
import { useNavigate } from 'react-router';
import { Spin } from 'antd';
import { useAuth } from '../lib/auth';

// WWW subdomain URL for SSR pages
const WWW_BASE_URL =
  import.meta.env.VITE_WWW_BASE_URL || 'http://www.localtest.me'

function RootRedirect() {
  const navigate = useNavigate();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        navigate('/workspaces', { replace: true });
      } else {
        // Redirect to www subdomain for landing page
        window.location.href = `${WWW_BASE_URL}/start`;
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
