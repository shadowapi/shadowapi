import { ReactElement, useEffect } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Item, Picker, Switch, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components, paths } from '@/api/v1'

export type DataSourceKind = 'email' | 'telegram' | 'whatsapp' | 'linkedin'

type DataSourceBase = {
  name: string
  type: string
  is_enabled: boolean
}

type DataSourceFormData =
  | (DataSourceBase & { type: 'email' } & components['schemas']['datasource_email'])
  | (DataSourceBase & { type: 'telegram' } & components['schemas']['datasource_telegram'])
  | (DataSourceBase & { type: 'whatsapp' } & components['schemas']['datasource_whatsapp'])
  | (DataSourceBase & { type: 'linkedin' } & components['schemas']['datasource_linkedin'])

const createEndpoints: Record<DataSourceKind, keyof paths> = {
  email: '/datasource/email',
  telegram: '/datasource/telegram',
  whatsapp: '/datasource/whatsapp',
  linkedin: '/datasource/linkedin',
}

const updateEndpoints: Record<DataSourceKind, keyof paths> = {
  email: '/datasource/email/{uuid}',
  telegram: '/datasource/telegram/{uuid}',
  whatsapp: '/datasource/whatsapp/{uuid}',
  linkedin: '/datasource/linkedin/{uuid}',
}

const deleteEndpoints: Record<DataSourceKind, keyof paths> = {
  email: '/datasource/email/{uuid}',
  telegram: '/datasource/telegram/{uuid}',
  whatsapp: '/datasource/whatsapp/{uuid}',
  linkedin: '/datasource/linkedin/{uuid}',
}

export function DataSourceForm({
  datasourceUUID,
  datasourceKind,
}: {
  datasourceUUID: string
  datasourceKind?: DataSourceKind
}): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<DataSourceFormData>({})

  if (datasourceUUID !== 'add' && !datasourceKind) {
    throw new Error('datasourceKind is required for editing datasource')
  }

  const getEndpoint = (kind: DataSourceKind, isCreate: boolean): keyof paths => {
    return isCreate ? createEndpoints[kind] : updateEndpoints[kind]
  }

  const query = useQuery({
    queryKey:
      datasourceUUID === 'add'
        ? ['/datasource', 'add']
        : [getEndpoint(datasourceKind!, false), { uuid: datasourceUUID }],
    queryFn: async ({ signal }) => {
      if (datasourceUUID === 'add') return {}
      const endpoint = getEndpoint(datasourceKind!, false)
      const { data } = await client.GET(endpoint, {
        params: { path: { uuid: datasourceUUID } },
        signal,
      })
      return data
    },
    enabled: datasourceUUID !== 'add',
  })

  useEffect(() => {
    if (query.data) {
      form.reset({
        ...query.data,
        type: datasourceKind ?? (query.data as Partial<DataSourceFormData>).type,
      } as DataSourceFormData)
    } else if (datasourceKind) {
      form.setValue('type', datasourceKind)
    }
  }, [query.data, datasourceKind, form])

  const dataSourceType = useWatch({ control: form.control, name: 'type' })

  const modifyMutation = useMutation({
    mutationFn: async (data: DataSourceFormData) => {
      if (datasourceUUID === 'add') {
        const currentKind = data.type as DataSourceKind
        const endpoint = createEndpoints[currentKind]
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
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              username: data.username,
              password: data.password,
              provider: data.provider,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        }
      } else {
        const endpoint = getEndpoint(datasourceKind!, false)
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
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: datasourceUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              username: data.username,
              password: data.password,
              provider: data.provider,
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
      const endpoint = deleteEndpoints[datasourceKind!]
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

  if (query.isPending && datasourceUUID !== 'add') {
    return <></>
  }

  console.log('@reactima remove DataSourceForm', {
    datasourceUUID,
    datasourceKind,
    query,
    form,
    dataSourceType,
    queryData: query.data,
  })

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
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
            name="type"
            control={form.control}
            rules={{ required: 'Type is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Type"
                isRequired
                isDisabled={!!datasourceKind || datasourceUUID !== 'add'}
                selectedKey={datasourceKind ?? field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                <Item key="email">Email</Item>
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
          {dataSourceType === 'email' && (
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
                render={({ field }) => 
                  <Switch isSelected={field.value} onChange={field.onChange}>
                    SMTP TLS
                  </Switch>
                }
              />
              <Controller
                name="password"
                control={form.control}
                render={({ field }) => <TextField label="Password" {...field} width="100%" type="password" />}
              />
            </Flex>
          )}
          {dataSourceType === 'telegram' && (
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
          {dataSourceType === 'whatsapp' && (
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
          {dataSourceType === 'linkedin' && (
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
