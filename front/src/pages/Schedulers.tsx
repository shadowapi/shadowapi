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

export function Schedulers() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/scheduler'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/scheduler', { signal })
      return data || []
    },
    retry: false,
    throwOnError: false,
  })

  if (query.isError) {
    return (
      <FullLayout>
        <View padding="size-500">
          <Text>Failed to load schedulers. Please try again later.</Text>
        </View>
      </FullLayout>
    )
  }

  if (query.isLoading) {
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
        <ActionButton alignSelf="start" onPress={() => navigate('/schedulers/add')}>
          <Add />
          <Text>Add Scheduler</Text>
        </ActionButton>
        <TableView aria-label="Schedulers list table" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="scheduleType">Schedule Type</Column>
            <Column key="nextRun">Next Run</Column>
            <Column key="actions" width={50} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item: components['schemas']['scheduler']) => (
              <Row key={item.id}>
                <Cell>{item.schedule_type}</Cell>
                <Cell>{item.next_run ? new Date(item.next_run).toLocaleString() : 'N/A'}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/schedulers/' + item.id)}>
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
