import { useLocation, useNavigate } from 'react-router-dom'
import { ActionButton, Flex, Text, View } from '@adobe/react-spectrum'

interface NavbarProps {
  Label: string
  AriaLabel: string
  Icon: React.JSX.Element
  URL: string
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

  return (
    <View width={{ base: '100%', M: 'size-2400' }} backgroundColor="gray-200" padding="size-200">
      <Flex direction="column" gap="size-100" alignItems="start">
        {props.elements.map((i: NavbarProps) => (
          <ActionButton
            key={i.Label}
            isQuiet
            isDisabled={isDisabled(i.URL)}
            onPress={() => navigate(i.URL)}
            aria-label={i.AriaLabel}
          >
            {i.Icon && i.Icon}
            <Text>{i.Label}</Text>
          </ActionButton>
        ))}
      </Flex>
    </View>
  )
}
