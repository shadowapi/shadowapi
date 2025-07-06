import { Button, Flex, Header, Link, View } from '@adobe/react-spectrum'

export function SignupPage() {
  const signupUrl = `${import.meta.env.VITE_ZITADEL_INSTANCE_URL}/oauth/v2/authorize?client_id=${import.meta.env.VITE_ZITADEL_CLIENT_ID}&response_type=code&scope=openid&redirect_uri=${encodeURIComponent(import.meta.env.VITE_ZITADEL_REDIRECT_URI)}`

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View padding="size-200" backgroundColor="gray-200" borderRadius="medium" width="size-3600">
        <Flex direction="column" gap="size-100">
          <Header>Sign Up</Header>
          <Button
            variant="cta"
            alignSelf="end"
            marginTop="size-150"
            width="size-1250"
            onPress={() => {
              window.location.href = signupUrl
            }}
          >
            Sign up with ZITADEL
          </Button>
          <Link href="/login" alignSelf="end" marginTop="size-100">
            Back to login
          </Link>
        </Flex>
      </View>
    </Flex>
  )
}
