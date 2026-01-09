import { useState, useEffect, type ReactNode } from 'react';
import { Layout, Menu, Drawer, theme, Breadcrumb, Dropdown, Button, Space, Typography } from 'antd';
import type { MenuProps } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router';
import {
  DashboardOutlined,
  MessageOutlined,
  UserOutlined,
  MailOutlined,
  LoginOutlined,
  DatabaseOutlined,
  NodeIndexOutlined,
  SettingOutlined,
  ScheduleOutlined,
  LogoutOutlined,
  DownOutlined,
  MenuOutlined,
  BookOutlined,
  SafetyOutlined,
  CrownOutlined,
  RobotOutlined,
  ThunderboltOutlined,
  KeyOutlined,
  LockOutlined,
} from '@ant-design/icons';

import { uiColors } from '../theme';
import { useAuth } from '../lib/auth';
import { useResponsive } from '../lib/useResponsive';
import { SmartLink } from '../lib/SmartLink';
import ChangePasswordModal from '../components/ChangePasswordModal';
import WorkspaceSwitcher from '../components/WorkspaceSwitcher';

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
      key: '/datasources-menu',
      icon: <MailOutlined />,
      label: 'Data Sources',
      children: [
        {
          key: '/datasources',
          icon: <MailOutlined />,
          label: <Link to={`${basePath}/datasources`}>Sources</Link>,
        },
        {
          key: '/oauth2/credentials',
          icon: <LoginOutlined />,
          label: <Link to={`${basePath}/oauth2/credentials`}>OAuth2 Credentials</Link>,
        },
      ],
    },
    {
      key: '/storages-menu',
      icon: <DatabaseOutlined />,
      label: 'Data Storages',
      children: [
        {
          key: '/storages',
          icon: <DatabaseOutlined />,
          label: <Link to={`${basePath}/storages`}>Storages</Link>,
        },
        {
          key: '/storages/messages',
          icon: <MessageOutlined />,
          label: <Link to={`${basePath}/storages/messages`}>Last Messages</Link>,
        },
      ],
    },
    {
      key: '/automation-menu',
      icon: <ThunderboltOutlined />,
      label: 'Automation',
      children: [
        {
          key: '/pipelines',
          icon: <NodeIndexOutlined />,
          label: <Link to={`${basePath}/pipelines`}>Pipelines</Link>,
        },
        {
          key: '/schedulers',
          icon: <ScheduleOutlined />,
          label: <Link to={`${basePath}/schedulers`}>Schedules</Link>,
        },
      ],
    },
    {
      key: '/workers-menu',
      icon: <SettingOutlined />,
      label: 'Workers',
      children: [
        {
          key: '/workers',
          icon: <RobotOutlined />,
          label: <Link to={`${basePath}/workers`}>Registered Workers</Link>,
        },
        {
          key: '/workers/jobs',
          icon: <ThunderboltOutlined />,
          label: <Link to={`${basePath}/workers/jobs`}>Active Jobs</Link>,
        },
        {
          key: '/workers/tokens',
          icon: <KeyOutlined />,
          label: <Link to={`${basePath}/workers/tokens`}>Enrollment Tokens</Link>,
        },
      ],
    },
    {
      key: '/rbac',
      icon: <SafetyOutlined />,
      label: 'Access Control',
      children: [
        {
          key: '/users',
          icon: <UserOutlined />,
          label: <Link to={`${basePath}/users`}>Users</Link>,
        },
        {
          key: '/invites',
          icon: <MailOutlined />,
          label: <Link to={`${basePath}/invites`}>Invites</Link>,
        },
        {
          key: '/rbac/roles',
          icon: <CrownOutlined />,
          label: <Link to={`${basePath}/rbac/roles`}>Roles</Link>,
        },
      ],
    },
    { type: 'divider' },
    {
      key: '/documentation',
      icon: <BookOutlined />,
      label: <SmartLink to="/documentation">Documentation</SmartLink>,
    },
  ];
}

interface AppLayoutProps {
  children: ReactNode;
  showSidebar?: boolean;
}

// Route configuration for breadcrumbs and menu state
interface RouteConfig {
  title: string;
  parent?: string;
}

const routeConfig: Record<string, RouteConfig> = {
  '/': { title: 'Dashboard' },
  '/users': { title: 'Users' },
  '/users/new': { title: 'Add', parent: '/users' },
  '/invites': { title: 'Invites' },
  '/rbac/roles': { title: 'Roles' },
  '/rbac/roles/new': { title: 'Create', parent: '/rbac/roles' },
  '/datasources': { title: 'Data Sources' },
  '/datasources/new': { title: 'Add', parent: '/datasources' },
  '/oauth2/credentials': { title: 'OAuth2 Credentials', parent: '/datasources' },
  '/oauth2/credentials/new': { title: 'Add', parent: '/oauth2/credentials' },
  '/storages': { title: 'Storages' },
  '/storages/new': { title: 'Add', parent: '/storages' },
  '/storages/messages': { title: 'Last Messages' },
  '/pipelines': { title: 'Pipelines' },
  '/pipelines/new': { title: 'Add', parent: '/pipelines' },
  '/workers': { title: 'Registered Workers' },
  '/workers/jobs': { title: 'Active Jobs' },
  '/workers/tokens': { title: 'Enrollment Tokens' },
  '/schedulers': { title: 'Schedules' },
  '/schedulers/new': { title: 'Add', parent: '/schedulers' },
};

// Breadcrumb map for non-workspace pages (documentation, about, etc.)
const publicBreadcrumbMap: Record<string, string> = {
  '/documentation': 'Documentation',
  '/documentation/datasource': 'Datasources',
  '/documentation/datasource/gmail': 'Gmail',
  '/documentation/datasource/telegram': 'Telegram',
  '/about': 'About',
};

// Map of routes to their menu parent keys (for expanding submenus)
const menuParentMap: Record<string, string> = {
  '/datasources': '/datasources-menu',
  '/oauth2/credentials': '/datasources-menu',
  '/storages': '/storages-menu',
  '/storages/messages': '/storages-menu',
  '/users': '/rbac',
  '/invites': '/rbac',
  '/rbac/roles': '/rbac',
  '/pipelines': '/automation-menu',
  '/schedulers': '/automation-menu',
  '/workers': '/workers-menu',
  '/workers/jobs': '/workers-menu',
  '/workers/tokens': '/workers-menu',
};

// Helper to find parent keys for a given relative path
function getOpenKeys(relativePath: string): string[] {
  // Normalize dynamic paths
  let normalizedPath = relativePath;
  if (relativePath.match(/^\/datasources\/[0-9a-f-]+$/i) || relativePath === '/datasources/new') {
    normalizedPath = '/datasources';
  } else if (relativePath.match(/^\/oauth2\/credentials\/[0-9a-f-]+$/i) || relativePath === '/oauth2/credentials/new') {
    normalizedPath = '/oauth2/credentials';
  } else if (relativePath.match(/^\/users\/[0-9a-f-]+$/i) || relativePath === '/users/new') {
    normalizedPath = '/users';
  } else if (relativePath.match(/^\/rbac\/roles\/[0-9a-f-]+$/i) || relativePath === '/rbac/roles/new') {
    normalizedPath = '/rbac/roles';
  } else if (relativePath.match(/^\/storages\/[0-9a-f-]+$/i) || relativePath === '/storages/new') {
    normalizedPath = '/storages';
  } else if (relativePath.match(/^\/pipelines\/[0-9a-f-]+$/i) || relativePath === '/pipelines/new') {
    normalizedPath = '/pipelines';
  } else if (relativePath.match(/^\/schedulers\/[0-9a-f-]+$/i) || relativePath === '/schedulers/new') {
    normalizedPath = '/schedulers';
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
  const datasourcesUuidMatch = relativePath.match(/^\/datasources\/([0-9a-f-]+)$/i);
  const oauth2UuidMatch = relativePath.match(/^\/oauth2\/credentials\/([0-9a-f-]+)$/i);
  const usersUuidMatch = relativePath.match(/^\/users\/([0-9a-f-]+)$/i);
  const rbacRolesUuidMatch = relativePath.match(/^\/rbac\/roles\/([0-9a-f-]+)$/i);
  const storagesUuidMatch = relativePath.match(/^\/storages\/([0-9a-f-]+)$/i);
  const pipelinesUuidMatch = relativePath.match(/^\/pipelines\/([0-9a-f-]+)$/i);
  const schedulersUuidMatch = relativePath.match(/^\/schedulers\/([0-9a-f-]+)$/i);
  const uuidMatch = datasourcesUuidMatch || oauth2UuidMatch || usersUuidMatch || rbacRolesUuidMatch || storagesUuidMatch || pipelinesUuidMatch || schedulersUuidMatch;

  // Determine effective path for breadcrumb chain
  let effectivePath = relativePath;
  if (datasourcesUuidMatch) {
    effectivePath = '/datasources';
  } else if (oauth2UuidMatch) {
    effectivePath = '/oauth2/credentials';
  } else if (usersUuidMatch) {
    effectivePath = '/users';
  } else if (rbacRolesUuidMatch) {
    effectivePath = '/rbac/roles';
  } else if (storagesUuidMatch) {
    effectivePath = '/storages';
  } else if (pipelinesUuidMatch) {
    effectivePath = '/pipelines';
  } else if (schedulersUuidMatch) {
    effectivePath = '/schedulers';
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

// Generate breadcrumb items for public pages (non-workspace)
function getPublicBreadcrumbItems(pathname: string): { title: React.ReactNode; key: string }[] {
  const items: { title: React.ReactNode; key: string }[] = [
    { title: <SmartLink to="/">Home</SmartLink>, key: 'home' },
  ];

  const pathSnippets = pathname.split('/').filter((i) => i);

  pathSnippets.forEach((_, index) => {
    const url = `/${pathSnippets.slice(0, index + 1).join('/')}`;
    const isLast = index === pathSnippets.length - 1;
    const name = publicBreadcrumbMap[url] || pathSnippets[index];

    items.push({
      title: isLast ? name : <Link to={url}>{name}</Link>,
      key: url,
    });
  });

  return items;
}

function AppLayout({ children, showSidebar = true }: AppLayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout, isAuthenticated, isLoading } = useAuth();
  const { isMobile } = useResponsive();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [passwordModalOpen, setPasswordModalOpen] = useState(false);
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
  if (relativePath.match(/^\/datasources\/[0-9a-f-]+$/i) || relativePath === '/datasources/new') {
    menuSelectedPath = '/datasources';
  } else if (relativePath.match(/^\/oauth2\/credentials\/[0-9a-f-]+$/i) || relativePath === '/oauth2/credentials/new') {
    menuSelectedPath = '/oauth2/credentials';
  } else if (relativePath.match(/^\/users\/[0-9a-f-]+$/i) || relativePath === '/users/new') {
    menuSelectedPath = '/users';
  } else if (relativePath.match(/^\/rbac\/roles\/[0-9a-f-]+$/i) || relativePath === '/rbac/roles/new') {
    menuSelectedPath = '/rbac/roles';
  } else if (relativePath.match(/^\/storages\/[0-9a-f-]+$/i) || relativePath === '/storages/new') {
    menuSelectedPath = '/storages';
  } else if (relativePath.match(/^\/pipelines\/[0-9a-f-]+$/i) || relativePath === '/pipelines/new') {
    menuSelectedPath = '/pipelines';
  } else if (relativePath.match(/^\/schedulers\/[0-9a-f-]+$/i) || relativePath === '/schedulers/new') {
    menuSelectedPath = '/schedulers';
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
      key: 'change-password',
      icon: <LockOutlined />,
      label: 'Change Password',
      onClick: () => setPasswordModalOpen(true),
    },
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
        {/* Hamburger button - mobile only, when sidebar is enabled */}
        {isMobile && showSidebar && (
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

        {/* Workspace Switcher - only when authenticated */}
        {!isLoading && isAuthenticated && (
          <WorkspaceSwitcher />
        )}

        {/* User dropdown or Login button */}
        {!isLoading && isAuthenticated ? (
          <Dropdown menu={{ items: userMenuItems }} trigger={['click']}>
            <Button type="text" style={{ color: '#fff' }}>
              <Space>
                <UserOutlined />
                {user?.first_name || user?.email?.split('@')[0] || 'User'}
                <DownOutlined />
              </Space>
            </Button>
          </Dropdown>
        ) : !isLoading ? (
          <Button
            type="primary"
            icon={<LoginOutlined />}
            onClick={() => navigate('/login')}
          >
            Login
          </Button>
        ) : null}
      </Header>

      {/* Main content area with sidebar */}
      <Layout hasSider>
        {/* Desktop Sider - hidden on mobile or when showSidebar is false */}
        {!isMobile && showSidebar && (
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
            <Breadcrumb items={basePath ? getBreadcrumbItems(relativePath, basePath) : getPublicBreadcrumbItems(location.pathname)} />
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

      {/* Mobile Drawer - only when sidebar is enabled */}
      {showSidebar && (
        <Drawer
          title="Navigation"
          placement="left"
          onClose={() => setDrawerOpen(false)}
          open={drawerOpen}
          styles={{ body: { padding: 0 }, wrapper: { width: 280 } }}
        >
          {sidebarMenu}
        </Drawer>
      )}

      {/* Change Password Modal */}
      <ChangePasswordModal
        open={passwordModalOpen}
        onClose={() => setPasswordModalOpen(false)}
      />

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
