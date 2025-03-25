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

export function Users() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/users'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/users', { signal })
      return data?.users || []
    },
    retry: false,
    throwOnError: false,
  })

  if (query.isError) {
    return (
      <FullLayout>
        <View padding="size-500">
          <Text>Failed to load users. Please try again later.</Text>
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
        <ActionButton alignSelf="start" onPress={() => navigate('/users/add')}>
          <Add />
          <Text>Add User</Text>
        </ActionButton>
        <TableView aria-label="Users list table" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="name">Name</Column>
            <Column key="actions" width={50} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item: components['schemas']['user']) => (
              <Row key={item.uuid}>
                <Cell>{item.name}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/users/' + item.uuid)}>
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
