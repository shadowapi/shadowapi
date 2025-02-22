import { ReactElement, useEffect, useState } from 'react'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Divider, Flex, Form, Header, Item, Picker, Switch, TextField, View } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { DataSourceGrant as DataSourceGrantForm } from './DataSourceGrant'

import client from '@/api/client'
import type { components } from '@/api/v1'

export function Storage({ datasourceUUID: storageUUID }: { datasourceUUID: string }): ReactElement {
  const navigate = useNavigate()
  const form = useForm<components['schemas']['storage']>({})

  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: ['/storage/postgres/{uuid}', { uuid: storageUUID }],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET(`/storage/postgres/{uuid}`, {
        params: { path: { uuid: storageUUID } },
        signal,
      })
      return data
    },
    enabled: storageUUID !== 'add',
  })
  const modifyMutation = useMutation({
    mutationFn: async (data: components['schemas']['storage']) => {
      let resp
      if (storageUUID === 'add') {
        resp = await client.POST('/storage/postgres', {
          body: {
            // is_enabled: data.is_enabled || true,
            // name: data.name || '',
            // email: data.email || '',
            // imap_server: data.imap_server,
            // password: data.password,
            // provider: data.provider,
            // smtp_server: data.smtp_server,
            // smtp_tls: data.smtp_tls,
          },
        })
      } else {
        resp = await client.PUT(`/storage/postgres/{uuid}`, {
          params: { path: { uuid: storageUUID } },
          body: {
            // is_enabled: data.is_enabled,
            // imap_server: data.imap_server,
            // name: data.name,
            // password: data.password,
            // smtp_server: data.smtp_server,
            // smtp_tls: data.smtp_tls,
          },
        })
      }
      if (resp.error) {
        form.setError('name', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
    },
    onSuccess: (data, variable) => {
      if (storageUUID === 'add') {
        queryClient.invalidateQueries({ queryKey: '/datasource/email' })
      } else {
        queryClient.setQueryData(['/datasource/email/{uuid}', { uuid: variable.uuid }], data)
      }
    },
  })
  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE(`/datasource/email/{uuid}`, {
        params: { path: { uuid: uuid } },
      })
      if (resp.error) {
        form.setError('name', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/datasource/email' })
    },
  })

  const datasourceType = useWatch({ control: form.control, name: 'type', defaultValue: '' })
  const datasourceEmail = useWatch({ control: form.control, name: 'email', defaultValue: '' })

  const [showOAuth2Token, setOAuth2Token] = useState(false)
  const [showEmailFields, setShowEmailFields] = useState(false)

  useEffect(() => {
    if (query.data) {
      form.reset(query.data)
    }
  }, [query.data, form])

  useEffect(() => {
    if (datasourceType === 'email') {
      if (datasourceEmail && datasourceEmail.includes('@gmail.com')) {
        form.setValue('provider', 'gmail')
        setShowEmailFields(false)
        setOAuth2Token(true)
      } else {
        setShowEmailFields(true)
        setOAuth2Token(false)
      }
    } else {
      setShowEmailFields(false)
    }
  }, [datasourceType, datasourceEmail, form])

  const onDelete = () => {
    deleteMutation.mutate(storageUUID, {
      onSuccess: () => {
        navigate('/datasources')
      },
    })
  }

  const onSubmit = async (data: components['schemas']['datasource']) => {
    modifyMutation.mutate(data, {
      onSuccess: () => {
        navigate('/datasources')
      },
    })
  }

  if (query.isPending && storageUUID !== 'add') {
    return <></>
  }

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
                type="text"
                width="100%"
                isRequired
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
            render={({ field, fieldState }): ReactElement => (
              <Picker
                label="Type"
                isRequired
                isDisabled={storageUUID !== 'add'}
                selectedKey={field.value}
                onSelectionChange={(key) => form.setValue('type', key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                <Item key="email">Email</Item>
                <Item key="other">Other</Item>
              </Picker>
            )}
          />

          {datasourceType === 'email' && (
            <Controller
              name="email"
              control={form.control}
              rules={{ required: 'Email is required' }}
              render={({ field, fieldState }) => (
                <TextField
                  label="Email"
                  type="text"
                  isRequired
                  width="100%"
                  isDisabled={storageUUID !== 'add'}
                  validationState={fieldState.invalid ? 'invalid' : undefined}
                  errorMessage={fieldState.error?.message}
                  // TODO remove ...spread?
                  // value={field.value || ''} // Ensure value is always a string
                  // onChange={(value) => field.onChange(value)} // Explicitly handle change event
                  // onBlur={field.onBlur}
                  {...field}
                />
              )}
            />
          )}

          {showEmailFields && (
            <>
              <Controller
                name="imap_server"
                control={form.control}
                render={({ field, fieldState }) => (
                  <TextField
                    label="IMAP Server"
                    type="text"
                    width="100%"
                    isRequired
                    validationState={fieldState.invalid ? 'invalid' : undefined}
                    errorMessage={fieldState.error?.message}
                    {...field}
                  />
                )}
              />
              <Controller
                name="password"
                control={form.control}
                render={({ field, fieldState }) => (
                  <TextField
                    label="Password"
                    type="password"
                    width="100%"
                    isRequired
                    validationState={fieldState.invalid ? 'invalid' : undefined}
                    errorMessage={fieldState.error?.message}
                    {...field}
                  />
                )}
              />
              <Controller
                name="smtp_server"
                control={form.control}
                render={({ field, fieldState }) => (
                  <TextField
                    label="SMTP Server"
                    type="text"
                    width="100%"
                    isRequired
                    validationState={fieldState.invalid ? 'invalid' : undefined}
                    errorMessage={fieldState.error?.message}
                    {...field}
                  />
                )}
              />
              <Controller
                name="smtp_tls"
                control={form.control}
                render={({ field }) => (
                  <Switch width="100%" name={field.name} isSelected={field.value} onChange={field.onChange}>
                    Use TLS
                  </Switch>
                )}
              />
            </>
          )}
          <Flex direction="row" gap="size-100" marginTop="size-300" justifyContent="center">
            <Button type="submit" variant="cta">
              {storageUUID === 'add' ? 'Create' : 'Update'}
            </Button>
            {storageUUID !== 'add' && (
              <Button type="button" variant="negative" onPress={onDelete}>
                Delete
              </Button>
            )}
          </Flex>
          {showOAuth2Token && storageUUID !== 'add' && (
            <View marginTop="size-300">
              <Divider marginBottom="size-300" size="S" />
              <DataSourceGrantForm datasourceUUID={storageUUID} tokenUUID={query.data?.oauth2_token_uuid} />
            </View>
          )}
        </Flex>
      </Form>
    </Flex>
  )
}
