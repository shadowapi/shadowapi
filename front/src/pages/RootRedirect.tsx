import { useEffect } from 'react';
import { useNavigate } from 'react-router';
import { Spin } from 'antd';
import { useAuth } from '../lib/auth';

function RootRedirect() {
  const navigate = useNavigate();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        navigate('/workspaces', { replace: true });
      } else {
        navigate('/page/start', { replace: true });
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
