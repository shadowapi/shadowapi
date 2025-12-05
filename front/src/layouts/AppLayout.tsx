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

// Helper to find parent keys for a given path
function getOpenKeys(pathname: string): string[] {
  const openKeys: string[] = [];

  // Map of child paths to their parent keys
  const parentMap: Record<string, string> = {
    '/files': '/messages',
    '/oauth2/credentials': '/datasources',
    '/schedulers': '/workers',
  };

  if (parentMap[pathname]) {
    openKeys.push(parentMap[pathname]);
  }

  return openKeys;
}

function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation();
  const { user, logout } = useAuth();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const selectedKeys = [location.pathname];
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
          <Breadcrumb
            items={[
              { title: <Link to="/">Service</Link> },
              { title: 'Dashboard' },
            ]}
          />
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
