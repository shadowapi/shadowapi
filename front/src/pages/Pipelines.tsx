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

export function Pipelines() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/pipelines'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline', { signal })
      return data?.pipelines || []
    },
    retry: false,
    throwOnError: false,
  })

  if (query.isError) {
    return (
      <FullLayout>
        <View padding="size-500">
          <Text>Failed to load data sources. Please try again later.</Text>
        </View>
      </FullLayout>
    )
  }

  if (query.isPending)
    return (
      <FullLayout>
        <Flex direction="column">
          <Text>Loading...</Text>
        </Flex>
      </FullLayout>
    )

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate('/pipelines/add')}>
          <Add />
          <Text>Add Data Pipeline</Text>
        </ActionButton>
        <TableView aria-label="Pipelines list table" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="name">Name</Column>
            <Column key="type">Type</Column>
            <Column key="datasource_uuid">Data Source</Column>
            <Column key="storage_uuid">Storage</Column>
            <Column key="schedulers">Schedulers</Column>
            <Column key="actions" width={50} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item: components['schemas']['pipeline']) => (
              <Row key={item.uuid}>
                <Cell>{item.name}</Cell>
                <Cell>{item.type}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/datasources/' + item.datasource_uuid)}>
                    {item.datasource_uuid}
                  </ActionButton>
                </Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/storages/' + item.storage_uuid + '/storageKind/hostfiles')}>
                    TODO @reactima {item.storage_uuid}
                  </ActionButton>
                </Cell>
                <Cell>-</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/pipelines/' + item.uuid)}>
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
