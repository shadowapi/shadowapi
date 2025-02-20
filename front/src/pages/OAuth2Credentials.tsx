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
} from '@adobe/react-spectrum'
import Edit from '@spectrum-icons/workflow/Edit';
import Add from '@spectrum-icons/workflow/Add';
import { useNavigate } from 'react-router-dom'
import { useQuery } from "@tanstack/react-query"

import client from '@/api/client'

import { FullLayout } from '@/layouts/FullLayout'

export function OAuth2Credentials() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/oauth2/clients'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/oauth2/client', { signal })
      return data?.clients || []
    },
    retry: false,
    throwOnError: true,
  })

  if (query.isPending) return (
    <FullLayout>
      <Flex direction="column">
        <Text>Loading...</Text>
      </Flex>
    </FullLayout>
  )

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate("/oauth2/credentials/add")}><Add /><Text>Add OAuth2 Credentials</Text></ActionButton>
        <TableView
          aria-label="OAuth2 Clients data table"
          overflowMode="wrap"
          maxWidth={580}
        >
          <TableHeader>
            <Column key="name" maxWidth={400}>Name</Column>
            <Column key="type" maxWidth={150}>Provider</Column>
            <Column key="actions" maxWidth={30} hideHeader>Actions</Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item) => (
              <Row>
                <Cell>{item.name}</Cell>
                <Cell>{item.provider}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate("/oauth2/credentials/" + item.id)}>
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
