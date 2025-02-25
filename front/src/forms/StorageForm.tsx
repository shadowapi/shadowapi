import { ReactElement, useEffect } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Item, Picker, Switch, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'

export function StorageForm({ storageUUID }: { storageUUID: string }): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // This form is based on the base storage schema
  const form = useForm<components['schemas']['storage']>({})

  // Fetch existing storage details if not adding
  const query = useQuery({
    queryKey: ['/storage/{uuid}', { uuid: storageUUID }],
    queryFn: async ({ signal }) => {
      if (storageUUID === 'add') return {}
      const { data } = await client.GET(`/storage/{uuid}`, {
        params: { path: { uuid: storageUUID } },
        signal,
      })
      return data
    },
    enabled: storageUUID !== 'add',
  })

  // Depending on the "type", we can show different sub-fields below
  const storageType = useWatch({ control: form.control, name: 'type' })

  // Either create or update storage
  const modifyMutation = useMutation({
    mutationFn: async (data: components['schemas']['storage']) => {
      if (storageUUID === 'add') {
        const resp = await client.POST('/storage', {
          body: {
            name: data.name,
            type: data.type,
            is_enabled: data.is_enabled,
            // Additional sub-fields can be included here if your backend expects them
            // for specific storage types (S3, file_system, database).
          },
        })
        if (resp.error) {
          form.setError('name', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      } else {
        const resp = await client.PUT(`/storage/{uuid}`, {
          params: { path: { uuid: storageUUID } },
          body: {
            name: data.name,
            type: data.type,
            is_enabled: data.is_enabled,
            // Additional sub-fields for updates
          },
        })
        if (resp.error) {
          form.setError('name', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/storage' })
      navigate('/storages')
    },
  })

  // Delete the storage
  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE(`/storage/{uuid}`, {
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

  const onSubmit = (data: components['schemas']['storage']) => {
    modifyMutation.mutate(data)
  }

  if (query.isPending && storageUUID !== 'add') {
    return <></>
  }

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
                <Item key="file_system">File System</Item>
                <Item key="postgresql">PostgreSQL</Item>
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

          {/* Optional sub-forms based on type */}
          {storageType === 's3' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="provider"
                control={form.control}
                render={({ field }) => <TextField label="S3 Provider" {...field} width="100%" type="text" />}
              />
              <Controller
                name="region"
                control={form.control}
                render={({ field }) => <TextField label="Region" {...field} width="100%" type="text" />}
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

          {storageType === 'file_system' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="path"
                control={form.control}
                render={({ field }) => <TextField label="File System Path" {...field} width="100%" type="text" />}
              />
            </Flex>
          )}

          {storageType === 'postgresql' && (
            <Flex direction="column" gap="size-100" marginTop="size-200">
              <Controller
                name="host"
                control={form.control}
                render={({ field }) => <TextField label="PostgreSQL Host" {...field} width="100%" type="text" />}
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
