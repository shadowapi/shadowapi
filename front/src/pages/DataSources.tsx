import { useNavigate } from 'react-router-dom'
import {
  ActionButton,
  Badge,
  Cell,
  Column,
  Flex,
  Row,
  TableBody,
  TableHeader,
  TableView,
  Text,
  View,
} from '@adobe/react-spectrum'
import Add from '@spectrum-icons/workflow/Add'
import Edit from '@spectrum-icons/workflow/Edit'
import Email from '@spectrum-icons/workflow/Email'
import { useQuery } from '@tanstack/react-query'

import client from '@/api/client'
import { FullLayout } from '@/layouts/FullLayout'

export function DataSources() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/datasource'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/datasource', { signal })
      return data || []
    },
    retry: false,
    throwOnError: false,
  })

  const rows = query.data?.map((item) => {
    return {
      key: item.uuid,
      name: item.name,
      type: item.type,
      provider: item.provider,
      state: item.is_enabled ? 'Enabled' : 'Disabled',
    }
  })

  const typeRender = (type: string) => {
    if (type === 'email') {
      return (
        <Badge variant="neutral">
          <Email /> <Text>Email</Text>
        </Badge>
      )
    }
    return type
  }

  const stateRedner = ({ state, authFailed }: { state: boolean; authFailed: boolean }) => {
    if (state) {
      if (authFailed) {
        return <Badge variant="negative">Not Authenticated</Badge>
      }
      return <Badge variant="positive">Authenticated</Badge>
    }
    return <Badge variant="negative">Disabled</Badge>
  }

  if (query.isError) {
    return (
      <FullLayout>
        <View padding="size-500">
          <Text>Failed to load data sources. Please try again later.</Text>
        </View>
      </FullLayout>
    )
  }

  if (query.isPending) {
    return <></>
  }

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate('/datasources/add')}>
          <Add />
          <Text>Add Data Source</Text>
        </ActionButton>
        <TableView aria-label="Example table with dynamic content" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="name">Name</Column>
            <Column key="provider">Provider</Column>
            <Column key="type" maxWidth={130}>
              Type
            </Column>
            <Column key="state" maxWidth={160}>
              State
            </Column>
            <Column key="actions" width={50} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={rows}>
            {(item) => (
              <Row>
                <Cell>{item.name}</Cell>
                <Cell>{item.provider}</Cell>
                <Cell>{typeRender(item.type)}</Cell>
                <Cell>
                  <Badge variant={item.state === 'Enabled' ? 'positive' : 'negative'}>{item.state}</Badge>
                </Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/datasources/' + item.key)}>
                    <Edit />
                  </ActionButton>
                </Cell>
              </Row>
            )}
          </TableBody>
        </TableView>
      </Flex>
    </FullLayout>
  )
}
