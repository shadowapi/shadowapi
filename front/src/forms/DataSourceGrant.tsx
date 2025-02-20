import {
  Button,
  Flex,
  StatusLight,
} from '@adobe/react-spectrum';

import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"

import type { components } from '@/api/v1'
import client from '@/api/client'

interface DataSourceGrantProps {
  datasourceUUID: string
  tokenUUID?: string
}

export const DataSourceGrant = (props: DataSourceGrantProps) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: ['/datasource/email/{uuid}', { uuid: props.datasourceUUID }],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET(`/datasource/email/{uuid}`, {
        params: { path: { uuid: props.datasourceUUID } },
        signal,
      })
      return data
    },
  })
  const mutation = useMutation({
    mutationKey: [
      '/oauth2/client/{datasource_uuid}/token/{uuid}',
      { datasource_uuid: props.datasourceUUID, uuid: props.tokenUUID },
    ],
    mutationFn: async () => {
      const resp = await client.DELETE('/oauth2/client/{datasource_uuid}/token/{uuid}', {
        params: { path: { datasource_uuid: props.datasourceUUID, uuid: props.tokenUUID! } }
      })
      if (resp.error) {
        throw new Error(resp.error.detail)
      }
    },
    onSuccess: () => {
      // Update the datasource to remove the token
      const datasource = queryClient.getQueryData<components["schemas"]["datasource"]>(
        ['/datasource/email/{uuid}', { uuid: props.datasourceUUID }],
      )
      if (datasource) {
        datasource.oauth2_token_uuid = ""
      }
      queryClient.setQueryData(['/datasource/email/{uuid}', { uuid: props.datasourceUUID }], datasource)
    },
  })

  return (
    <Flex direction="column" gap="size-200" alignItems="center">
      <StatusLight variant={query.data?.oauth2_token_uuid ? "positive" : "negative"}>
        {query.data?.oauth2_token_uuid ? "Connection Is Authenticated" : "Connection Is Not Authenticated"}
      </StatusLight>
      <Flex direction="row" gap="size-100">
        <Button
          variant="accent"
          isDisabled={query.isFetching || mutation.isPending}
          onPress={() => navigate(`/datasources/${query.data?.uuid}/auth`)}
        >
          {query.data?.oauth2_token_uuid ? "Re-Auth" : "Authenticate"}
        </Button>
        {query.data?.oauth2_token_uuid && (
          <Button
            variant="negative"
            isDisabled={query.isFetching || mutation.isPending}
            onPress={() => mutation.mutate()}
          >
            Revoke
          </Button>
        )}
      </Flex>
    </Flex>
  )
}
