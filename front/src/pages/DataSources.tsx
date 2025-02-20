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
import Add from '@spectrum-icons/workflow/Add';
import Email from '@spectrum-icons/workflow/Email';
import Edit from '@spectrum-icons/workflow/Edit';
import { useNavigate } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"

import client from "@/api/client"

import { FullLayout } from '@/layouts/FullLayout'

export function DataSources() {
  const navigate = useNavigate()
  const query = useQuery({
    queryKey: ["/datasource/email"],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET("/datasource/email", { signal });
      return data || []
    },
    retry: false,
    throwOnError: true,
  })

  const rows = query.data?.map((item) => {
    return {
      key: item.uuid,
      name: item.name,
      type: item.type,
      accountId: item.email,
      state: item.is_enabled,
      authFailed: !item.oauth2_client_id || (item.oauth2_client_id && !item.oauth2_token_uuid) ? true : false,
    }
  })

  const typeRender = (type: string) => {
    if (type === "email") {
      return <Badge variant="neutral"><Email /> <Text>Email</Text></Badge>
    }
    return type
  }

  const stateRedner = ({ state, authFailed }: { state: boolean, authFailed: boolean }) => {
    if (state) {
      if (authFailed) {
        return <Badge variant="negative">Not Authenticated</Badge>
      }
      return <Badge variant="positive">Authenticated</Badge>
    }
    return <Badge variant="negative">Disabled</Badge>
  }

  if (query.isPending) {
    return <></>
  }

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate("/datasources/add")}><Add /><Text>Add Data Source</Text></ActionButton>
        <TableView
          aria-label="Example table with dynamic content"
          overflowMode="wrap"
          maxWidth={1000}
        >
          <TableHeader>
            <Column key="name">Name</Column>
            <Column key="accountId">Account ID</Column>
            <Column key="type" maxWidth={130}>Type</Column>
            <Column key="state" maxWidth={160}>State</Column>
            <Column key="actions" width={50} hideHeader>Actions</Column>
          </TableHeader>
          <TableBody items={rows}>
            {(item) => (
              <Row>
                <Cell>{item.name}</Cell>
                <Cell>{item.accountId}</Cell>
                <Cell>{typeRender(item.type)}</Cell>
                <Cell>{stateRedner(item)}</Cell>
                <Cell>
                  <ActionButton onPress={() => navigate("/datasources/" + item.key)}>
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
