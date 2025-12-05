import { type ReactNode } from 'react'
import { Layout, Menu, theme, Breadcrumb, Dropdown, Button, Space, Typography } from 'antd'
import type { MenuProps } from 'antd'
import { Link, useLocation } from 'react-router'
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
} from '@ant-design/icons'

import BaseLayout from './BaseLayout'
import { useAuth } from '../lib/auth'


const { Sider, Content } = Layout

type MenuItem = Required<MenuProps>['items'][number]

const menuItems: MenuItem[] = [
  {
    key: '/',
    icon: <DashboardOutlined />,
    label: <Link to="/">Dashboard</Link>,
  },
  {
    key: '/messages',
    icon: <MessageOutlined />,
    label: <Link to="/messages">Messages</Link>,
    children: [
      {
        key: '/files',
        icon: <FileOutlined />,
        label: <Link to="/files">Files</Link>,
      },
    ],
  },
  {
    key: '/users',
    icon: <UserOutlined />,
    label: <Link to="/users">Users</Link>,
  },
  {
    key: '/datasources',
    icon: <MailOutlined />,
    label: <Link to="/datasources">Data Sources</Link>,
    children: [
      {
        key: '/oauth2/credentials',
        icon: <LoginOutlined />,
        label: <Link to="/oauth2/credentials">OAuth2 Credentials</Link>,
      },
    ],
  },
  {
    key: '/storages',
    icon: <DatabaseOutlined />,
    label: <Link to="/storages">Data Storages</Link>,
  },
  {
    key: '/syncpolicies',
    icon: <ClockCircleOutlined />,
    label: <Link to="/syncpolicies">Sync Policies</Link>,
  },
  {
    key: '/pipelines',
    icon: <NodeIndexOutlined />,
    label: <Link to="/pipelines">Data Pipelines</Link>,
  },
  {
    key: '/workers',
    icon: <SettingOutlined />,
    label: <Link to="/workers">Workers</Link>,
    children: [
      {
        key: '/schedulers',
        icon: <ScheduleOutlined />,
        label: <Link to="/schedulers">Schedulers</Link>,
      },
    ],
  },
  {
    key: '/logs',
    icon: <UnorderedListOutlined />,
    label: <Link to="/logs">Logs</Link>,
  },
];

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

// Helper to find parent keys for a given path
function getOpenKeys(pathname: string): string[] {
  // Normalize dynamic paths
  let normalizedPath = pathname;
  if (pathname.match(/^\/oauth2\/credentials\/[0-9a-f-]+$/i) || pathname === '/oauth2/credentials/new') {
    normalizedPath = '/oauth2/credentials';
  }

  const parentKey = menuParentMap[normalizedPath];
  return parentKey ? [parentKey] : [];
}

// Generate breadcrumb items for a given path
function getBreadcrumbItems(pathname: string): { title: React.ReactNode; key: string }[] {
  const items: { title: React.ReactNode; key: string }[] = [
    { title: <Link to="/">Service</Link>, key: 'service' },
  ];

  // Check if this is an edit page (uuid pattern)
  const uuidMatch = pathname.match(/^\/oauth2\/credentials\/([0-9a-f-]+)$/i);
  const effectivePath = uuidMatch ? '/oauth2/credentials/:uuid' : pathname;

  // Build the breadcrumb chain by following parent links
  const chain: string[] = [];
  let currentPath = effectivePath === '/oauth2/credentials/:uuid' ? '/oauth2/credentials' : effectivePath;

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
      items.push({ title: <Link to={path}>{config.title}</Link>, key: path });
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
  const { user, logout } = useAuth();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // Normalize path for menu selection (edit/new pages should highlight parent)
  let menuSelectedPath = location.pathname;
  if (location.pathname.match(/^\/oauth2\/credentials\/[0-9a-f-]+$/i) || location.pathname === '/oauth2/credentials/new') {
    menuSelectedPath = '/oauth2/credentials';
  }
  const selectedKeys = [menuSelectedPath];
  const defaultOpenKeys = getOpenKeys(location.pathname);

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
      onClick: logout,
    },
  ];

  return (
    <BaseLayout>
      <div
        style={{
          padding: '0 48px',
          display: 'flex',
          flexDirection: 'column',
          flex: 1,
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '16px 0', flexShrink: 0 }}>
          <Breadcrumb items={getBreadcrumbItems(location.pathname)} />
          <Dropdown menu={{ items: userMenuItems }} trigger={['click']}>
            <Button type="text">
              <Space>
                <UserOutlined />
                {user?.first_name || user?.email?.split('@')[0] || 'User'}
                <DownOutlined />
              </Space>
            </Button>
          </Dropdown>
        </div>
        <Layout
          style={{
            padding: '24px 0',
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
            flex: 1,
            marginBottom: 24,
          }}
        >
          <Sider style={{ background: colorBgContainer }} width={250}>
            <Menu
              mode="inline"
              selectedKeys={selectedKeys}
              defaultOpenKeys={defaultOpenKeys}
              style={{ height: '100%' }}
              items={menuItems}
            />
          </Sider>
          <Content style={{ padding: '0 24px' }}>
            {children}
          </Content>
        </Layout>
      </div>
    </BaseLayout>
  );
}

export default AppLayout;
