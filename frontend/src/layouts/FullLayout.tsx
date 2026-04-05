import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Layout, Menu, Dropdown, Button, theme } from 'antd'
import type { MenuProps } from 'antd'
import {
  HomeOutlined,
  MailOutlined,
  FileOutlined,
  UserOutlined,
  DatabaseOutlined,
  KeyOutlined,
  HddOutlined,
  SyncOutlined,
  NodeIndexOutlined,
  ThunderboltOutlined,
  ScheduleOutlined,
  UnorderedListOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  ProfileOutlined,
  AppstoreOutlined,
  CloudOutlined,
  SettingOutlined,
} from '@ant-design/icons'

import { useSession } from '@/shauth/query'
import { useLogout } from '@/shauth/api'

const { Header, Sider, Content } = Layout

const menuItems: MenuProps['items'] = [
  { key: '/', icon: <HomeOutlined />, label: 'Dashboard' },
  {
    key: 'data',
    icon: <AppstoreOutlined />,
    label: 'Data',
    children: [
      { key: '/messages', icon: <MailOutlined />, label: 'Messages' },
      { key: '/files', icon: <FileOutlined />, label: 'Files' },
      { key: '/datasources', icon: <DatabaseOutlined />, label: 'Data Sources' },
      { key: '/storages', icon: <HddOutlined />, label: 'Storages' },
    ],
  },
  {
    key: 'processing',
    icon: <CloudOutlined />,
    label: 'Processing',
    children: [
      { key: '/pipelines', icon: <NodeIndexOutlined />, label: 'Pipelines' },
      { key: '/syncpolicies', icon: <SyncOutlined />, label: 'Sync Policies' },
      { key: '/workers', icon: <ThunderboltOutlined />, label: 'Workers' },
      { key: '/schedulers', icon: <ScheduleOutlined />, label: 'Schedulers' },
    ],
  },
  {
    key: 'admin',
    icon: <SettingOutlined />,
    label: 'Admin',
    children: [
      { key: '/users', icon: <UserOutlined />, label: 'Users' },
      { key: '/oauth2/credentials', icon: <KeyOutlined />, label: 'OAuth2 Credentials' },
      { key: '/logs', icon: <UnorderedListOutlined />, label: 'Logs' },
    ],
  },
]

export function FullLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { data: session } = useSession()

  const allGroupKeys = ['data', 'processing', 'admin']
  const logout = useLogout()

  const {
    token: { colorBgContainer },
  } = theme.useToken()

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <ProfileOutlined />,
      label: 'Profile',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Logout',
      onClick: () => logout(),
    },
  ]

  const onMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsed={collapsed}
        breakpoint="lg"
        onBreakpoint={(broken) => setCollapsed(broken)}
        trigger={null}
        style={{ background: colorBgContainer }}
      >
        <div
          style={{
            height: 56,
            padding: '0 8px',
            display: 'flex',
            alignItems: 'center',
            gap: 4,
          }}
        >
          <img src="/logo.png" alt="ShadowAPI" style={{ width: 40, height: 40 }} />
          {!collapsed && (
            <span style={{ fontWeight: 600, fontSize: 16 }}>ShadowAPI</span>
          )}
          <div style={{ flex: 1 }} />
          <Button
            type="text"
            size="small"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
        </div>
        <Menu mode="inline" selectedKeys={[location.pathname]} defaultOpenKeys={allGroupKeys} items={menuItems} onClick={onMenuClick} />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 16px',
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            height: 40,
            lineHeight: '40px',
          }}
        >
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Button type="text" size="small" icon={<UserOutlined />}>
              {session?.email ?? 'User'}
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ margin: 16 }}>{children}</Content>
      </Layout>
    </Layout>
  )
}
