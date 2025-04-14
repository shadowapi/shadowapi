import { ReactElement, useEffect } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Item, Picker, Switch, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components, paths } from '@/api/v1'

type DataSourceBase = {
  name: string
  type: string
  is_enabled: boolean
  user_uuid?: string
}

export type DataSourceKind = 'email' | 'email_oauth' | 'telegram' | 'whatsapp' | 'linkedin'

export type DataSourceFormData =
  | (DataSourceBase & { type: 'email' } & components['schemas']['datasource_email'])
  | (DataSourceBase & { type: 'email_oauth' } & components['schemas']['datasource_email_oauth'])
  | (DataSourceBase & { type: 'telegram' } & components['schemas']['datasource_telegram'])
  | (DataSourceBase & { type: 'whatsapp' } & components['schemas']['datasource_whatsapp'])
  | (DataSourceBase & { type: 'linkedin' } & components['schemas']['datasource_linkedin'])

const createEndpoints: Record<DataSourceKind, keyof paths> = {
  email: '/datasource/email',
  email_oauth: '/datasource/email_oauth',
  telegram: '/datasource/telegram',
  whatsapp: '/datasource/whatsapp',
  linkedin: '/datasource/linkedin',
}

const updateEndpoints: Record<DataSourceKind, keyof paths> = {
  email: '/datasource/email/{uuid}',
  email_oauth: '/datasource/email_oauth/{uuid}',
  telegram: '/datasource/telegram/{uuid}',
  whatsapp: '/datasource/whatsapp/{uuid}',
  linkedin: '/datasource/linkedin/{uuid}',
}

const deleteEndpoints: Record<DataSourceKind, keyof paths> = {
  email: '/datasource/email/{uuid}',
  email_oauth: '/datasource/email_oauth/{uuid}',
  telegram: '/datasource/telegram/{uuid}',
  whatsapp: '/datasource/whatsapp/{uuid}',
  linkedin: '/datasource/linkedin/{uuid}',
}

export function DataSourceForm({ datasourceUUID }: { datasourceUUID: string }): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<DataSourceFormData>({})

  // Always watch the "type" field so the hook order is fixed.
  const watchedType = useWatch({ control: form.control, name: 'type' })

  const usersQuery = useQuery({
    queryKey: ['users'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/user', { signal })
      return data as components['schemas']['user'][]
    },
  })

  // Fetch OAuth2 clients from the backend.
  const oauth2ClientsQuery = useQuery({
    queryKey: ['oauth2Clients'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/oauth2/client', { signal })
      // Expecting data to have a "clients" property.
      return data.clients as components['schemas']['oauth2_client'][]
    },
  })

  // In edit mode, fetch generic datasource record to get the type (dsKind)
  const genericQuery = useQuery({
    queryKey: datasourceUUID === 'add' ? null : ['genericDatasource', datasourceUUID],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/datasource', { signal })
      const ds = (data as components['schemas']['datasource'][]).find((ds) => ds.uuid === datasourceUUID)
      if (!ds) throw new Error('Datasource not found')
      return ds
    },
    enabled: datasourceUUID !== 'add',
  })

  // dsKind comes from the loaded record; default to "email" if missing.
  const dsKind = datasourceUUID === 'add' ? undefined : ((genericQuery.data?.type || 'email') as DataSourceKind)

  // currentType is used for conditional rendering.
  const currentType = datasourceUUID === 'add' ? watchedType : dsKind || watchedType

  const getEndpoint = (kind: DataSourceKind, isCreate: boolean): keyof paths => {
    return isCreate ? createEndpoints[kind] : updateEndpoints[kind]
  }

  // Enable the detailed query once generic data is loaded.
  const query = useQuery({
    queryKey:
      datasourceUUID === 'add' ? ['/datasource', 'add'] : [getEndpoint(dsKind!, false), { uuid: datasourceUUID }],
    queryFn: async ({ signal }) => {
      if (datasourceUUID === 'add') return {}
      const endpoint = getEndpoint(dsKind!, false)
      const { data } = await client.GET(endpoint, {
        params: { path: { uuid: datasourceUUID } },
        signal,
      })
      return data
    },
    enabled: datasourceUUID === 'add' ? true : genericQuery.data !== undefined,
  })

  // Reset the form when the record is loaded.
  useEffect(() => {
    if (query.data) {
      form.reset({
        ...query.data,
        type: dsKind ?? (query.data as Partial<DataSourceFormData>).type,
      } as DataSourceFormData)
    }
  }, [query.data, dsKind, form])

  const modifyMutation = useMutation({
    mutationFn: async (data: DataSourceFormData) => {
      if (datasourceUUID === 'add') {
        const currentKind = data.type as DataSourceKind
        const endpoint = createEndpoints[currentKind]

        console.log({ currentKind, data, endpoint })
        if (data.type === 'email') {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              email: data.email,
              provider: data.provider,
              imap_server: data.imap_server,
              smtp_server: data.smtp_server,
              smtp_tls: data.smtp_tls,
              password: data.password,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'email_oauth') {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              email: data.email,
              provider: data.provider,
              oauth2_client_uuid: data.oauth2_client_uuid,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'telegram') {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              phone_number: data.phone_number,
              provider: data.provider,
              api_id: data.api_id,
              api_hash: data.api_hash,
              password: data.password,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'whatsapp') {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              phone_number: data.phone_number,
              provider: data.provider,
              device_name: data.device_name,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else {
          // linkedin
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              username: data.username,
              password: data.password,
              provider: data.provider,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        }
      } else {
        const endpoint = getEndpoint(dsKind!, false)
        if (data.type === 'email') {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: datasourceUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              email: data.email,
              provider: data.provider,
              imap_server: data.imap_server,
              smtp_server: data.smtp_server,
              smtp_tls: data.smtp_tls,
              password: data.password,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'email_oauth') {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: datasourceUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              email: data.email,
              provider: data.provider,
              oauth2_client_uuid: data.oauth2_client_uuid,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'telegram') {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: datasourceUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              phone_number: data.phone_number,
              provider: data.provider,
              api_id: data.api_id,
              api_hash: data.api_hash,
              password: data.password,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'whatsapp') {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: datasourceUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              phone_number: data.phone_number,
              provider: data.provider,
              device_name: data.device_name,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else {
          // linkedin
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: datasourceUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              username: data.username,
              password: data.password,
              provider: data.provider,
              user_uuid: data.user_uuid,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/datasource' })
      navigate('/datasources')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const endpoint = deleteEndpoints[dsKind!]
      const resp = await client.DELETE(endpoint, {
        params: { path: { uuid } },
      })
      if (resp.error) {
        form.setError('name', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
      return resp
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/datasource' })
      navigate('/datasources')
    },
  })

  const onDelete = () => {
    deleteMutation.mutate(datasourceUUID)
  }

  const onSubmit = (data: DataSourceFormData) => {
    modifyMutation.mutate(data)
  }

  console.log('@reactima remove DataSourceForm !', {
    datasourceUUID,
    query,
    form,
    currentType,
    queryData: query.data,
  })

  if (datasourceUUID !== 'add' && (genericQuery.isPending || query.isPending)) {
    return <></>
  }

  return (
    <Flex direction="row" justifyContent="center" flexBasis="100%" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600">
          <Header marginBottom="size-160">Data Source</Header>
          <Controller
            name="name"
            control={form.control}
            rules={{ required: 'Name is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Name"
                isRequired
                type="text"
                width="100%"
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
                {...field}
              />
            )}
          />
          <Controller
            name="user_uuid"
            control={form.control}
            render={({ field, fieldState }) => (
              <Picker
                label="User"
                isRequired
                selectedKey={field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                {usersQuery.data?.map((user: components['schemas']['user']) => 
                  <Item key={user.uuid}>
                    <span
                      style={{
                        whiteSpace: 'nowrap',
                        height: '24px',
                        lineHeight: '24px',
                        marginLeft: 10,
                        marginRight: 10,
                      }}
                    >
                      {user.email} {user.first_name} {user.last_name}
                    </span>
                  </Item>
                )}
              </Picker>
            )}
          />
          <Controller
            name="type"
            control={form.control}
            rules={{ required: 'Type is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Type"
                isRequired
                isDisabled={datasourceUUID !== 'add'}
                selectedKey={field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                <Item key="email">Email IMAP</Item>
                <Item key="email_oauth">Email OAuth</Item>
                <Item key="telegram">Telegram</Item>
                <Item key="whatsapp">WhatsApp</Item>
                <Item key="linkedin">LinkedIn</Item>
              </Picker>
            )}
          />
          <Controller
            name="is_enabled"
            control={form.control}
            render={({ field }) => (
              <Switch isSelected={field.value} onChange={field.onChange}>
                Enabled
              </Switch>
            )}
          />
          {currentType === 'email' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="email"
                control={form.control}
                render={({ field }) => <TextField label="Email" {...field} width="100%" />}
              />
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="Provider" {...field} width="100%" />}
              />
              <Controller
                name="imap_server"
                control={form.control}
                render={({ field }) => <TextField label="IMAP Server" {...field} width="100%" />}
              />
              <Controller
                name="smtp_server"
                control={form.control}
                render={({ field }) => <TextField label="SMTP Server" {...field} width="100%" />}
              />
              <Controller
                name="smtp_tls"
                control={form.control}
                render={({ field }) => (
                  <Switch isSelected={field.value} onChange={field.onChange}>
                    SMTP TLS
                  </Switch>
                )}
              />
              <Controller
                name="password"
                control={form.control}
                render={({ field }) => <TextField label="Password" {...field} width="100%" type="password" />}
              />
            </Flex>
          )}
          {currentType === 'email_oauth' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="email"
                control={form.control}
                render={({ field }) => <TextField label="Email" {...field} width="100%" />}
              />
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="Provider" {...field} width="100%" />}
              />
              <Controller
                name="oauth2_client_uuid"
                control={form.control}
                render={({ field, fieldState }) => (
                  <Picker
                    label="OAuth2 Client"
                    selectedKey={field.value}
                    onSelectionChange={(key) => field.onChange(key.toString())}
                    errorMessage={fieldState.error?.message}
                    width="100%"
                  >
                    {oauth2ClientsQuery.data?.map((client) => 
                      <Item key={client.uuid}>
                        {client.name} ({client.client_id})
                      </Item>
                    )}
                  </Picker>
                )}
              />
            </Flex>
          )}
          {currentType === 'telegram' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="phone_number"
                control={form.control}
                render={({ field }) => <TextField label="Phone Number" {...field} width="100%" />}
              />
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="Provider" {...field} width="100%" />}
              />
              <Controller
                name="api_id"
                control={form.control}
                render={({ field }) => <TextField label="API ID" {...field} width="100%" />}
              />
              <Controller
                name="api_hash"
                control={form.control}
                render={({ field }) => <TextField label="API Hash" {...field} width="100%" />}
              />
              <Controller
                name="password"
                control={form.control}
                render={({ field }) => <TextField label="Password" {...field} width="100%" type="password" />}
              />
            </Flex>
          )}
          {currentType === 'whatsapp' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="phone_number"
                control={form.control}
                render={({ field }) => <TextField label="Phone Number" {...field} width="100%" />}
              />
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="Provider" {...field} width="100%" />}
              />
              <Controller
                name="device_name"
                control={form.control}
                render={({ field }) => <TextField label="Device Name" {...field} width="100%" />}
              />
            </Flex>
          )}
          {currentType === 'linkedin' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="username"
                control={form.control}
                render={({ field }) => <TextField label="Username" {...field} width="100%" />}
              />
              <Controller
                name="password"
                control={form.control}
                render={({ field }) => <TextField label="Password" {...field} width="100%" type="password" />}
              />
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="Provider" {...field} width="100%" />}
              />
            </Flex>
          )}
          <Flex direction="row" gap="size-100" marginTop="size-300" justifyContent="center">
            <Button type="submit" variant="cta">
              {datasourceUUID === 'add' ? 'Create' : 'Update'}
            </Button>
            {datasourceUUID !== 'add' && (
              <Button variant="negative" onPress={onDelete}>
                Delete
              </Button>
            )}
          </Flex>
        </Flex>
      </Form>
    </Flex>
  )
}
