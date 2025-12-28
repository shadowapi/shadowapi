import { useState, useEffect, type ReactNode } from 'react';
import { Layout, Menu, Drawer, theme, Breadcrumb, Dropdown, Button, Space, Typography } from 'antd';
import type { MenuProps } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router';
import {
  DashboardOutlined,
  MessageOutlined,
  FileOutlined,
  UserOutlined,
  MailOutlined,
  LoginOutlined,
  DatabaseOutlined,
  ClockCircleOutlined,
  NodeIndexOutlined,
  SettingOutlined,
  ScheduleOutlined,
  UnorderedListOutlined,
  LogoutOutlined,
  DownOutlined,
  MenuOutlined,
} from '@ant-design/icons';

import { uiColors } from '../theme';
import { useAuth } from '../lib/auth';
import { useResponsive } from '../lib/useResponsive';
import { SmartLink } from '../lib/SmartLink';

const { Header, Sider, Content, Footer } = Layout;

type MenuItem = Required<MenuProps>['items'][number];

// Helper to extract workspace info from pathname
function getWorkspaceInfo(pathname: string): { basePath: string; relativePath: string } {
  const match = pathname.match(/^\/w\/([^/]+)(.*)/);
  if (match) {
    return {
      basePath: `/w/${match[1]}`,
      relativePath: match[2] || '/',
    };
  }
  return { basePath: '', relativePath: pathname };
}

// Generate menu items with workspace-aware paths
function getMenuItems(basePath: string): MenuItem[] {
  return [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: <Link to={`${basePath}/`}>Dashboard</Link>,
    },
    {
      key: '/messages',
      icon: <MessageOutlined />,
      label: <Link to={`${basePath}/messages`}>Messages</Link>,
      children: [
        {
          key: '/files',
          icon: <FileOutlined />,
          label: <Link to={`${basePath}/files`}>Files</Link>,
        },
      ],
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: <Link to={`${basePath}/users`}>Users</Link>,
    },
    {
      key: '/datasources',
      icon: <MailOutlined />,
      label: <Link to={`${basePath}/datasources`}>Data Sources</Link>,
      children: [
        {
          key: '/oauth2/credentials',
          icon: <LoginOutlined />,
          label: <Link to={`${basePath}/oauth2/credentials`}>OAuth2 Credentials</Link>,
        },
      ],
    },
    {
      key: '/storages',
      icon: <DatabaseOutlined />,
      label: <Link to={`${basePath}/storages`}>Data Storages</Link>,
    },
    {
      key: '/syncpolicies',
      icon: <ClockCircleOutlined />,
      label: <Link to={`${basePath}/syncpolicies`}>Sync Policies</Link>,
    },
    {
      key: '/pipelines',
      icon: <NodeIndexOutlined />,
      label: <Link to={`${basePath}/pipelines`}>Data Pipelines</Link>,
    },
    {
      key: '/workers',
      icon: <SettingOutlined />,
      label: <Link to={`${basePath}/workers`}>Workers</Link>,
      children: [
        {
          key: '/schedulers',
          icon: <ScheduleOutlined />,
          label: <Link to={`${basePath}/schedulers`}>Schedulers</Link>,
        },
      ],
    },
    {
      key: '/logs',
      icon: <UnorderedListOutlined />,
      label: <Link to={`${basePath}/logs`}>Logs</Link>,
    },
  ];
}

interface AppLayoutProps {
  children: ReactNode;
}

// Route configuration for breadcrumbs and menu state
interface RouteConfig {
  title: string;
  parent?: string;
}

const routeConfig: Record<string, RouteConfig> = {
  '/': { title: 'Dashboard' },
  '/messages': { title: 'Messages' },
  '/files': { title: 'Files', parent: '/messages' },
  '/users': { title: 'Users' },
  '/users/new': { title: 'Add', parent: '/users' },
  '/oauth2/credentials': { title: 'OAuth2 Credentials' },
  '/oauth2/credentials/new': { title: 'Add', parent: '/oauth2/credentials' },
  '/storages': { title: 'Data Storages' },
  '/syncpolicies': { title: 'Sync Policies' },
  '/pipelines': { title: 'Data Pipelines' },
  '/workers': { title: 'Workers' },
  '/schedulers': { title: 'Schedulers', parent: '/workers' },
  '/logs': { title: 'Logs' },
};

// Map of routes to their menu parent keys (for expanding submenus)
const menuParentMap: Record<string, string> = {
  '/files': '/messages',
  '/oauth2/credentials': '/datasources',
  '/schedulers': '/workers',
};

// Helper to find parent keys for a given relative path
function getOpenKeys(relativePath: string): string[] {
  // Normalize dynamic paths
  let normalizedPath = relativePath;
  if (relativePath.match(/^\/oauth2\/credentials\/[0-9a-f-]+$/i) || relativePath === '/oauth2/credentials/new') {
    normalizedPath = '/oauth2/credentials';
  } else if (relativePath.match(/^\/users\/[0-9a-f-]+$/i) || relativePath === '/users/new') {
    normalizedPath = '/users';
  }

  const parentKey = menuParentMap[normalizedPath];
  return parentKey ? [parentKey] : [];
}

// Generate breadcrumb items for a given path
function getBreadcrumbItems(relativePath: string, basePath: string): { title: React.ReactNode; key: string }[] {
  const items: { title: React.ReactNode; key: string }[] = [
    { title: <Link to={`${basePath}/`}>Service</Link>, key: 'service' },
  ];

  // Check if this is an edit page (uuid pattern)
  const oauth2UuidMatch = relativePath.match(/^\/oauth2\/credentials\/([0-9a-f-]+)$/i);
  const usersUuidMatch = relativePath.match(/^\/users\/([0-9a-f-]+)$/i);
  const uuidMatch = oauth2UuidMatch || usersUuidMatch;

  // Determine effective path for breadcrumb chain
  let effectivePath = relativePath;
  if (oauth2UuidMatch) {
    effectivePath = '/oauth2/credentials';
  } else if (usersUuidMatch) {
    effectivePath = '/users';
  }

  // Build the breadcrumb chain by following parent links
  const chain: string[] = [];
  let currentPath = effectivePath;

  while (currentPath) {
    chain.unshift(currentPath);
    const config = routeConfig[currentPath];
    currentPath = config?.parent || '';
  }

  // Add each item in the chain
  chain.forEach((path, index) => {
    const config = routeConfig[path];
    if (!config) return;

    const isLast = index === chain.length - 1 && !uuidMatch;
    if (isLast) {
      items.push({ title: config.title, key: path });
    } else {
      items.push({ title: <Link to={`${basePath}${path}`}>{config.title}</Link>, key: path });
    }
  });

  // Add "Edit" for uuid pages
  if (uuidMatch) {
    items.push({ title: 'Edit', key: 'edit' });
  }

  return items;
}

function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { isMobile } = useResponsive();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // Close drawer when switching to desktop
  useEffect(() => {
    if (!isMobile) {
      setDrawerOpen(false);
    }
  }, [isMobile]);

  // Extract workspace info from path
  const { basePath, relativePath } = getWorkspaceInfo(location.pathname);

  const handleLogout = async () => {
    await logout();
    navigate('/workspaces');
  };

  // Normalize path for menu selection (edit/new pages should highlight parent)
  let menuSelectedPath = relativePath;
  if (relativePath.match(/^\/oauth2\/credentials\/[0-9a-f-]+$/i) || relativePath === '/oauth2/credentials/new') {
    menuSelectedPath = '/oauth2/credentials';
  } else if (relativePath.match(/^\/users\/[0-9a-f-]+$/i) || relativePath === '/users/new') {
    menuSelectedPath = '/users';
  }
  const selectedKeys = [menuSelectedPath];
  const defaultOpenKeys = getOpenKeys(relativePath);
  const menuItems = getMenuItems(basePath);

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'email',
      label: (
        <Typography.Text type="secondary" style={{ cursor: 'default' }}>
          {user?.email}
        </Typography.Text>
      ),
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Sign out',
      onClick: handleLogout,
    },
  ];

  // Handle menu click - close drawer on mobile
  const handleMenuClick = () => {
    if (isMobile) {
      setDrawerOpen(false);
    }
  };

  // Sidebar menu component (reused in Sider and Drawer)
  const sidebarMenu = (
    <Menu
      mode="inline"
      selectedKeys={selectedKeys}
      defaultOpenKeys={defaultOpenKeys}
      style={{ height: '100%', borderRight: 0 }}
      items={menuItems}
      onClick={handleMenuClick}
    />
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* Sticky Header */}
      <Header
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 1001,
          display: 'flex',
          alignItems: 'center',
          background: uiColors.headerBg,
          borderBottom: `1px solid ${uiColors.headerBorder}`,
          padding: '0 16px',
        }}
      >
        {/* Hamburger button - mobile only */}
        {isMobile && (
          <Button
            type="text"
            icon={<MenuOutlined />}
            onClick={() => setDrawerOpen(true)}
            style={{ color: '#fff', marginRight: 16, fontSize: 18 }}
          />
        )}

        {/* Logo */}
        <SmartLink
          to="/"
          style={{
            height: 36,
            margin: '0 24px 0 0',
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

        {/* Spacer */}
        <div style={{ flex: 1 }} />

        {/* User dropdown */}
        <Dropdown menu={{ items: userMenuItems }} trigger={['click']}>
          <Button type="text" style={{ color: '#fff' }}>
            <Space>
              <UserOutlined />
              {user?.first_name || user?.email?.split('@')[0] || 'User'}
              <DownOutlined />
            </Space>
          </Button>
        </Dropdown>
      </Header>

      {/* Main content area with sidebar */}
      <Layout hasSider>
        {/* Desktop Sider - hidden on mobile */}
        {!isMobile && (
          <Sider
            width={250}
            style={{
              background: colorBgContainer,
              overflow: 'auto',
              height: 'calc(100vh - 64px)',
              position: 'sticky',
              top: 64,
              left: 0,
            }}
          >
            {sidebarMenu}
          </Sider>
        )}

        {/* Content area */}
        <Layout style={{ padding: isMobile ? '16px' : '0 24px 24px' }}>
          {/* Breadcrumb row */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              margin: '16px 0',
            }}
          >
            <Breadcrumb items={getBreadcrumbItems(relativePath, basePath)} />
          </div>

          {/* Main content */}
          <Content
            style={{
              padding: 24,
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
              minHeight: 280,
            }}
          >
            {children}
          </Content>
        </Layout>
      </Layout>

      {/* Mobile Drawer */}
      <Drawer
        title="Navigation"
        placement="left"
        onClose={() => setDrawerOpen(false)}
        open={drawerOpen}
        styles={{ body: { padding: 0 }, wrapper: { width: 280 } }}
      >
        {sidebarMenu}
      </Drawer>

      {/* Footer */}
      <Footer
        style={{
          textAlign: 'center',
          background: uiColors.footerBg,
          color: uiColors.footerText,
          borderTop: `1px solid ${uiColors.footerBorder}`,
        }}
      >
        MeshPump {new Date().getFullYear()}
      </Footer>
    </Layout>
  );
}

export default AppLayout;
