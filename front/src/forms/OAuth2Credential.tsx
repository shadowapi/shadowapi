import {
  Button,
  Flex,
  Form,
  Header,
  Item,
  Picker,
  TextField,
} from '@adobe/react-spectrum';
import { useEffect, ReactElement } from 'react'
import { useForm, Controller } from "react-hook-form"
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"

import type { components } from '@/api/v1'
import client from '@/api/client'


export function OAuth2Credential({ clientID: clientID }: { clientID: string }): ReactElement {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: ['/oauth2/client/{id}', { id: clientID }],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/oauth2/client/{id}', {
        params: { path: { id: clientID } },
        signal,
      })
      return data
    },
    enabled: clientID !== "add",
  })
  const updateMutation = useMutation({
    mutationFn: async (data: components["schemas"]["oauth2_client"]) => {
      let resp
      if (clientID === "add") {
        resp = await client.POST('/oauth2/client', {
          body: {
            provider: data.provider,
            id: data.id,
            name: data.name,
            secret: data.secret,
          }
        })
      } else {
        resp = await client.PUT('/oauth2/client/{id}', {
          params: { path: { id: clientID } },
          body: {
            provider: data.provider,
            name: data.name,
            secret: data.secret,
          },
        })
      }
      if (resp.error) {
        methods.setError('name', { message: resp.error.detail })
        return
      }
    },
    onSuccess: (data, variables) => {
      if (clientID === "add") {
        queryClient.invalidateQueries({ queryKey: '/oauth2/client' })
      } else {
        queryClient.setQueryData(['/oauth2/client/{id}', { id: variables.id }], data)
      }
    },
  })
  const deleteMutation = useMutation({
    mutationFn: async (clientID: string) => {
      const resp = await client.DELETE('/oauth2/client/{id}', {
        params: { path: { id: clientID } },
      })
      if (resp.error) {
        methods.setError('name', { message: resp.error.detail })
        return
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/oauth2/client' })
    }
  })

  const methods = useForm<components["schemas"]["oauth2_client"]>({})

  // load form data from the query
  useEffect(() => {
    if (query.data) {
      methods.reset(query.data)
    }
  }, [query.data, methods])

  const onSubmit = async (data: components["schemas"]["oauth2_client"]) => {
    updateMutation.mutate(data, {
      onSuccess: (_, variables) => {
        queryClient.invalidateQueries({ queryKey: ['/oauth2/credentials'] })
        const datasourceUUID = searchParams.get('datasource_uuid')
        if (datasourceUUID) {
          navigate(`/datasource/${datasourceUUID}/auth?client_id=${variables.id}`)
          return
        }
        navigate('/oauth2/credentials')
      }
    })
  }

  const onDelete = async () => {
    deleteMutation.mutate(clientID, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['/oauth2/credentials'] })
        navigate('/oauth2/credentials')
      }
    })
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <Form onSubmit={methods.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600">
          <Header marginBottom="size-160">OAuth2 Credential</Header>
          <Controller
            name="name"
            control={methods.control}
            rules={{ required: 'Name is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Name" type="text" width="100%" isRequired
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
                {...field}
              />
            )}
          />

          <Controller
            name="provider"
            control={methods.control}
            rules={{ required: 'Provider is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Provider"
                isRequired
                selectedKey={field.value}
                onSelectionChange={(key) => methods.setValue('provider', key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                <Item key="GMAIL">Gmail</Item>
              </Picker>
            )}
          />

          <Controller
            name="id"
            control={methods.control}
            rules={{ required: 'Client ID is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Client ID" type="text" width="100%" isRequired
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
                {...field}
              />
            )}
          />

          <Controller
            name="secret"
            control={methods.control}
            rules={{ required: 'Client Secret is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Client Secret" type="password" width="100%" isRequired
                validationState={fieldState.invalid ? 'invalid' : undefined}
                defaultValue={field.value}
                errorMessage={fieldState.error?.message}
                {...field}
              />
            )}
          />
          <Flex direction="row" gap="size-100" marginTop="size-300">
            <Button
              type="submit"
              variant="cta"
              isPending={updateMutation.isPending}
            >
              {clientID === "add" ? "Create" : "Update"}
            </Button>
            {clientID !== "add" && (
              <Button
                type="button"
                variant="negative"
                isPending={deleteMutation.isPending}
                onPress={onDelete}
              >
                Delete
              </Button>
            )}
          </Flex>
        </Flex>
      </Form>
    </Flex>
  )
}
