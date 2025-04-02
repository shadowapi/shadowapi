import { useLocation, useNavigate } from 'react-router-dom'
import { ActionButton, Flex, Text, View } from '@adobe/react-spectrum'

export interface NavbarProps {
  Label: string
  AriaLabel: string
  Icon: React.JSX.Element
  URL: string
  Childrens?: NavbarProps[]
}

export function Navbar(props: { elements: NavbarProps[] }) {
  const location = useLocation()
  const navigate = useNavigate()

  const isDisabled = (url: string) => {
    if (url === '/') {
      return location.pathname === '/'
    }
    return location.pathname.startsWith(url)
  }

  const renderNavItem = (item: NavbarProps, level = 0) => (
    <View key={item.Label} marginStart={level * 16}>
      <ActionButton
        isQuiet
        isDisabled={isDisabled(item.URL)}
        onPress={() => navigate(item.URL)}
        aria-label={item.AriaLabel}
      >
        {item.Icon}
        <Text>{item.Label}</Text>
      </ActionButton>

      {item.Childrens?.length > 0 && 
        <Flex direction="column" gap="size-100" alignItems="start">
          {item.Childrens.map((child) => renderNavItem(child, level + 1))}
        </Flex>
      }
    </View>
  )

  return (
    <View width={{ base: '100%', M: 'size-2400' }} backgroundColor="gray-200" padding="size-200">
      <Flex direction="column" gap="size-100" alignItems="start">
        {props.elements.map((item) => renderNavItem(item))}
      </Flex>
    </View>
  )
}
