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
        <TableView aria-label="Schedulers list table" overflowMode="wrap" maxWidth={1400}>
          <TableHeader>
            <Column key="pipeline_uuid">Pipeline UUID</Column>
            <Column key="schedule_type">Schedule Type</Column>
            <Column key="cron_expression">Cron Expression</Column>
            <Column key="run_at">Run At</Column>
            <Column key="timezone">Timezone</Column>
            <Column key="next_run">Next Run</Column>
            <Column key="last_run">Last Run</Column>
            <Column key="is_enabled">Enabled</Column>
            <Column key="is_paused">Paused</Column>
            <Column key="actions" width={50} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item: components['schemas']['scheduler']) => (
              <Row key={item.id}>
                <Cell>
                  <ActionButton onPress={() => navigate('/pipelines/' + item.pipeline_uuid)}>
                    ...{item.pipeline_uuid.slice(-8)}
                  </ActionButton>
                </Cell>
                <Cell>{item.schedule_type}</Cell>
                <Cell>{item.cron_expression || 'N/A'}</Cell>
                <Cell>{item.run_at ? new Date(item.run_at).toLocaleString() : 'N/A'}</Cell>
                <Cell>{item.timezone}</Cell>
                <Cell>{item.next_run ? new Date(item.next_run).toLocaleString() : 'N/A'}</Cell>
                <Cell>{item.last_run ? new Date(item.last_run).toLocaleString() : 'N/A'}</Cell>
                <Cell>
                  <Badge variant={item.is_enabled ? 'positive' : 'negative'}>
                    {item.is_enabled ? 'Enable' : 'Disable'}
                  </Badge>
                </Cell>
                <Cell>
                  <Badge variant={item.is_paused ? 'positive' : 'negative'}>
                    {item.is_enabled ? 'Running' : 'Paused'}
                  </Badge>
                </Cell>
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
