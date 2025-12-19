import { type ReactNode, useMemo, useCallback } from 'react';
import { Layout, Menu, Button, type MenuProps } from 'antd';
import { useLocation, useNavigate } from 'react-router';
import { LoginOutlined } from '@ant-design/icons';
import { uiColors } from '../theme';
import { useAuth } from '../lib/auth';
import { SmartLink } from '../lib/SmartLink';

// Subdomain URLs from environment
const WWW_BASE_URL = import.meta.env.VITE_WWW_BASE_URL || 'http://www.localtest.me';
const APP_BASE_URL = import.meta.env.VITE_APP_BASE_URL || 'http://localtest.me';

// SSR routes that live on www subdomain
const SSR_PATHS = ['/start', '/about', '/documentation'];

// Check if a path is an SSR route (www subdomain)
function isSSRPath(path: string): boolean {
  return SSR_PATHS.some((p) => path === p || path.startsWith(p + '/'));
}

// Check if a path is an app route (root domain)
function isAppPath(path: string): boolean {
  return (
    path === '/' ||
    path.startsWith('/workspaces') ||
    path.startsWith('/w/') ||
    path.startsWith('/login')
  );
}

// Get current subdomain context
function getCurrentContext(): 'www' | 'app' {
  if (typeof window === 'undefined') return 'app';
  const hostname = window.location.hostname;
  return hostname.startsWith('www.') ? 'www' : 'app';
}

const { Header, Footer } = Layout;

const menuItems: MenuProps['items'] = [
  {
    key: '/',
    label: 'Service'
  },
  {
    key: '/documentation',
    label: 'Documentation'
  },
  {
    key: '/about',
    label: 'About'
  },
];

interface BaseLayoutProps {
  children: ReactNode;
}

function BaseLayout({ children }: BaseLayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { isAuthenticated, isLoading } = useAuth();

  // Smart navigation that handles cross-subdomain routing
  const smartNavigate = useCallback((path: string) => {
    const currentContext = getCurrentContext();
    const targetIsSSR = isSSRPath(path);
    const targetIsApp = isAppPath(path);

    // Cross-subdomain navigation requires full page redirect
    if (currentContext === 'app' && targetIsSSR) {
      window.location.href = `${WWW_BASE_URL}${path}`;
      return;
    }

    if (currentContext === 'www' && targetIsApp) {
      window.location.href = `${APP_BASE_URL}${path}`;
      return;
    }

    // Same-subdomain navigation uses React Router
    navigate(path);
  }, [navigate]);

  const selectedKeys = useMemo(() => {
    const pathname = location.pathname;

    // Find the menu item whose key is a prefix of the current path
    for (const item of menuItems || []) {
      if (item && 'key' in item && typeof item.key === 'string') {
        // Exact match for root path
        if (item.key === '/' && pathname === '/') {
          return [item.key];
        }
        // Prefix match for other paths (but not for root)
        if (item.key !== '/' && pathname.startsWith(item.key)) {
          return [item.key];
        }
      }
    }

    return [];
  }, [location.pathname]);

  return (
    <Layout style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          background: uiColors.headerBg,
          borderBottom: `1px solid ${uiColors.headerBorder}`,
          flexShrink: 0,
        }}
      >
        <SmartLink
          to="/"
          style={{
            height: 36,
            margin: '0 38px 0 0',
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
          ShadowAPI
        </SmartLink>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={selectedKeys}
          items={menuItems}
          onClick={({ key }) => smartNavigate(key)}
          style={{
            flex: 1,
            minWidth: 0,
            background: 'transparent',
            borderBottom: 'none',
          }}
        />
        {!isLoading && !isAuthenticated && (
          <Button
            type="primary"
            icon={<LoginOutlined />}
            onClick={() => smartNavigate('/login')}
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
        ShadowAPI ©{new Date().getFullYear()}
      </Footer>
    </Layout>
  );
}

export default BaseLayout;
