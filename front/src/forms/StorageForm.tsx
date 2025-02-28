import { ReactElement, useEffect } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Item, Picker, Switch, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components, paths } from '@/api/v1'

export type StorageKind = 's3' | 'hostfiles' | 'postgres'

type StorageBase = {
  name: string
  type: string
  is_enabled: boolean
}

type StorageFormData =
  | (StorageBase & { type: 's3' } & components['schemas']['storage_s3'])
  | (StorageBase & { type: 'hostfiles' } & components['schemas']['storage_hostfiles'])
  | (StorageBase & { type: 'postgres' } & components['schemas']['storage_postgres'])

const createEndpoints: Record<StorageKind, keyof paths> = {
  s3: '/storage/s3',
  hostfiles: '/storage/hostfiles',
  postgres: '/storage/postgres',
}

const updateEndpoints: Record<StorageKind, keyof paths> = {
  s3: '/storage/s3/{uuid}',
  hostfiles: '/storage/hostfiles/{uuid}',
  postgres: '/storage/postgres/{uuid}',
}

const deleteEndpoints: Record<StorageKind, keyof paths> = {
  s3: '/storage/s3/{uuid}',
  hostfiles: '/storage/hostfiles/{uuid}',
  postgres: '/storage/postgres/{uuid}',
}

export function StorageForm({
  storageUUID,
  storageKind,
}: {
  storageUUID: string
  storageKind?: StorageKind
}): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<StorageFormData>({})

  if (storageUUID !== 'add' && !storageKind) {
    throw new Error('storageKind is required for editing storage')
  }

  const getEndpoint = (kind: StorageKind, isCreate: boolean): keyof paths => {
    return isCreate ? createEndpoints[kind] : updateEndpoints[kind]
  }

  const query = useQuery({
    queryKey: storageUUID === 'add' ? ['/storage', 'add'] : [getEndpoint(storageKind!, false), { uuid: storageUUID }],
    queryFn: async ({ signal }) => {
      if (storageUUID === 'add') return {}
      const endpoint = getEndpoint(storageKind!, false)
      const { data } = await client.GET(endpoint, {
        params: { path: { uuid: storageUUID } },
        signal,
      })
      return data
    },
    enabled: storageUUID !== 'add',
  })

  const storageType = useWatch({ control: form.control, name: 'type' })

  const modifyMutation = useMutation({
    mutationFn: async (data: StorageFormData) => {
      if (storageUUID === 'add') {
        const currentKind = data.type as StorageKind
        const endpoint = createEndpoints[currentKind]
        if (data.type === 's3') {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              provider: data.provider,
              region: data.region,
              bucket: data.bucket,
              access_key_id: data.access_key_id,
              secret_access_key: data.secret_access_key,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'hostfiles') {
          const resp = await client.POST(endpoint, {
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              path: data.path,
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
              host: data.host,
              port: data.port,
              user: data.user,
              password: data.password,
              options: data.options,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        }
      } else {
        const endpoint = getEndpoint(storageKind!, false)
        if (data.type === 's3') {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: storageUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              provider: data.provider,
              region: data.region,
              bucket: data.bucket,
              access_key_id: data.access_key_id,
              secret_access_key: data.secret_access_key,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else if (data.type === 'hostfiles') {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: storageUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              path: data.path,
            },
          })
          if (resp.error) {
            form.setError('name', { message: resp.error.detail })
            throw new Error(resp.error.detail)
          }
          return resp
        } else {
          const resp = await client.PUT(endpoint, {
            params: { path: { uuid: storageUUID } },
            body: {
              name: data.name,
              is_enabled: data.is_enabled,
              host: data.host,
              port: data.port,
              user: data.user,
              password: data.password,
              options: data.options,
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
      queryClient.invalidateQueries({ queryKey: '/storage' })
      navigate('/storages')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const endpoint = deleteEndpoints[storageKind!]
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
      queryClient.invalidateQueries({ queryKey: '/storage' })
      navigate('/storages')
    },
  })

  useEffect(() => {
    if (query.data) {
      form.reset(query.data)
    }
  }, [query.data, form])

  const onDelete = () => {
    deleteMutation.mutate(storageUUID)
  }

  const onSubmit = (data: StorageFormData) => {
    modifyMutation.mutate(data)
  }

  if (query.isPending && storageUUID !== 'add') return <></>

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600">
          <Header marginBottom="size-160">Storage</Header>
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
                isDisabled={storageUUID !== 'add'}
                selectedKey={field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                <Item key="s3">S3</Item>
                <Item key="hostfiles">File System</Item>
                <Item key="postgres">PostgreSQL</Item>
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
          {storageType === 's3' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="S3 Provider" {...field} width="100%" />}
              />
              <Controller
                name="region"
                control={form.control}
                render={({ field }) => <TextField label="Region" {...field} width="100%" />}
              />
              <Controller
                name="bucket"
                control={form.control}
                render={({ field }) => <TextField label="Bucket Name" {...field} width="100%" type="text" />}
              />
              <Controller
                name="access_key_id"
                control={form.control}
                render={({ field }) => <TextField label="Access Key ID" {...field} width="100%" type="text" />}
              />
              <Controller
                name="secret_access_key"
                control={form.control}
                render={({ field }) => <TextField label="Secret Access Key" {...field} width="100%" type="password" />}
              />
            </Flex>
          )}
          {storageType === 'hostfiles' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="path"
                control={form.control}
                render={({ field }) => <TextField label="File System Path" {...field} width="100%" />}
              />
            </Flex>
          )}
          {storageType === 'postgres' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="host"
                control={form.control}
                render={({ field }) => <TextField label="PostgreSQL Host" {...field} width="100%" />}
              />
              <Controller
                name="port"
                control={form.control}
                render={({ field }) => <TextField label="Port" {...field} width="100%" type="text" />}
              />
              <Controller
                name="user"
                control={form.control}
                render={({ field }) => <TextField label="User" {...field} width="100%" type="text" />}
              />
              <Controller
                name="password"
                control={form.control}
                render={({ field }) => <TextField label="Password" {...field} width="100%" type="password" />}
              />
              <Controller
                name="name"
                control={form.control}
                render={({ field }) => <TextField label="DB Name" {...field} width="100%" type="text" />}
              />
              <Controller
                name="options"
                control={form.control}
                render={({ field }) => <TextField label="Options" {...field} width="100%" type="text" />}
              />
            </Flex>
          )}
          <Flex direction="row" gap="size-100" marginTop="size-300" justifyContent="center">
            <Button type="submit" variant="cta">
              {storageUUID === 'add' ? 'Create' : 'Update'}
            </Button>
            {storageUUID !== 'add' && (
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
