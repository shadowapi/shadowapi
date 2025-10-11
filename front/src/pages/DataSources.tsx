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
import Login from '@spectrum-icons/workflow/Login'
import Remove from '@spectrum-icons/workflow/Remove'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'
import { FullLayout } from '@/layouts/FullLayout'

type Datasource = components['schemas']['datasource'] &
  Partial<components['schemas']['datasource_email']> &
  Partial<components['schemas']['datasource_email_oauth']>

type Row = {
  key: string
  name: string
  type: string
  state: 'Enabled' | 'Disabled'
}

export function DataSources() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const listQuery = useQuery({
    queryKey: ['/datasource'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/datasource', { signal })
      return (data || []) as Datasource[]
    },
    retry: false,
    throwOnError: false,
  })

  // Revoke **all** OAuth2 tokens from datasourceUUID

  const mutationRevokeTokens = useMutation({
    mutationKey: ['revokeTokens'],
    mutationFn: async (datasourceUUID: string) => {
      // TODO @reactima simplify
      /* 1. fetch every token bound to the datasource */
      const listResp = await client.GET('/oauth2/client/{datasource_uuid}/token', {
        params: { path: { datasource_uuid: datasourceUUID } },
      })
      if (listResp.error) throw new Error(listResp.error.detail)

      const tokens = listResp.data ?? []
      /* 2. delete them one by one */
      for (const tok of tokens) {
        await client.DELETE('/oauth2/client/{datasource_uuid}/token/{uuid}', {
          params: { path: { datasource_uuid: datasourceUUID, uuid: tok.uuid! } },
        })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['/datasource'] })
    },
  })

  // Start Gmail OAuth2 login
  const handleOauthLogin = async (ds: Row) => {
    if (ds.type !== 'email_oauth') return

    /* 2. ask backend for auth_code_url */
    const loginResp = await client.POST('/oauth2/login', {
      body: {
        query: { datasource_uuid: [ds.key] },
      },
    })
    if (loginResp.error) {
      alert(loginResp.error.detail || 'Failed to initiate login')
      return
    }
    if (loginResp.data?.auth_code_url) window.location.href = loginResp.data.auth_code_url
  }

  const typeBadge = (type: string) => {
    if (type === 'email') {
      return (
        <Badge variant="neutral">
          <Email /> <Text UNSAFE_style={{ textWrap: 'nowrap' }}>Email IMAP</Text>
        </Badge>
      )
    }
    if (type === 'email_oauth') {
      return (
        <Badge variant="neutral">
          <Email /> <Text UNSAFE_style={{ textWrap: 'nowrap' }}>Email OAuth</Text>
        </Badge>
      )
    }
    return (
      <Badge variant="neutral">
        <Text UNSAFE_style={{ textWrap: 'nowrap' }}>{type}</Text>
      </Badge>
    )
  }

  const rows: Row[] | undefined = listQuery.data?.map((ds) => ({
    key: ds.uuid!,
    name: ds.name,
    type: ds.type,
    state: ds.is_enabled ? 'Enabled' : 'Disabled',
  }))

  if (listQuery.isError) {
    return (
      <FullLayout>
        <View padding="size-500">
          <Text>Failed to load data sources. Please try again later.</Text>
        </View>
      </FullLayout>
    )
  }

  if (listQuery.isPending) return <></>

  console.log({ rows, listQuery })

  return (
    <FullLayout>
      <Flex direction="column" margin="size-500" gap="size-100" minWidth={0} minHeight={0}>
        <ActionButton alignSelf="start" onPress={() => navigate('/datasources/add')}>
          <Add />
          <Text>Add Data Source</Text>
        </ActionButton>

        <TableView aria-label="Data sources table" overflowMode="wrap" maxWidth={1000}>
          <TableHeader>
            <Column key="name">Name</Column>
            <Column key="type" maxWidth={160}>
              Type
            </Column>
            <Column key="state" maxWidth={160}>
              State
            </Column>
            <Column key="login" width={120}>
              Re/Auth
            </Column>
            <Column key="revoke" width={120}>
              Revoke
            </Column>
            <Column key="actions" width={120} hideHeader>
              Actions
            </Column>
          </TableHeader>
          <TableBody items={rows}>
            {(item) => (
              <Row>
                <Cell>{item.name}</Cell>
                <Cell>{typeBadge(item.type)}</Cell>
                <Cell>
                  <Badge variant={item.state === 'Enabled' ? 'positive' : 'negative'}>{item.state}</Badge>
                </Cell>
                <Cell>
                  {item.type == 'email_oauth' ? (
                    <ActionButton onPress={() => handleOauthLogin(item)}>
                      <Login />
                    </ActionButton>
                  ) : (
                    <span>-</span>
                  )}
                </Cell>
                <Cell>
                  {item.type == 'email_oauth' ? (
                    <ActionButton onPress={() => mutationRevokeTokens.mutate(item.key)}>
                      <Remove />
                    </ActionButton>
                  ) : (
                    <span>-</span>
                  )}
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
