import {
  Flex,
  Text,
} from '@adobe/react-spectrum'

import { FullLayout } from '@/layouts/FullLayout'

export function Workers() {
  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <Text>Coming soon...</Text>
      </Flex>
    </FullLayout>
  )
}
