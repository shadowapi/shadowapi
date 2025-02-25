import { ReactElement, useEffect, useState } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Flex, Form, Header, Item, Picker } from '@adobe/react-spectrum'

import client from '@/api/client'
import type { components } from '@/api/v1'

export interface FormData {
  oauth2_clients: components['schemas']['Oauth2Client'][]
  id: string
  name: string
  provider: string
  secret: string
}

export function DataSourceAuthForm({ datasourceUUID: datasourceUUID }: { datasourceUUID: string }): ReactElement {
  const navigate = useNavigate()
  const [isLoaded, setIsLoaded] = useState(false)
  const [searchParams] = useSearchParams()
  const methods = useForm<FormData>({
    defaultValues: {
      oauth2_clients: [],
    },
  })

  const clientID = useWatch({ control: methods.control, name: 'id', defaultValue: '' })
  const clientsCount = useWatch({ control: methods.control, name: 'oauth2_clients', defaultValue: [] }).length
  useEffect(() => {
    const getClients = async () => {
      const resp = await client.GET('/oauth2/client', {
        params: { query: { limit: 1000 } },
      })
      methods.setValue('oauth2_clients', resp.data?.clients || [])
    }
    getClients()
  }, [clientID, methods])

  useEffect(() => {
    const getClients = async () => {
      const resp = await client.GET('/datasource/email/{uuid}', {
        params: { path: { uuid: datasourceUUID } },
      })
      if (resp.data?.oauth2_client_id) {
        methods.setValue('id', resp.data?.oauth2_client_id)
      }
      setIsLoaded(true)
    }

    getClients()
  }, [datasourceUUID, methods, setIsLoaded])

  useEffect(() => {
    const clientID = searchParams.get('client_id')
    if (clientID) {
      methods.setValue('id', clientID)
    }
  }, [searchParams, methods])

  const onSubmit = async () => {
    const resp = await client.PUT('/datasource/{uuid}/oauth2/client', {
      params: { path: { uuid: datasourceUUID } },
      body: { client_id: clientID },
    })
    if (resp.error) {
      methods.setError('id', { type: 'manual', message: 'Failed to authenticate data source' })
      return
    }
    const loginResp = await client.POST('/oauth2/login', {
      body: {
        client_id: clientID,
        query: { datasource_uuid: [datasourceUUID] },
      },
    })
    if (loginResp.error) {
      methods.setError('id', { type: 'manual', message: 'Failed to perform login' })
      return
    }
    if (loginResp.data?.auth_code_url) window.location.href = loginResp.data?.auth_code_url
    return
  }

  if (!isLoaded) {
    return <></>
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <Form onSubmit={methods.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600">
          <Header marginBottom="size-160">Authenticate Datasource</Header>
          <Controller
            name="id"
            control={methods.control}
            rules={{ required: 'Client ID is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Type"
                isRequired
                selectedKey={field.value}
                onSelectionChange={field.onChange}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                {methods.watch('oauth2_clients').map((client) => 
                  <Item key={client.id}>{client.name}</Item>
                )}
              </Picker>
            )}
          />
          <Flex direction="row" gap="size-100" marginTop="size-300">
            <Button
              type="button"
              variant="primary"
              onPress={() => navigate(`/oauth2/credentials/add?datasource_uuid=${datasourceUUID}`)}
            >
              Add New
            </Button>
            {clientID !== 'add' && (
              <Button type="submit" variant="cta" isDisabled={clientsCount === 0 || clientID == ''}>
                Authenticate
              </Button>
            )}
          </Flex>
        </Flex>
      </Form>
    </Flex>
  )
}
