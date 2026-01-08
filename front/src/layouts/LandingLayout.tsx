import { type ReactNode, useCallback } from 'react';
import { Layout, Button } from 'antd';
import { LoginOutlined } from '@ant-design/icons';
import { uiColors } from '../theme';
import { useAuth } from '../lib/auth';
import { SmartLink } from '../lib/SmartLink';

const APP_URL = import.meta.env.VITE_APP_URL || 'http://app.localtest.me';

const { Header, Footer } = Layout;

interface LandingLayoutProps {
  children: ReactNode;
}

function LandingLayout({ children }: LandingLayoutProps) {
  const { isAuthenticated, isLoading } = useAuth();

  const handleLogin = useCallback(() => {
    window.location.href = `${APP_URL}/login`;
  }, []);

  return (
    <Layout style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          background: uiColors.headerBg,
          borderBottom: `1px solid ${uiColors.headerBorder}`,
          flexShrink: 0,
        }}
      >
        <SmartLink
          to="/"
          style={{
            height: 36,
            padding: '0 18px',
            background: uiColors.logoBg,
            borderRadius: 6,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: uiColors.logoText,
            fontWeight: 600,
            fontSize: 18,
            letterSpacing: '0.5px',
            textDecoration: 'none',
          }}
        >
          <img src="/logo.svg" alt="MeshPump logo" style={{ height: 28, marginRight: 8 }} />
          MeshPump
        </SmartLink>
        {!isLoading && !isAuthenticated && (
          <Button
            type="primary"
            icon={<LoginOutlined />}
            onClick={handleLogin}
          >
            Login
          </Button>
        )}
      </Header>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        {children}
      </div>
      <Footer
        style={{
          textAlign: 'center',
          background: uiColors.footerBg,
          color: uiColors.footerText,
          borderTop: `1px solid ${uiColors.footerBorder}`,
          flexShrink: 0,
        }}
      >
        MeshPump ©{new Date().getFullYear()}
      </Footer>
    </Layout>
  );
}

export default LandingLayout;
