import { useNavigate } from 'react-router-dom'
import { Button, Content, Flex, Heading, IllustratedMessage, Text } from '@adobe/react-spectrum'
import NotFound from '@spectrum-icons/illustrations/NotFound'
import Add from '@spectrum-icons/workflow/Add'

import { FullLayout } from '@/layouts/FullLayout'

// Dashboard of the application
export function Dashboard() {
  const navigate = useNavigate()
  return (
    <FullLayout>
      <Flex direction="column" justifyContent="center" alignContent="center" flexBasis="100%" gap="size-200">
        <IllustratedMessage flex={0}>
          <NotFound />
          <Heading>No data sources are set up</Heading>
          <Content>
            <Text>Add new data source to start processing data</Text>
          </Content>
        </IllustratedMessage>
        <Button maxWidth="size-2000" alignSelf="center" onPress={() => navigate('/datasources/add')} variant="cta">
          <Add /> Add Data Source
        </Button>
      </Flex>
    </FullLayout>
  )
}
