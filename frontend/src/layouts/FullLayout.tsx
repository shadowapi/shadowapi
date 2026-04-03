import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Layout, Menu, Dropdown, Button, Typography, theme } from 'antd'
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
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        breakpoint="lg"
        style={{ background: colorBgContainer }}
      >
        <div
          style={{
            height: 32,
            margin: 16,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {!collapsed && (
            <Typography.Title level={4} style={{ margin: 0 }}>
              ShadowAPI
            </Typography.Title>
          )}
        </div>
        <Menu mode="inline" selectedKeys={[location.pathname]} defaultOpenKeys={allGroupKeys} items={menuItems} onClick={onMenuClick} />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Button type="text" icon={<UserOutlined />}>
              {session?.email ?? 'User'}
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ margin: 16 }}>{children}</Content>
      </Layout>
    </Layout>
  )
}
