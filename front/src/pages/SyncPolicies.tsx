import { useNavigate } from 'react-router-dom'
import {
  ActionButton,
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
import { useQuery } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'
import { FullLayout } from '@/layouts/FullLayout'

export function SyncPolicies() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/syncpolicy'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/syncpolicy', { signal })
      return data?.policies || []
    },
    retry: false,
    throwOnError: false,
  })

  if (query.isError) {
    return (
      <FullLayout>
        <View padding="size-500">
          <Text>Failed to load policies. Please try again later.</Text>
        </View>
      </FullLayout>
    )
  }

  if (query.isPending) {
    return (
      <FullLayout>
        <Flex direction="column">
          <Text>Loading...</Text>
        </Flex>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate('/syncpolicy/add')}>
          <Add />
          <Text>Add Policy</Text>
        </ActionButton>
        <TableView aria-label="Policies list table" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="service">Service</Column>
            <Column key="sync_all">Sync All</Column>
            <Column key="created_at">Created At</Column>
            <Column key="actions" width={50} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item: components['schemas']['sync_policy']) => (
              <Row key={item.uuid}>
                <Cell>{item.service}</Cell>
                <Cell>{item.sync_all ? 'Yes' : 'No'}</Cell>
                <Cell>{item.created_at}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/syncpolicy/' + item.uuid)}>
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
