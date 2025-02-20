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
import Add from '@spectrum-icons/workflow/Add';
import { useNavigate } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import Edit from '@spectrum-icons/workflow/Edit';

import { FullLayout } from '@/layouts/FullLayout'
import type { components } from '@/api/v1'
import client from '@/api/client'

export function Pipelines() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ['/pipelines'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline', { signal })
      return data?.pipelines || []
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
        <ActionButton alignSelf="start" onPress={() => navigate("/pipelines/add")}>
          <Add /><Text>Add Data Pipeline</Text>
        </ActionButton>
        <TableView
          aria-label="Pipelines list table"
          overflowMode="wrap"
          maxWidth={1000}
        >
          <TableHeader>
            <Column key="name">Name</Column>
            <Column key="actions" width={50} hideHeader>Actions</Column>
          </TableHeader>
          <TableBody items={query.data}>
            {(item: components["schemas"]["pipeline"]) => (
              <Row key={item.uuid}>
                <Cell>{item.name}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate("/pipelines/" + item.uuid)}>
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
