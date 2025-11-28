import { useEffect } from 'react'
import { Button, Flex, Header, Text, View } from '@adobe/react-spectrum'

// New OIDC-based login page: redirect to backend to start provider flow.
// The legacy email/password form is intentionally removed from UI but kept in the repo.
export function LoginPage() {
  const startLogin = () => {
    window.location.href = '/api/v1/auth/login'
  }

  useEffect(() => {
    // Optionally auto-start login for streamlined UX
    startLogin()
  }, [])

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" height="100vh">
      <View padding="size-300" backgroundColor="gray-200" borderRadius="medium" width="size-4600">
        <Flex direction="column" gap="size-300" alignItems="center">
          <Header>Sign in to ShadowAPI</Header>
          <Text>Continue with your identity provider.</Text>
          <Button variant="cta" onPress={startLogin}>Continue</Button>
        </Flex>
      </View>
    </Flex>
  )
}
