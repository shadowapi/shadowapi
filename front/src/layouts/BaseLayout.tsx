import { type ReactNode, useMemo } from 'react';
import { Layout, Menu, type MenuProps, ConfigProvider } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router';

const { Header, Footer } = Layout;

// Gray color palette
const colors = {
  lightest: '#f8f9fa',
  light: '#e9ecef',
  lightMedium: '#dee2e6',
  mediumLight: '#ced4da',
  medium: '#adb5bd',
  mediumDark: '#6c757d',
  dark: '#495057',
  veryDark: '#343a40',
  darkest: '#212529',
};

const menuItems: MenuProps['items'] = [
  {
    key: '/',
    label: 'Dashboard'
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
          background: colors.veryDark,
          borderBottom: `1px solid ${colors.dark}`,
          flexShrink: 0,
        }}
      >
        <Link
          to="/"
          style={{
            height: 36,
            margin: '0 38px 0 0',
            padding: '0 18px',
            background: colors.dark,
            borderRadius: 6,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: colors.lightest,
            fontWeight: 600,
            fontSize: 18,
            letterSpacing: '0.5px',
            textDecoration: 'none',
          }}
        >
          ShadowAPI
        </Link>
        <ConfigProvider
          theme={{
            components: {
              Menu: {
                darkItemBg: 'transparent',
                darkItemColor: colors.mediumLight,
                darkItemHoverColor: colors.lightest,
                darkItemHoverBg: colors.dark,
                darkItemSelectedBg: colors.dark,
                darkItemSelectedColor: colors.lightest,
                horizontalItemBorderRadius: 6,
                horizontalItemHoverBg: colors.dark,
                horizontalItemSelectedBg: colors.dark,
              },
            },
          }}
        >
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
        </ConfigProvider>
      </Header>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        {children}
      </div>
      <Footer
        style={{
          textAlign: 'center',
          background: colors.light,
          color: colors.mediumDark,
          borderTop: `1px solid ${colors.lightMedium}`,
          flexShrink: 0,
        }}
      >
        ShadowAPI ©{new Date().getFullYear()}
      </Footer>
    </Layout>
  );
}

export default BaseLayout;
