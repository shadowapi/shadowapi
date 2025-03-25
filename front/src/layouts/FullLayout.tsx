import { ReactNode } from 'react'
import { ActionButton, Flex, Heading, Item, Menu, MenuTrigger, View } from '@adobe/react-spectrum'
import AssetsExpired from '@spectrum-icons/workflow/AssetsExpired'
import Data from '@spectrum-icons/workflow/Data'
import EmailGear from '@spectrum-icons/workflow/EmailGear'
import Gears from '@spectrum-icons/workflow/Gears'
import Homepage from '@spectrum-icons/workflow/Homepage'
import Login from '@spectrum-icons/workflow/Login'
import Organize from '@spectrum-icons/workflow/Organize'
import User from '@spectrum-icons/workflow/User'
import Workflow from '@spectrum-icons/workflow/Workflow'

import { Navbar } from '@/components/Navbar'
import { useLogout } from '@/shauth'

export function FullLayout({ children }: { children: ReactNode }) {
  const logout = useLogout()
  // Main navigation items
  const navItems = [
    { Label: 'Dashboard', AriaLabel: 'Go to dashboard page', Icon: <Homepage />, URL: '/' },
    { Label: 'Users', AriaLabel: 'Go to data users page', Icon: <User />, URL: '/users' },
    { Label: 'Data Sources', AriaLabel: 'Go to data sources page', Icon: <EmailGear />, URL: '/datasources' },
    { Label: 'Data Pipelines', AriaLabel: 'Go to data pipelines page', Icon: <Workflow />, URL: '/pipelines' },
    { Label: 'Data Storages', AriaLabel: 'Go to data storages page', Icon: <Data />, URL: '/storages' },
    {
      Label: 'OAuth2 Credentials',
      AriaLabel: 'Go to OAuth2 credentials page',
      Icon: <Login />,
      URL: '/oauth2/credentials',
    },
    { Label: 'Workers', AriaLabel: 'Go to workers page', Icon: <Gears />, URL: '/workers' },
    { Label: 'SyncPolicies', AriaLabel: 'Go to sync policies page', Icon: <AssetsExpired />, URL: '/syncpolicies' },
    { Label: 'Logs', AriaLabel: 'Go to logs page', Icon: <Organize />, URL: '/logs' },
  ]

  return (
    <Flex direction="column" height="100vh">
      <View height="size-800" colorVersion={6} backgroundColor="gray-200">
        <Flex justifyContent="space-between" height="100%">
          <Heading level={2} justifySelf="start" marginX="size-600">
            ShadowAPI
          </Heading>
          <View justifySelf="end" alignSelf="center" marginEnd="size-600">
            <MenuTrigger align="start">
              <ActionButton>
                <User />
              </ActionButton>
              <Menu onAction={(key) => key === 'logout' && logout()}>
                <Item key="edit-profile">Edit Profile</Item>
                <Item key="logout">Logout</Item>
              </Menu>
            </MenuTrigger>
          </View>
        </Flex>
      </View>
      <Flex direction={{ base: 'column', M: 'row' }} flex>
        <Navbar elements={navItems} />
        <Flex direction="column" flexGrow={1} minWidth={0} minHeight={0}>
          {children}
        </Flex>
      </Flex>
    </Flex>
  )
}
