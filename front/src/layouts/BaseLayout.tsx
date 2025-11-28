import { type ReactNode, useMemo } from 'react';
import { Layout, Menu, type MenuProps, theme } from 'antd';
import { useLocation, useNavigate } from 'react-router';

const { Header, Footer } = Layout;

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
  const {
    token: { borderRadiusLG },
  } = theme.useToken();

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
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center' }}>
        <div
          style={{
            height: 32,
            margin: '0 38px 0 0',
            padding: '0 12px',
            background: 'rgba(255, 255, 255, 0.2)',
            borderRadius: borderRadiusLG,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontWeight: 'bold',
            overflow: 'hidden',
          }}
        >
          {'ShadowAPI'}
        </div>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={selectedKeys}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ flex: 1, minWidth: 0 }}
        />
      </Header>
      {children}
      <Footer style={{ textAlign: 'center' }}>
        ShadowAPI ©{new Date().getFullYear()}
      </Footer>
    </Layout>
  );
};

export default BaseLayout;
