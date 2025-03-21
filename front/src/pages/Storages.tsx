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
} from '@adobe/react-spectrum'
import Add from '@spectrum-icons/workflow/Add'
import Data from '@spectrum-icons/workflow/Data'
import Edit from '@spectrum-icons/workflow/Edit'
import { useQuery } from '@tanstack/react-query'

import client from '@/api/client'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

export function Storages() {
  const navigate = useNavigate()

  const pageTitle = 'Storages'
  useTitle(pageTitle)

  const query = useQuery({
    queryKey: ['/storages'],
    queryFn: async ({ signal }) => {
      const { data, error } = await client.GET('/storage', { signal })
      if (error) {
        console.error(error)
        return []
      }
      return data
    },
    throwOnError: false,
  })

  const rows =
    query.data?.map((item) => ({
      id: item.uuid,
      type: item.type,
      name: item.name,
      state: item.is_enabled ? 'Enabled' : 'Disabled',
    })) ?? [] // fallback to empty array

  const typeRender = (type: string) => {
    if (type === 'postgres') {
      return (
        <Badge variant="neutral">
          <Data /> <Text>PostgreSQL</Text>
        </Badge>
      )
    }
    if (type === 's3') {
      return (
        <Badge variant="neutral">
          <Data /> <Text>S3</Text>
        </Badge>
      )
    }
    if (type === 'hostfiles') {
      return (
        <Badge variant="neutral">
          <Data /> <Text>Hostfiles</Text>
        </Badge>
      )
    }
    return <Badge variant="neutral">... missing type</Badge>
  }

  if (query.isPending) {
    return <></>
  }

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate('/storages/add')}>
          <Add />
          <Text>Add Data Source</Text>
        </ActionButton>
        <TableView aria-label="Example table with dynamic content" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="type" maxWidth={160}>
              Type
            </Column>
            <Column key="name" maxWidth={160}>
              Name
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
              <Row key={item.id}>
                <Cell>{typeRender(item.type)}</Cell>
                <Cell>{item.name}</Cell>
                <Cell>
                  <Badge variant={item.state === 'Enabled' ? 'positive' : 'negative'}>{item.state}</Badge>
                </Cell>
                <Cell>
                  <ActionButton onPress={() => navigate('/storages/' + item.id + '/storageKind/' + item.type)}>
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
