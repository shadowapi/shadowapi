import { type ReactNode, useMemo } from 'react';
import { Layout, Menu, type MenuProps } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router';
import { uiColors } from '../theme';

const { Header, Footer } = Layout;

const menuItems: MenuProps['items'] = [
  {
    key: '/',
    label: 'Service'
  },
  {
    key: '/page/documentation',
    label: 'Documentation'
  },
  {
    key: '/page/about',
    label: 'About'
  },
];

interface BaseLayoutProps {
  children: ReactNode;
}

function BaseLayout({ children }: BaseLayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();

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
        <Link
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
        </Link>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={selectedKeys}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{
            flex: 1,
            minWidth: 0,
            background: 'transparent',
            borderBottom: 'none',
          }}
        />
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
